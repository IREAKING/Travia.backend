package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"travia.backend/api/helpers"
	"travia.backend/api/models"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

// Đăng ký đối tác
// @Summary Đăng ký đối tác
// @Description Đăng ký đối tác và chờ admin duyệt
// @Tags Supplier
// @Accept json
// @Produce json
// @Param supplier body models.CreateSupplierRequest true "Thông tin đối tác"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/createSupplier [post]
func (s *Server) CreateSupplier(c *gin.Context) {
	var req models.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	hashedPassword, err := utils.HashPassword(req.ThongTinDangNhap.MatKhau)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result, err := s.z.CreateSupplierWithUser(context.Background(), db.CreateSupplierWithUserParams{
		CreateUserParams: db.CreateUserParams{
			HoTen:        req.ThongTinDangNhap.HoTen,
			Email:        req.ThongTinDangNhap.Email,
			MatKhauMaHoa: hashedPassword,
			SoDienThoai:  req.ThongTinDangNhap.SoDienThoai,
			VaiTro:       db.NullVaiTroNguoiDung{VaiTroNguoiDung: db.VaiTroNguoiDung(db.VaiTroNguoiDungNhaCungCap), Valid: true},
			DangHoatDong: helpers.NewBool(true),
			XacThuc:      helpers.NewBool(false),
			NgayTao:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			NgayCapNhat:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
		CreateSupplierParams: db.CreateSupplierParams{
			Ten:     req.ThongTinNhaCungCap.Ten,
			DiaChi:  req.ThongTinNhaCungCap.DiaChi,
			Website: req.ThongTinNhaCungCap.Website,
			MoTa:    req.ThongTinNhaCungCap.MoTa,
			Logo:    req.ThongTinNhaCungCap.LogoUrl,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng ký đối tác thành công. Tài khoản của bạn đang chờ admin duyệt.",
		"data":    result,
	})
}

// Đăng ký đối tác (công khai, không cần auth)
// @Summary Đăng ký đối tác
// @Description Đăng ký đối tác và chờ admin duyệt (endpoint công khai). Upload logo (ảnh) và giấy phép kinh doanh (PDF)
// @Tags Supplier
// @Accept multipart/form-data
// @Produce json
// @Param nguoi_dai_dien formData string true "Họ tên người đại diện"
// @Param email formData string true "Email"
// @Param mat_khau formData string true "Mật khẩu"
// @Param so_dien_thoai formData string false "Số điện thoại"
// @Param ten formData string true "Tên nhà cung cấp"
// @Param dia_chi formData string false "Địa chỉ"
// @Param website formData string false "Website"
// @Param mo_ta formData string false "Mô tả"
// @Param logo formData file false "Logo (ảnh)"
// @Param nam_thanh_lap formData string true "Năm thành lập (YYYY-MM-DD)"
// @Param thanh_pho formData string false "Thành phố"
// @Param quoc_gia formData string false "Quốc gia"
// @Param ma_so_thue formData string false "Mã số thuế"
// @Param so_nhan_vien formData string false "Số nhân viên"
// @Param giay_to_kinh_doanh formData file false "Giấy phép kinh doanh (PDF)"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/register [post]
func (s *Server) RegisterPartner(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse form data
	hoTen := c.PostForm("nguoi_dai_dien")
	email := c.PostForm("email")
	matKhau := c.PostForm("mat_khau")
	soDienThoai := c.PostForm("so_dien_thoai")
	ten := c.PostForm("ten")
	diaChi := c.PostForm("dia_chi")
	website := c.PostForm("website")
	moTa := c.PostForm("mo_ta")
	namThanhLapStr := c.PostForm("nam_thanh_lap")
	thanhPho := c.PostForm("thanh_pho")
	quocGia := c.PostForm("quoc_gia")
	maSoThue := c.PostForm("ma_so_thue")
	soNhanVien := c.PostForm("so_nhan_vien")

	// Validate required fields
	if hoTen == "" || email == "" || matKhau == "" || ten == "" || namThanhLapStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Các trường bắt buộc: nguoi_dai_dien, email, mat_khau, ten, nam_thanh_lap"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(matKhau)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Parse nam thanh lap
	namThanhLap, err := time.Parse(time.DateOnly, namThanhLapStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Định dạng năm thành lập không hợp lệ. Sử dụng YYYY-MM-DD"})
		return
	}

	// Handle logo upload (image)
	var logoURL *string
	logoFile, err := c.FormFile("logo")
	if err == nil && logoFile != nil {
		// Validate image type
		allowedImageTypes := map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
			"image/bmp":  true,
		}

		contentType := logoFile.Header.Get("Content-Type")
		if !allowedImageTypes[contentType] {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Logo phải là file ảnh (JPEG, PNG, GIF, WebP, BMP)"})
			return
		}

		// Open and read logo file
		logoReader, err := logoFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Không thể đọc file logo: " + err.Error()})
			return
		}
		defer logoReader.Close()

		logoContent, err := io.ReadAll(logoReader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Không thể đọc nội dung logo: " + err.Error()})
			return
		}

		// Upload logo to default bucket
		bucketName := s.config.SupabaseConfig.Bucket
		if bucketName == "" {
			bucketName = "images"
		}
		uploadedLogoURL, err := s.uploadFileToSupabase(ctx, logoContent, logoFile.Filename, bucketName, "supplier-logos", contentType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Không thể upload logo: " + err.Error()})
			return
		}
		logoURL = &uploadedLogoURL
	}

	// Handle business license upload (PDF)
	var giayToKinhDoanhURL *string
	giayToFile, err := c.FormFile("giay_to_kinh_doanh")
	if err == nil && giayToFile != nil {
		// Validate PDF type
		contentType := giayToFile.Header.Get("Content-Type")
		if contentType != "application/pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Giấy phép kinh doanh phải là file PDF"})
			return
		}

		// Open and read PDF file
		pdfReader, err := giayToFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Không thể đọc file giấy phép kinh doanh: " + err.Error()})
			return
		}
		defer pdfReader.Close()

		pdfContent, err := io.ReadAll(pdfReader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Không thể đọc nội dung giấy phép kinh doanh: " + err.Error()})
			return
		}

		// Upload PDF to giay_phep_kinh_doanh bucket
		uploadedPDFURL, err := s.uploadFileToSupabase(ctx, pdfContent, giayToFile.Filename, "giay_phep_kinh_doanh", "", contentType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Không thể upload giấy phép kinh doanh: " + err.Error()})
			return
		}
		giayToKinhDoanhURL = &uploadedPDFURL
	}

	// Prepare optional string pointers
	var soDienThoaiPtr *string
	if soDienThoai != "" {
		soDienThoaiPtr = &soDienThoai
	}

	var diaChiPtr *string
	if diaChi != "" {
		diaChiPtr = &diaChi
	}

	var websitePtr *string
	if website != "" {
		websitePtr = &website
	}

	var moTaPtr *string
	if moTa != "" {
		moTaPtr = &moTa
	}

	var thanhPhoPtr *string
	if thanhPho != "" {
		thanhPhoPtr = &thanhPho
	}

	var quocGiaPtr *string
	if quocGia != "" {
		quocGiaPtr = &quocGia
	}

	var maSoThuePtr *string
	if maSoThue != "" {
		maSoThuePtr = &maSoThue
	}

	var soNhanVienPtr *string
	if soNhanVien != "" {
		soNhanVienPtr = &soNhanVien
	}

	// Create supplier with user
	result, err := s.z.CreateSupplierWithUser(context.Background(), db.CreateSupplierWithUserParams{
		CreateUserParams: db.CreateUserParams{
			HoTen:        hoTen,
			Email:        email,
			MatKhauMaHoa: hashedPassword,
			SoDienThoai:  soDienThoaiPtr,
			VaiTro:       db.NullVaiTroNguoiDung{VaiTroNguoiDung: db.VaiTroNguoiDung(db.VaiTroNguoiDungNhaCungCap), Valid: true},
			DangHoatDong: helpers.NewBool(true),
			XacThuc:      helpers.NewBool(false),
			NgayTao:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			NgayCapNhat:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
		CreateSupplierParams: db.CreateSupplierParams{
			Ten:             ten,
			DiaChi:          diaChiPtr,
			Website:         websitePtr,
			MoTa:            moTaPtr,
			Logo:            logoURL,
			NamThanhLap:     pgtype.Date{Time: namThanhLap, Valid: true},
			ThanhPho:        thanhPhoPtr,
			QuocGia:         quocGiaPtr,
			MaSoThue:        maSoThuePtr,
			SoNhanVien:      soNhanVienPtr,
			GiayToKinhDoanh: giayToKinhDoanhURL,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng ký đối tác thành công. Tài khoản của bạn đang chờ admin duyệt.",
		"data":    result,
	})
}

// cập nhật nhà cung cấp
// @summary Cập nhật nhà cung cấp
// @description Cập nhật nhà cung cấp
// @tags Supplier
// @accept json
// @produce json
// @param req body db.UpdateSupplierParams true "Thông tin cập nhật"
// @success 200 {object} gin.H "Thành công"
// @failure 400 {object} gin.H "Lỗi yêu cầu không hợp lệ"
// @failure 500 {object} gin.H "Lỗi server"
// @router /supplier/updateSupplier [put]
func (s *Server) UpdateSupplier(c *gin.Context) {
	var req db.UpdateSupplierParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	result, err := s.z.UpdateSupplier(context.Background(), db.UpdateSupplierParams{
		ID:      req.ID,
		Ten:     req.Ten,
		DiaChi:  req.DiaChi,
		Website: req.Website,
		MoTa:    req.MoTa,
		Logo:    req.Logo,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier updated successfully", "data": result})
}

// lấy danh sách tour của nhà cung cấp

// @Summary lấy danh sách tour của nhà cung cấp
// @Description lấy danh sách tour của nhà cung cấp
// @Tags Supplier
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param trang_thai query string false "Trạng thái" default("")
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/tours/my [get]
func (s *Server) GetMyTours(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	trang_thai := c.Query("trang_thai")

	if limit == 0 {
		limit = 10
	}
	if offset == 0 {
		offset = 0
	}

	// Validate trang_thai: giá trị rỗng "" là hợp lệ (lấy tất cả tours)
	// Nếu có giá trị, chỉ chấp nhận: nhap, cong_bo, luu_tru
	var trangThaiPtr *string
	if trang_thai != "" {
		validStatuses := map[string]bool{
			"nhap":    true,
			"cong_bo": true,
			"luu_tru": true,
		}
		if !validStatuses[trang_thai] {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Trạng thái không hợp lệ. Chỉ chấp nhận: nhap, cong_bo, luu_tru hoặc rỗng để lấy tất cả",
			})
			return
		}
		trangThaiPtr = &trang_thai
	} else {
		// Khi trang_thai == "", vẫn truyền pointer với giá trị rỗng để SQL query xử lý đúng
		emptyStr := ""
		trangThaiPtr = &emptyStr
	}

	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	data, err := s.z.GetMyTours(context.Background(), db.GetMyToursParams{
		NhaCungCapID: claimsMap.Id,
		Limit:        int32(limit),
		Offset:       int32(offset),
		TrangThai:    trangThaiPtr,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tours fetched successfully", "data": data})
}

// Cập nhật trạng thái tour
// @Summary Cập nhật trạng thái tour
// @Description Cập nhật trạng thái tour
// @Tags Supplier
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param trang_thai body string true "Trạng thái"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/tours/update-status/{id} [put]
func (s *Server) UpdateTourStatus(c *gin.Context) {
	_id := c.Param("id")
	id, err := strconv.Atoi(_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}
	trang_thai := c.Query("trang_thai")
	if trang_thai == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Trạng thái không được để rỗng"})
		return
	}
	var trangThaiPtr *string
	if trang_thai != "" {
		validStatuses := map[string]bool{
			"nhap":    true,
			"cong_bo": true,
			"luu_tru": true,
		}
		if !validStatuses[trang_thai] {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Trạng thái không hợp lệ. Chỉ chấp nhận: nhap, cong_bo, luu_tru hoặc rỗng để lấy tất cả",
			})
			return
		}
		trangThaiPtr = &trang_thai
		result, err := s.z.UpdateTourStatus(context.Background(), db.UpdateTourStatusParams{
			ID:        int32(id),
			TrangThai: trangThaiPtr,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Tour status updated successfully", "data": result})
	}
}

// Lấy thông tin nhà cung cấp
// @Summary Lấy thông tin nhà cung cấp
// @Description Lấy thông tin nhà cung cấp
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/info [get]
func (s *Server) GetInfoSupplier(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	data, err := s.z.GetSupplierById(context.Background(), claimsMap.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier fetched successfully", "data": data})
}

// Lấy tất cả nhà cung cấp bao gồm cả nhà cung cấp đã xóa
// @Summary Lấy tất cả nhà cung cấp bao gồm cả nhà cung cấp đã xóa
// @Description Lấy tất cả nhà cung cấp bao gồm cả nhà cung cấp đã xóa
// @Tags Supplier
// @Accept json
// @Produce json
// @Success 200 {object} db.NhaCungCap
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/all [get]
func (s *Server) GetAllSuppliersIncludingDeleted(c *gin.Context) {
	data, err := s.z.GetAllSuppliersIncludingDeleted(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Suppliers fetched successfully", "data": data})
}

// Lấy nhà cung cấp đang hoạt động
// @Summary Lấy nhà cung cấp đang hoạt động
// @Description Lấy nhà cung cấp đang hoạt động
// @Tags Supplier
// @Accept json
// @Produce json
// @Success 200 {object} db.NhaCungCap
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/active [get]
func (s *Server) GetActiveSuppliers(c *gin.Context) {
	data, err := s.z.GetActiveSuppliers(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Suppliers fetched successfully", "data": data})
}

// Xóa nhà cung cấp
// @Summary Xóa nhà cung cấp
// @Description Xóa nhà cung cấp
// @Tags Supplier
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/delete/{id} [delete]
func (s *Server) DeleteSupplier(c *gin.Context) {
	_id := c.Param("id")
	var id pgtype.UUID
	if err := id.Scan(_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}
	err := s.z.DeleteSupplier(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier deleted successfully"})
}

// Lấy nhà cung cấp với số lượng tour
// @Summary Lấy nhà cung cấp với số lượng tour
// @Description Lấy nhà cung cấp với số lượng tour
// @Tags Supplier
// @Accept json
// @Produce json
// @Success 200 {object} db.GetSupplierWithTourCountRow
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/with-tour-count [get]
func (s *Server) GetSupplierWithTourCount(c *gin.Context) {
	email := c.Query("email")
	data, err := s.z.GetSupplierWithTourCount(context.Background(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier with tour count fetched successfully", "data": data})
}

// Tìm kiếm nhà cung cấp
// @Summary Tìm kiếm nhà cung cấp
// @Description Tìm kiếm nhà cung cấp
// @Tags Supplier
// @Accept json
// @Produce json
// @Param keyword path string true "Keyword"
// @Success 200 {object} db.NhaCungCap
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/search/{keyword} [get]
func (s *Server) SearchSuppliers(c *gin.Context) {
	keyword := c.Param("keyword")
	data, err := s.z.SearchSuppliers(context.Background(), keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Suppliers fetched successfully", "data": data})
}

// Đếm số lượng nhà cung cấp
// @Summary Đếm số lượng nhà cung cấp
// @Description Đếm số lượng nhà cung cấp
// @Tags Supplier
// @Accept json
// @Produce json
// @Success 200 {object} int64
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/count [get]
func (s *Server) CountSuppliers(c *gin.Context) {
	data, err := s.z.CountSuppliers(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Suppliers fetched successfully", "data": data})
}

// Đếm số lượng nhà cung cấp theo trạng thái
// @Summary Đếm số lượng nhà cung cấp theo trạng thái
// @Description Đếm số lượng nhà cung cấp theo trạng thái
// @Tags Supplier
// @Accept json
// @Produce json
// @Success 200 {object} db.CountSuppliersByStatusRow
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/count-by-status [get]
func (s *Server) CountSuppliersByStatus(c *gin.Context) {
	email := c.Query("email")
	data, err := s.z.CountSuppliersByStatus(context.Background(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Suppliers fetched successfully", "data": data})
}

// ===========================================
// DASHBOARD SUPPLIER HANDLERS
// ===========================================

// Lấy tổng quan dashboard
// @Summary Lấy tổng quan dashboard
// @Description Lấy tổng quan dashboard với các thống kê tổng hợp
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/overview [get]
func (s *Server) GetSupplierDashboardOverview(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	data, err := s.z.GetSupplierDashboardOverview(context.Background(), claimsMap.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Dashboard overview fetched successfully", "data": data})
}

// Lấy doanh thu theo khoảng thời gian
// @Summary Lấy doanh thu theo khoảng thời gian
// @Description Lấy doanh thu theo khoảng thời gian (day, week, month)
// @Tags Supplier
// @Accept json
// @Produce json
// @Param period query string true "Period (day, week, month)" default(day)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/revenue-by-time [get]
func (s *Server) GetSupplierRevenueByTimeRange(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	period := c.Query("period")
	if period == "" {
		period = "day"
	}
	validPeriods := map[string]bool{"day": true, "week": true, "month": true}
	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Period must be: day, week, or month"})
		return
	}
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.DateOnly, startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.DateOnly, endDateStr); err == nil {
			endDate = &t
		}
	}
	var startDatePg, endDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}
	data, err := s.z.GetSupplierRevenueByTimeRange(context.Background(), db.GetSupplierRevenueByTimeRangeParams{
		ID:      claimsMap.Id,
		Column2: period,
		Column3: startDatePg,
		Column4: endDatePg,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error(), "error": "loiccc"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Revenue by time range fetched successfully", "data": data})
}

// Lấy top tours bán chạy nhất
// @Summary Lấy top tours bán chạy nhất
// @Description Lấy top tours bán chạy nhất theo doanh thu, số booking hoặc đánh giá
// @Tags Supplier
// @Accept json
// @Produce json
// @Param sort_by query string false "Sort by (revenue, bookings, rating)" default(revenue)
// @Param limit query int false "Limit" default(10)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/top-tours [get]
func (s *Server) GetSupplierTopTours(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	sortBy := c.Query("sort_by")
	if sortBy == "" {
		sortBy = "revenue"
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}
	var startDatePg, endDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}
	data, err := s.z.GetSupplierTopTours(context.Background(), db.GetSupplierTopToursParams{
		ID:      claimsMap.Id,
		Column2: startDatePg,
		Column3: endDatePg,
		Column4: sortBy,
		Limit:   int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Top tours fetched successfully", "data": data})
}

// Lấy thống kê booking theo trạng thái
// @Summary Lấy thống kê booking theo trạng thái
// @Description Lấy thống kê booking theo trạng thái và thời gian
// @Tags Supplier
// @Accept json
// @Produce json
// @Param period query string true "Period (day, week, month)" default(day)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/booking-stats [get]
func (s *Server) GetSupplierBookingStatsByStatus(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}
	var startDatePg, endDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}
	data, err := s.z.GetSupplierBookingStatsByStatus(context.Background(), db.GetSupplierBookingStatsByStatusParams{
		ID:        claimsMap.Id,
		Column2:   "day",
		NgayDat:   startDatePg,
		NgayDat_2: endDatePg,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Booking stats fetched successfully", "data": data})
}

// Lấy thống kê tour theo trạng thái
// @Summary Lấy thống kê tour theo trạng thái
// @Description Lấy thống kê tour theo trạng thái
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/tour-stats [get]
func (s *Server) GetSupplierTourStatsByStatus(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	data, err := s.z.GetSupplierTourStatsByStatus(context.Background(), claimsMap.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tour stats fetched successfully", "data": data})
}

// Lấy biểu đồ doanh thu
// @Summary Lấy biểu đồ doanh thu
// @Description Lấy dữ liệu doanh thu theo thời gian cho biểu đồ
// @Tags Supplier
// @Accept json
// @Produce json
// @Param period query string true "Period (day, week, month)" default(day)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/revenue-chart [get]
func (s *Server) GetSupplierRevenueChart(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	period := c.Query("period")
	if period == "" {
		period = "day"
	}
	validPeriods := map[string]bool{"day": true, "week": true, "month": true}
	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Period must be: day, week, or month"})
		return
	}
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}
	var startDatePg, endDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}
	data, err := s.z.GetSupplierRevenueChart(context.Background(), db.GetSupplierRevenueChartParams{
		ID:      claimsMap.Id,
		Column2: period,
		Column3: startDatePg,
		Column4: endDatePg,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Revenue chart data fetched successfully", "data": data})
}

// Lấy thống kê tour theo danh mục
// @Summary Lấy thống kê tour theo danh mục
// @Description Lấy thống kê số lượng tour theo từng danh mục của supplier
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/tour-stats-by-category [get]
func (s *Server) GetSupplierTourStatsByCategory(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}

	data, err := s.z.GetSupplierTourStatsByCategory(context.Background(), claimsMap.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tour stats by category fetched successfully", "data": data})
}

// Lấy thống kê khách hàng
// @Summary Lấy thống kê khách hàng
// @Description Lấy top khách hàng theo số lần đặt hoặc tổng tiền
// @Tags Supplier
// @Accept json
// @Produce json
// @Param sort_by query string false "Sort by (spent, bookings)" default(spent)
// @Param limit query int false "Limit" default(10)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/customer-stats [get]
func (s *Server) GetSupplierCustomerStats(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	sortBy := c.Query("sort_by")
	if sortBy == "" {
		sortBy = "spent"
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}
	var startDatePg, endDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}
	data, err := s.z.GetSupplierCustomerStats(context.Background(), db.GetSupplierCustomerStatsParams{
		ID:      claimsMap.Id,
		Column2: startDatePg,
		Column3: endDatePg,
		Column4: sortBy,
		Limit:   int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer stats fetched successfully", "data": data})
}
// Lấy phân tích đánh giá
// @Summary Lấy phân tích đánh giá
// @Description Lấy phân tích đánh giá tour với số lượng theo từng sao
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/rating-analysis [get]
func (s *Server) GetSupplierRatingAnalysis(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	data, err := s.z.GetSupplierRatingAnalysis(context.Background(), claimsMap.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rating analysis fetched successfully", "data": data})
}

// Lấy lịch khởi hành sắp tới
// @Summary Lấy lịch khởi hành sắp tới
// @Description Lấy danh sách lịch khởi hành sắp tới
// @Tags Supplier
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/upcoming-departures [get]
func (s *Server) GetSupplierUpcomingDepartures(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	data, err := s.z.GetSupplierUpcomingDepartures(context.Background(), db.GetSupplierUpcomingDeparturesParams{
		ID:    claimsMap.Id,
		Limit: int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Upcoming departures fetched successfully", "data": data})
}

// Lấy booking gần đây
// @Summary Lấy booking gần đây
// @Description Lấy danh sách booking gần đây nhất
// @Tags Supplier
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/recent-bookings [get]
func (s *Server) GetSupplierRecentBookings(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	data, err := s.z.GetSupplierRecentBookings(context.Background(), db.GetSupplierRecentBookingsParams{
		ID:    claimsMap.Id,
		Limit: int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recent bookings fetched successfully", "data": data})
}

// Lấy thống kê doanh thu tổng hợp
// @Summary Lấy thống kê doanh thu tổng hợp
// @Description Lấy thống kê doanh thu tổng hợp: tổng doanh thu, doanh thu tháng này, tháng trước, số đặt chỗ, trung bình/đơn
// @Tags Supplier
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/revenue/statistics [get]
func (s *Server) GetSupplierRevenueStatistics(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}

	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}
	var startDatePg, endDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}

	data, err := s.z.GetSupplierRevenueStatistics(context.Background(), db.GetSupplierRevenueStatisticsParams{
		ID:      claimsMap.Id,
		Column2: startDatePg,
		Column3: endDatePg,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Tính tỷ lệ tăng trưởng
	var tyLeTangTruong float64
	if data.DoanhThuThangTruoc.Valid {
		prevRevenue, _ := data.DoanhThuThangTruoc.Float64Value()
		if prevRevenue.Float64 > 0 {
			currentRevenue := float64(0)
			if data.DoanhThuThangNay.Valid {
				current, _ := data.DoanhThuThangNay.Float64Value()
				currentRevenue = current.Float64
			}
			tyLeTangTruong = ((currentRevenue - prevRevenue.Float64) / prevRevenue.Float64) * 100
		}
	}

	// Tính doanh thu trung bình/đơn
	var doanhThuTrungBinhDon float64
	if data.SoDatCho > 0 && data.DoanhThuTrongKy.Valid {
		revenue, _ := data.DoanhThuTrongKy.Float64Value()
		doanhThuTrungBinhDon = revenue.Float64 / float64(data.SoDatCho)
	}

	// Format response
	response := gin.H{
		"tong_doanh_thu":              getNumericValue(data.TongDoanhThu),
		"doanh_thu_thang_nay":         getNumericValue(data.DoanhThuThangNay),
		"doanh_thu_thang_truoc":       getNumericValue(data.DoanhThuThangTruoc),
		"ty_le_tang_truong":           tyLeTangTruong,
		"so_dat_cho":                  data.SoDatCho,
		"doanh_thu_trung_binh_don":    doanhThuTrungBinhDon,
	}

	c.JSON(http.StatusOK, gin.H{"message": "Revenue statistics fetched successfully", "data": response})
}

// Helper function để lấy giá trị numeric
func getNumericValue(n pgtype.Numeric) float64 {
	if n.Valid {
		val, _ := n.Float64Value()
		return val.Float64
	}
	return 0
}

// Lấy danh sách giao dịch (transactions)
// @Summary Lấy danh sách giao dịch
// @Description Lấy danh sách giao dịch đã thanh toán với thông tin chi tiết
// @Tags Supplier
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/revenue/transactions [get]
func (s *Server) GetSupplierTransactions(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}

	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}
	var startDatePg, endDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	if offset < 0 {
		offset = 0
	}

	data, err := s.z.GetSupplierTransactions(context.Background(), db.GetSupplierTransactionsParams{
		ID:      claimsMap.Id,
		Column2: startDatePg,
		Column3: endDatePg,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Format response
	transactions := make([]gin.H, 0, len(data))
	for _, item := range data {
		transactions = append(transactions, gin.H{
			"id":                 item.ID,
			"ma_dat_cho":         item.MaDatCho,
			"tour_tieu_de":       item.TourTieuDe,
			"nguoi_dung_ten":     item.NguoiDungTen,
			"so_tien":            getNumericValue(item.SoTien),
			"phi_dich_vu":        getNumericValue(item.PhiDichVu),
			"so_tien_thuc_nhan":  getNumericValue(item.SoTienThucNhan),
			"ngay_thanh_toan":    item.NgayThanhToan,
			"trang_thai":         item.TrangThai,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transactions fetched successfully", "data": transactions})
}

// Lấy danh sách đặt chỗ theo trạng thái với filter nâng cao
// @Summary Lấy danh sách đặt chỗ theo trạng thái với filter nâng cao
// @Description Lấy danh sách đặt chỗ với nhiều filter: trạng thái, tour, thời gian, search, pagination, sort
// @Tags Supplier
// @Accept json
// @Produce json
// @Param trang_thai query string false "Trạng thái (cho_xac_nhan, da_xac_nhan, da_thanh_toan, hoan_thanh, da_huy)"
// @Param tour_id query int false "Tour ID"
// @Param start_date query string false "Ngày đặt chỗ từ (YYYY-MM-DD)"
// @Param end_date query string false "Ngày đặt chỗ đến (YYYY-MM-DD)"
// @Param departure_start_date query string false "Ngày khởi hành từ (YYYY-MM-DD)"
// @Param departure_end_date query string false "Ngày khởi hành đến (YYYY-MM-DD)"
// @Param search_keyword query string false "Tìm kiếm theo tên khách hàng, email hoặc tên tour"
// @Param phuong_thuc_thanh_toan query string false "Phương thức thanh toán"
// @Param min_amount query number false "Số tiền tối thiểu"
// @Param max_amount query number false "Số tiền tối đa"
// @Param sort_by query string false "Sắp xếp (ngay_dat_asc, ngay_dat_desc, tong_tien_asc, tong_tien_desc, ngay_khoi_hanh_asc, ngay_khoi_hanh_desc)" default(ngay_dat_desc)
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/bookings/advanced [get]
func (s *Server) GetSupplierBookingsByStatusAdvanced(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	if offset < 0 {
		offset = 0
	}

	trangThai := c.Query("trang_thai")
	var trangThaiPtr *string
	if trangThai != "" {
		trangThaiPtr = &trangThai
	}

	tourIDStr := c.Query("tour_id")
	var tourIDPtr *int32
	if tourIDStr != "" {
		if tourID, err := strconv.Atoi(tourIDStr); err == nil {
			tourIDInt32 := int32(tourID)
			tourIDPtr = &tourIDInt32
		}
	}

	// Parse dates
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}
	var startDatePg, endDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}

	// Parse departure dates
	var departureStartDate, departureEndDate *time.Time
	if departureStartDateStr := c.Query("departure_start_date"); departureStartDateStr != "" {
		if t, err := time.Parse("2006-01-02", departureStartDateStr); err == nil {
			departureStartDate = &t
		}
	}
	if departureEndDateStr := c.Query("departure_end_date"); departureEndDateStr != "" {
		if t, err := time.Parse("2006-01-02", departureEndDateStr); err == nil {
			departureEndDate = &t
		}
	}
	var departureStartDatePg, departureEndDatePg pgtype.Date
	if departureStartDate != nil {
		departureStartDatePg = pgtype.Date{Time: *departureStartDate, Valid: true}
	}
	if departureEndDate != nil {
		departureEndDatePg = pgtype.Date{Time: *departureEndDate, Valid: true}
	}

	searchKeyword := c.Query("search_keyword")
	var searchKeywordPtr *string
	if searchKeyword != "" {
		searchKeywordPtr = &searchKeyword
	}

	phuongThucThanhToan := c.Query("phuong_thuc_thanh_toan")
	var phuongThucThanhToanPtr *string
	if phuongThucThanhToan != "" {
		phuongThucThanhToanPtr = &phuongThucThanhToan
	}

	// Parse amount filters
	var minAmount, maxAmount pgtype.Numeric
	if minAmountStr := c.Query("min_amount"); minAmountStr != "" {
		if amount, err := strconv.ParseFloat(minAmountStr, 64); err == nil {
			// Use pgtype.Numeric with proper initialization
			var num pgtype.Numeric
			if err := num.Scan(amount); err == nil {
				minAmount = num
			}
		}
	}
	if maxAmountStr := c.Query("max_amount"); maxAmountStr != "" {
		if amount, err := strconv.ParseFloat(maxAmountStr, 64); err == nil {
			var num pgtype.Numeric
			if err := num.Scan(amount); err == nil {
				maxAmount = num
			}
		}
	}

	sortBy := c.Query("sort_by")
	if sortBy == "" {
		sortBy = "ngay_dat_desc"
	}
	var sortByPtr *string
	if sortBy != "" {
		sortByPtr = &sortBy
	}

	// Fetch data
	data, err := s.z.GetSupplierBookingsByStatusAdvanced(context.Background(), db.GetSupplierBookingsByStatusAdvancedParams{
		ID:                  claimsMap.Id,
		Limit:               int32(limit),
		Offset:              int32(offset),
		TrangThai:           trangThaiPtr,
		TourID:              tourIDPtr,
		StartDate:           startDatePg,
		EndDate:             endDatePg,
		DepartureStartDate:  departureStartDatePg,
		DepartureEndDate:    departureEndDatePg,
		SearchKeyword:       searchKeywordPtr,
		PhuongThucThanhToan: phuongThucThanhToanPtr,
		MinAmount:           minAmount,
		MaxAmount:           maxAmount,
		SortBy:              sortByPtr,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get total count
	totalCount, err := s.z.CountSupplierBookingsByStatusAdvanced(context.Background(), db.CountSupplierBookingsByStatusAdvancedParams{
		ID:                  claimsMap.Id,
		TrangThai:           trangThaiPtr,
		TourID:              tourIDPtr,
		StartDate:           startDatePg,
		EndDate:             endDatePg,
		DepartureStartDate:  departureStartDatePg,
		DepartureEndDate:    departureEndDatePg,
		SearchKeyword:       searchKeywordPtr,
		PhuongThucThanhToan: phuongThucThanhToanPtr,
		MinAmount:           minAmount,
		MaxAmount:           maxAmount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Bookings fetched successfully",
		"data":     data,
		"total":    totalCount,
		"limit":    limit,
		"offset":   offset,
		"has_more": (offset + limit) < int(totalCount),
	})
}

// Thống kê chi tiết các chỉ số đánh giá của nhà cung cấp
// @Summary Thống kê chi tiết các chỉ số đánh giá của nhà cung cấp
// @Description Thống kê chi tiết các chỉ số đánh giá của nhà cung cấp
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tour_id query int false "Tour ID" default(0)
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/review-statistics [get]
func (s *Server) GetSupplierReviewStatistics(c *gin.Context) {
	tourID, _ := strconv.Atoi(c.Query("tour_id"))
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	data, err := s.z.GetSupplierReviewStatistics(context.Background(), db.GetSupplierReviewStatisticsParams{
		ID:      claimsMap.Id,
		Column2: int32(tourID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Review fetched successfully", "data": data})
}

// Lấy danh sách đánh giá chi tiết với các bộ lọc theo sao và tour
// @Summary Lấy danh sách đánh giá chi tiết với các bộ lọc theo sao và tour
// @Description Lấy danh sách đánh giá chi tiết với các bộ lọc theo sao và tour
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param rating query int false "Rating" default(0)
// @Param tour_id query int false "Tour ID" default(0)
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/reviews [get]
func (s *Server) GetDetailedSupplierReviews(c *gin.Context) {
	rating, _ := strconv.Atoi(c.Query("rating"))
	tourID, _ := strconv.Atoi(c.Query("tour_id"))
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	data, err := s.z.GetDetailedSupplierReviews(context.Background(), db.GetDetailedSupplierReviewsParams{
		NhaCungCapID: claimsMap.Id,
		Column2:      int32(rating),
		Column3:      int32(tourID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reviews fetched successfully", "data": data})
}

// Lấy danh sách tour của nhà cung cấp
// @Summary Lấy danh sách tour của nhà cung cấp
// @Description Lấy danh sách tour của nhà cung cấp
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/options-tour [get]
func (s *Server) GetOptionTour(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	data, err := s.z.OptionTour(context.Background(), pgtype.UUID(claimsMap.Id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tours fetched successfully", "data": data})

}

// Phản hồi đánh giá
// @Summary Phản hồi đánh giá
// @Description Phản hồi đánh giá
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param danh_gia_id path int true "Danh gia ID"
// @Param noi_dung body string true "Noi dung phan hoi"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /supplier/dashboard/feedback-review/{danh_gia_id} [post]
func (s *Server) FeedbackReview(c *gin.Context) {
	danhGiaID, _ := strconv.Atoi(c.Param("danh_gia_id"))
	noiDung := c.PostForm("noi_dung")
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if claimsMap.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		return
	}
	id, err := s.z.FeedbackReview(context.Background(), db.FeedbackReviewParams{
		DanhGiaID:   int32(danhGiaID),
		NguoiDungID: claimsMap.Id,
		NoiDung:     noiDung,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Feedback review created successfully", "data": id})
}
