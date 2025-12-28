package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/google"
	"travia.backend/api/helpers"
	"travia.backend/api/models"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

func (s *Server) SetupAuthProviders() {
	// Sử dụng secret key từ config
	key := s.config.ServerConfig.SecretKey
	maxAge := 86400 * 30 // 30 days
	isProd := s.config.ServerConfig.Environment == "production"

	store := cookie.NewStore([]byte(key))
	options := sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   maxAge,
		Secure:   isProd,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	store.Options(options)

	gothic.Store = store
	goth.UseProviders(
		google.New(
			s.config.GoogleCloudConfig.GoogleClientId,
			s.config.GoogleCloudConfig.GoogleClientSecret,
			s.config.GoogleCloudConfig.GoogleRedirectUris,
			"email",
			"profile",
		),
		facebook.New("", "", "", ""),
	)
}

func (s *Server) AuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		fmt.Printf("AuthHandler called with provider: %s\n", provider)

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), gothic.ProviderParamKey, provider))

		if provider != "google" {
			helpers.BadRequest(c, "Unsupported OAuth provider", nil)
			return
		}

		fmt.Println("Starting Google OAuth...")
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
}

func (s *Server) AuthCallbackHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), gothic.ProviderParamKey, provider))

		// Check if provider is supported
		if provider != "google" {
			helpers.BadRequest(c, "Unsupported OAuth provider", nil)
			return
		}
		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			helpers.BadRequest(c, "Failed to complete OAuth authentication", err)
			return
		}
		var id pgtype.UUID
		err = id.Scan(user.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "xx",
			})
		}

		tokenPair, err := utils.GenerateToken(id, user.Email, "user", s.config.ServerConfig.SecretKey)
		if err != nil {
			helpers.InternalServerError(c, "Failed to generate token", err)
			return
		}

		response := map[string]interface{}{
			"message": "Google OAuth login successful",
			"user": map[string]interface{}{
				"id":       user.UserID,
				"email":    user.Email,
				"name":     user.Name,
				"picture":  user.AvatarURL,
				"provider": user.Provider,
			},
			"token": tokenPair.AccessToken,
		}

		helpers.Ok(c, response)
	}
}

// tạo tài khoản từ form
// @summary Tạo tài khoản từ form
// @description Tạo tài khoản từ form
// @tags auth
// @accept json
// @produce json
// @param req body models.CreateUser true "Thông tin tài khoản"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 409 {object} gin.H "Email đã được đăng ký"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/createUserForm [post]
func (s *Server) CreateUserForm(c *gin.Context) {
	var req models.CreateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu đầu vào không hợp lệ",
			"message": err.Error(),
		})
		return
	}

	// Validate input fields
	if req.FullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Họ tên không được để trống",
		})
		return
	}

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email không được để trống",
		})
		return
	}

	if !helpers.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email không hợp lệ",
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Mật khẩu không được để trống",
		})
		return
	}
	_, err := s.z.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		// Kiểm tra nếu lỗi là "no rows found" - có nghĩa là email chưa được đăng ký
		if errors.Is(err, pgx.ErrNoRows) {
			// Email chưa được đăng ký, tiếp tục xử lý
		} else {
			// Lỗi khác từ database
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Lỗi khi kiểm tra email",
			})
			return
		}
	} else {
		// Không có lỗi, có nghĩa là tìm thấy user với email này
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email đã được đăng ký",
		})
		return
	}
	if !helpers.ValidatePassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Mật khẩu không hợp lệ. Mật khẩu phải có ít nhất 8 ký tự, bao gồm chữ hoa, chữ thường và số",
		})
		return
	}

	// Generate verification code
	verificationCode := helpers.GenerateVerificationCode()

	// Store pending user data in Redis FIRST (before sending email)
	pendingUser := models.PendingUser{
		FullName: req.FullName,
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
		Otp:      verificationCode,
	}

	data, err := json.Marshal(pendingUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Lỗi khi xử lý dữ liệu người dùng",
			"message": err.Error(),
		})
		return
	}

	// Store in Redis with 5 minutes expiration
	if err := s.redis.SetEx(context.Background(), "register:"+pendingUser.Email, data, 5*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể lưu trữ dữ liệu tạm thời",
		})
		return
	}

	// 🚀 Send verification email ASYNCHRONOUSLY (non-blocking)
	// This won't block the response even if email fails
	go helpers.SendVerificationEmailAsync(req.Email, verificationCode, s.config.EmailConfig)

	c.JSON(http.StatusOK, gin.H{
		"message": "Mã xác thực đã được gửi đến email của bạn",
		"email":   pendingUser.Email,
	})
}

// tạo tài khoản từ OTP
// @summary Tạo tài khoản từ OTP
// @description Tạo tài khoản từ OTP
// @tags auth
// @accept json
// @produce json
// @param req body models.OTP true "Thông tin OTP"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "OTP không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/createUser [post]
func (s *Server) CreateUser(c *gin.Context) {
	var req models.OTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu đầu vào không hợp lệ",
			"message": err.Error(),
		})
		return
	}

	// Validate input
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email không được để trống",
		})
		return
	}

	if req.Otp == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Mã OTP không được để trống",
		})
		return
	}

	// Get pending user data from Redis
	stored, err := s.redis.Get(context.Background(), "register:"+req.Email).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Mã OTP đã hết hạn hoặc không tồn tại. Vui lòng đăng ký lại",
		})
		return
	}

	var pendingUser models.PendingUser
	if err := json.Unmarshal([]byte(stored), &pendingUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Lỗi khi xử lý dữ liệu người dùng",
		})
		return
	}

	// Verify OTP
	if req.Otp != pendingUser.Otp {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Mã OTP không chính xác",
		})
		return
	}
	// Hash password
	hashedPassword, err := utils.HashPassword(pendingUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể mã hóa mật khẩu",
		})
		return
	}

	// Create user in database
	user, err := s.z.CreateUser(context.Background(), db.CreateUserParams{
		HoTen:        pendingUser.FullName,
		Email:        pendingUser.Email,
		MatKhauMaHoa: hashedPassword,
		SoDienThoai:  pendingUser.Phone,
		VaiTro:       db.NullVaiTroNguoiDung{VaiTroNguoiDung: "khach_hang", Valid: true},
		DangHoatDong: helpers.NewBool(true),
		XacThuc:      helpers.NewBool(true),
		NgayTao:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		NgayCapNhat:  pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể tạo tài khoản người dùng",
		})
		return
	}

	// Clean up Redis data after successful registration
	s.redis.Del(context.Background(), "register:"+pendingUser.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng ký tài khoản thành công",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.HoTen,
		},
	})
}

// loginWithRole là hàm helper để xử lý đăng nhập với kiểm tra vai trò
func (s *Server) loginWithRole(c *gin.Context, requiredRole db.VaiTroNguoiDung) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu đầu vào không hợp lệ",
			"message": err.Error(),
		})
		return
	}

	// Validate input
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email không được để trống",
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Mật khẩu không được để trống",
		})
		return
	}

	// Get user by email
	user, err := s.z.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		// Kiểm tra nếu lỗi là "no rows found" - có nghĩa là tài khoản không tồn tại
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Email hoặc mật khẩu không chính xác",
			})
			return
		} else {
			// Lỗi khác từ database
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Lỗi hệ thống",
			})
			return
		}
	}

	// Check if user is active
	if user.DangHoatDong != nil && !*user.DangHoatDong {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tài khoản đã bị khóa",
		})
		return
	}

	// Check user role matches the required role
	if !user.VaiTro.Valid || user.VaiTro.VaiTroNguoiDung != requiredRole {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Bạn không có quyền đăng nhập vào hệ thống này",
		})
		return
	}

	// Verify password
	if !utils.CheckHashPassword(req.Password, user.MatKhauMaHoa) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Email hoặc mật khẩu không chính xác",
		})
		return
	}

	// Generate JWT tokens
	tokenPair, err := utils.GenerateToken(user.ID, user.Email, string(user.VaiTro.VaiTroNguoiDung), s.config.ServerConfig.ApiSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể tạo token xác thực",
		})
		return
	}

	// Return tokens in response body (no cookies)
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập thành công",
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"name":     user.HoTen,
			"role":     user.VaiTro.VaiTroNguoiDung,
			"ngay_tao": user.NgayTao.Time.Format(time.DateTime),
		},
		"tokens": gin.H{
			"accessToken":  tokenPair.AccessToken,
			"refreshToken": tokenPair.RefreshToken,
		},
	})
}

// đăng nhập (deprecated - sử dụng endpoint cụ thể theo vai trò)
// @summary Đăng nhập (deprecated)
// @description Đăng nhập chung (deprecated - sử dụng /login/user, /login/admin hoặc /login/supplier)
// @tags auth
// @accept json
// @produce json
// @param req body models.LoginRequest true "Thông tin đăng nhập"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/login [post]
// @deprecated
func (s *Server) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu đầu vào không hợp lệ",
			"message": err.Error(),
		})
		return
	}

	// Validate input
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email không được để trống",
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Mật khẩu không được để trống",
		})
		return
	}

	// Get user by email
	user, err := s.z.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		// Kiểm tra nếu lỗi là "no rows found" - có nghĩa là tài khoản không tồn tại
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Email hoặc mật khẩu không chính xác",
			})
			return
		} else {
			// Lỗi khác từ database
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Lỗi hệ thống",
			})
			return
		}
	}

	// Check if user is active
	if user.DangHoatDong != nil && !*user.DangHoatDong {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tài khoản đã bị khóa",
		})
		return
	}

	// Verify password
	if !utils.CheckHashPassword(req.Password, user.MatKhauMaHoa) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Email hoặc mật khẩu không chính xác",
		})
		return
	}
	// Generate JWT tokens
	tokenPair, err := utils.GenerateToken(user.ID, user.Email, string(user.VaiTro.VaiTroNguoiDung), s.config.ServerConfig.ApiSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể tạo token xác thực",
		})
		return
	}

	// Return tokens in response body (no cookies)
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập thành công",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.HoTen,
			"role":  user.VaiTro.VaiTroNguoiDung,
		},
		"tokens": gin.H{
			"accessToken":  tokenPair.AccessToken,
			"refreshToken": tokenPair.RefreshToken,
		},
	})
}

// đăng nhập cho người dùng/khách hàng
// @summary Đăng nhập cho người dùng/khách hàng
// @description Đăng nhập dành cho người dùng với vai trò khách hàng
// @tags auth
// @accept json
// @produce json
// @param req body models.LoginRequest true "Thông tin đăng nhập"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 403 {object} gin.H "Không có quyền truy cập"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/login/user [post]
func (s *Server) LoginUser(c *gin.Context) {
	s.loginWithRole(c, db.VaiTroNguoiDungKhachHang)
}

// đăng nhập cho admin
// @summary Đăng nhập cho admin
// @description Đăng nhập dành cho quản trị viên
// @tags auth
// @accept json
// @produce json
// @param req body models.LoginRequest true "Thông tin đăng nhập"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 403 {object} gin.H "Không có quyền truy cập"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/login/admin [post]
func (s *Server) LoginAdmin(c *gin.Context) {
	s.loginWithRole(c, db.VaiTroNguoiDungQuanTri)
}

// đăng nhập cho nhà cung cấp
// @summary Đăng nhập cho nhà cung cấp
// @description Đăng nhập dành cho nhà cung cấp
// @tags auth
// @accept json
// @produce json
// @param req body models.LoginRequest true "Thông tin đăng nhập"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 403 {object} gin.H "Không có quyền truy cập"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/login/supplier [post]
func (s *Server) LoginSupplier(c *gin.Context) {
	s.loginWithRole(c, db.VaiTroNguoiDungNhaCungCap)
}

// làm mới token
// @summary Làm mới token
// @description Làm mới access token bằng refresh token
// @tags auth
// @accept json
// @produce json
// @param req body models.RefreshTokenRequest true "Refresh token"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 401 {object} gin.H "Refresh token không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/refresh [post]
func (s *Server) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Refresh token không được để trống",
		})
		return
	}

	// Validate refresh token
	claims, err := utils.ValidateToken(req.RefreshToken, s.config.ServerConfig.ApiSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Refresh token không hợp lệ hoặc đã hết hạn",
		})
		return
	}

	// Generate new token pair
	tokenPair, err := utils.GenerateToken(claims.Id, claims.Email, claims.Vaitro, s.config.ServerConfig.ApiSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể tạo token mới",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Làm mới token thành công",
		"tokens": gin.H{
			"accessToken":  tokenPair.AccessToken,
			"refreshToken": tokenPair.RefreshToken,
		},
	})
}

// đăng xuất
// @summary Đăng xuất
// @description Đăng xuất (tokens được xóa ở frontend)
// @tags auth
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @router /auth/logout [post]
func (s *Server) Logout(c *gin.Context) {
	// Tokens are stored in localStorage on frontend, so no need to clear cookies
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng xuất thành công",
	})
}

// lấy thông tin user
// @summary Lấy thông tin user
// @description Lấy thông tin user
// @tags auth
// @accept json
// @produce json
// @param id path string true "ID của user"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/getUserById/{id} [get]
func (s *Server) GetUserById(c *gin.Context) {
	_id := c.Param("id")
	var id pgtype.UUID
	if err := id.Scan(_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID không hợp lệ",
		})
		return
	}

	_user, err := s.z.GetUserById(context.Background(), id)
	if err != nil {
		// Kiểm tra nếu lỗi là "no rows found" - có nghĩa là user không tồn tại
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Không tìm thấy user",
			})
			return
		} else {
			// Lỗi khác từ database
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	timeNow := time.Now().Format(time.DateTime)
	ngay_cap_nhat, err := time.Parse(time.DateTime, timeNow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể parse ngày",
		})
		return
	}
	user := models.GetUser{
		ID:                       _user.ID.String(),
		FullName:                 _user.HoTen,
		Email:                    _user.Email,
		Phone:                    _user.SoDienThoai,
		TongDatCho:               _user.TongDatCho,
		TongDatChoDaThanhToan:    _user.TongDatChoDaThanhToan,
		TongDatChoDangChoXacNhan: _user.TongDatChoDangChoXacNhan,
		NgayCapNhat:              ngay_cap_nhat.String(),
		NgayTao:                  _user.NgayTao.Time.Format(time.DateTime),
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy thông tin user thành công",
		"data":    user,
	})
}

// cập nhật thông tin user
// @summary Cập nhật thông tin user
// @description Cập nhật thông tin user
// @tags auth
// @accept json
// @produce json
// @param id path string true "ID của user"
// @param req body models.GetUser true "Thông tin cập nhật"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/updateUserById/{id} [put]
func (s *Server) UpdateUserById(c *gin.Context) {
	_id := c.Param("id")
	var id pgtype.UUID
	if err := id.Scan(_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID không hợp lệ",
		})
		return
	}
	var req models.GetUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu đầu vào không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	// Enforce self-or-admin at handler level as defense-in-depth
	if v, exists := c.Get("claims"); exists {
		if claims, ok := v.(*utils.JwtClams); ok {
			// allow admin
			if strings.ToLower(claims.Vaitro) != "quan_tri" {
				if !strings.EqualFold(claims.Id.String(), _id) {
					c.JSON(http.StatusForbidden, gin.H{"error": "Không có quyền cập nhật user này"})
					return
				}
			}
		}
	}

	user, err := s.z.UpdateUserById(context.Background(), db.UpdateUserByIdParams{
		ID:          id,
		HoTen:       req.FullName,
		Email:       req.Email,
		SoDienThoai: req.Phone,
		NgayCapNhat: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể cập nhật thông tin user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật thông tin user thành công",
		"data": gin.H{
			"id":            user.ID,
			"email":         user.Email,
			"name":          user.HoTen,
			"phone":         user.SoDienThoai,
			"ngay_cap_nhat": user.NgayCapNhat.Time.Format(time.DateTime),
		},
	})
}

// đặt lại mật khẩu
// @summary Đặt lại mật khẩu
// @description Đặt lại mật khẩu
// @tags auth
// @accept json
// @produce json
// @param req body db.ResetPasswordParams true "Thông tin đặt lại mật khẩu"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/resetPassword/{email} [put]
func (s *Server) ResetPassword(c *gin.Context) {
	var req db.ResetPasswordParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu đầu vào không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	hashedPassword, err := utils.HashPassword(req.MatKhauMaHoa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể mã hóa mật khẩu",
		})
		return
	}
	user, err := s.z.ResetPassword(context.Background(), db.ResetPasswordParams{
		Email:        req.Email,
		MatKhauMaHoa: hashedPassword,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể đặt lại mật khẩu",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Đặt lại mật khẩu thành công",
		"data":    user,
	})
}

// cập nhật thông tin user
// @summary Cập nhật thông tin user
// @description Cập nhật thông tin user
// @tags auth
// @accept json
// @produce json
// @param req body models.UpdateUser true "Thông tin cập nhật"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/updateUser [put]
func (s *Server) UpdateUser(c *gin.Context) {
	var req models.UpdateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu đầu vào không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Không có quyền truy cập",
		})
		return
	}
	id := claims.(*utils.JwtClams).Id
	user, err := s.z.UpdateUser(context.Background(), db.UpdateUserParams{
		ID:          id,
		HoTen:       &req.FullName,
		Email:       &req.Email,
		SoDienThoai: req.Phone,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể cập nhật thông tin user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật thông tin user thành công",
		"data":    user,
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"mat_khau_cu"`
	NewPassword string `json:"mat_khau_moi"`
}

// đổi mật khẩu
// @summary Đổi mật khẩu
// @description Đổi mật khẩu
// @tags auth
// @accept json
// @produce json
// @param req body ChangePasswordRequest true "Thông tin đổi mật khẩu"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/changePassword [put]
func (s *Server) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu đầu vào không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Không có quyền truy cập",
		})
		return
	}

	user, err := s.z.GetUserById(context.Background(), claims.(*utils.JwtClams).Id)
	fmt.Println(claims.(*utils.JwtClams).Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy thông tin user",
			"message": err.Error(),
		})
		return
	}
	if !utils.CheckHashPassword(req.OldPassword, user.MatKhauMaHoa) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Mật khẩu cũ không chính xác",
		})
		return
	}
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể mã hóa mật khẩu",
		})
		return
	}
	err = s.z.ChangePassword(context.Background(), db.ChangePasswordParams{
		ID:           claims.(*utils.JwtClams).Id,
		MatKhauMaHoa: hashedPassword,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể đổi mật khẩu",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Đổi mật khẩu thành công",
	})
}
