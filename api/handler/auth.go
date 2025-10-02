package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

func newbool(x bool) *bool {
	return &x
}
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

	// Generate verification code and send email
	verificationCode := helpers.GenerateVerificationCode()
	if err := helpers.SendVerificationEmail(req.Email, verificationCode, s.config.EmailConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể gửi email xác thực",
		})
		return
	}

	// Store pending user data in Redis
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
			"error": "Lỗi khi xử lý dữ liệu người dùng",
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
		DangHoatDong: newbool(true),
		XacThuc:      newbool(true),
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

// đăng nhập
// @summary Đăng nhập
// @description Đăng nhập
// @tags auth
// @accept json
// @produce json
// @param req body models.LoginRequest true "Thông tin đăng nhập"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/login [post]
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

	// Set secure cookies
	c.SetCookie("accessToken", tokenPair.AccessToken, 3600*24*7, "/", "", false, true)    // 7 days
	c.SetCookie("refreshToken", tokenPair.RefreshToken, 3600*24*30, "/", "", false, true) // 30 days

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

// đăng xuất
// @summary Đăng xuất
// @description Đăng xuất
// @tags auth
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /auth/logout [post]
func (s *Server) Logout(c *gin.Context) {
	// Clear authentication cookies
	c.SetCookie("accessToken", "", -1, "/", "", false, true)
	c.SetCookie("refreshToken", "", -1, "/", "", false, true)

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
	user := models.GetUpdateUser{
		ID:          _user.ID.String(),
		FullName:    _user.HoTen,
		Email:       _user.Email,
		Phone:       _user.SoDienThoai,
		NgayCapNhat: _user.NgayCapNhat.Time.Format(time.DateTime),
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
// @param req body models.GetUpdateUser true "Thông tin cập nhật"
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
	var req models.GetUpdateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu đầu vào không hợp lệ",
			"message": err.Error(),
		})
		return
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
