package handler

import (
	"context"
	"encoding/json"
	"fmt"
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

// lấy danh mục tour
// @summary Lấy danh mục tour
// @description Lấy danh mục tour
// @tags tour
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/categories [get]
func (s *Server) GetAllTourCategory(c *gin.Context) {
	tourCategories, err := s.z.GetAllTourCategory(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy danh mục tour",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh mục tour thành công",
		"data":    tourCategories,
	})
}

// lấy danh sách tour
// @summary Lấy danh sách tour
// @description Lấy danh sách tour
// @tags tour
// @accept json
// @produce json
// @param limit query int false "Limit" default(10)
// @param offset query int false "Offset" default(1)
// @success 200 {object} db.GetAllTourRow "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/ [get]
func (s *Server) GetAllTour(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	if limit == 0 {
		limit = 10
	}
	if offset == 0 {
		offset = 0
	}
	tours, err := s.z.GetAllTour(context.Background(), db.GetAllTourParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy danh sách tour",
			"message": err.Error(),
		})
		return
	}
	fmt.Println("ok")
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh sách tour thành công",
		"data":    tours,
	})
}

// lấy chi tiết tour
// @summary Lấy chi tiết tour
// @description Lấy chi tiết tour
// @tags tour
// @accept json
// @produce json
// @param id path int true "ID tour"
// @success 200 {object} db.GetTourDetailByIDRow "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/{id} [get]
func (s *Server) GetTourDetailByID(c *gin.Context) {
	_id := c.Param("id")
	id, err := strconv.ParseInt(_id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "ID không hợp lệ",
		})
		return
	}

	// Create context with timeout for database query
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	tour, err := s.z.GetTourDetailByID(ctx, int32(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"error":   "Không thể lấy chi tiết tour",
		})
		return
	}

	// Convert pgtype.UUID to string for nha_cung_cap_id
	var nhaCungCapIDStr *string
	if tour.NhaCungCapID.Valid {
		uuidStr := tour.NhaCungCapID.String()
		nhaCungCapIDStr = &uuidStr
	}

	response := gin.H{
		"id":                 tour.ID,
		"tieu_de":            tour.TieuDe,
		"mo_ta":              tour.MoTa,
		"danh_muc_id":        tour.DanhMucID,
		"so_ngay":            tour.SoNgay,
		"so_dem":             tour.SoDem,
		"gia_nguoi_lon":      tour.GiaNguoiLon,
		"gia_tre_em":         tour.GiaTreEm,
		"don_vi_tien_te":     tour.DonViTienTe,
		"trang_thai":         tour.TrangThai,
		"noi_bat":            tour.NoiBat,
		"nha_cung_cap_id":    nhaCungCapIDStr,
		"ngay_tao":           tour.NgayTao,
		"ngay_cap_nhat":      tour.NgayCapNhat,
		"ten_danh_muc":       tour.TenDanhMuc,
		"ten_nha_cung_cap":   tour.TenNhaCungCap,
		"logo_ncc":           tour.LogoNcc,
		"hinh_anh":           json.RawMessage(tour.Images),
		"diem_den":           json.RawMessage(tour.Destinations),
		"lich_trinh":         json.RawMessage(tour.Itinerary),
		"lich_khoi_hanh":     json.RawMessage(tour.Departures),
		"giam_gia_phan_tram": tour.GiamGiaPhanTram,
		"giam_gia_tu":        tour.GiamGiaTu,
		"giam_gia_den":       tour.GiamGiaDen,
		"so_nho_nhat":        tour.SoNhoNhat,
		"so_lon_nhat":        tour.SoLonNhat,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy chi tiết tour thành công",
		"data":    response,
	})
}

// ==================== CREATE TOUR WITH FULL DETAILS ====================

// CreateTourFull godoc
// @Summary      Tạo tour với đầy đủ thông tin (1 transaction)
// @Description  Tạo tour bao gồm ảnh, điểm đến, lịch trình và hoạt động trong 1 transaction. Đảm bảo tính toàn vẹn dữ liệu (all or nothing)
// @Tags         tour
// @Accept       json
// @Produce      json
// @Param        request body models.CreateTourFullRequest true "Tour data"
// @Success      201 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /tour [post]
func (s *Server) CreateTour(c *gin.Context) {
	var req models.CreateTourRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu không hợp lệ",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from JWT claims
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	jwtClaims, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication claims"})
		return
	}

	userUUID := jwtClaims.Id

	// Convert giá sang pgtype.Numeric
	var giaMoiNguoi pgtype.Numeric
	if err := giaMoiNguoi.Scan(fmt.Sprintf("%.2f", req.GiaNguoiLon)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Giá không hợp lệ",
		})
		return
	}

	// Tạo params cho transaction
	trangThai := req.TrangThai
	if trangThai == "" {
		trangThai = "nhap" // Default status
	}

	donViTienTe := req.DonViTienTe
	if donViTienTe == "" {
		donViTienTe = "VND"
	}
	var giaNguoiLon, giaTreEm pgtype.Numeric
	if err := giaNguoiLon.Scan(fmt.Sprintf("%.2f", req.GiaNguoiLon)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Giá không hợp lệ",
		})
		return
	}
	if err := giaTreEm.Scan(fmt.Sprintf("%.2f", req.GiaTreEm)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Giá không hợp lệ",
		})
		return
	}
	params := db.CreateTourWithDetailsParams{
		Tour: db.CreateTourParams{
			TieuDe:       req.TieuDe,
			MoTa:         &req.MoTa,
			DanhMucID:    &req.DanhMucID,
			SoNgay:       req.SoNgay,
			SoDem:        req.SoDem,
			GiaNguoiLon:  giaNguoiLon,
			GiaTreEm:     giaTreEm,
			DonViTienTe:  &donViTienTe,
			TrangThai:    &trangThai,
			NoiBat:       &req.NoiBat,
			NhaCungCapID: userUUID,
			DangHoatDong: helpers.NewBool(true),
		},
		HinhAnhTours:      make([]db.TourImageInput, 0),
		DiaDiemTours:      make([]db.TourDestinationInput, 0),
		LichTrinhTours:    make([]db.ItineraryWithActivitiesInput, 0),
		LichKhoiHanhTours: make([]db.DepartureInput, 0),
	}

	// Convert uploaded images to tour images
	for _, img := range req.HinhAnhTours {
		params.HinhAnhTours = append(params.HinhAnhTours, db.TourImageInput{
			Link:         img.Link,
			MoTaAlt:      stringPtrIfNotEmpty(img.MoTaAlt),
			LaAnhChinh:   &img.LaAnhChinh,
			ThuTuHienThi: &img.ThuTuHienThi,
		})
	}

	// Convert destinations
	for _, dest := range req.DiaDiemTours {
		params.DiaDiemTours = append(params.DiaDiemTours, db.TourDestinationInput{
			DiemDenID:     dest.DiemDenID,
			ThuTuThamQuan: &dest.ThuTuThamQuan,
		})
	}

	// Convert itineraries with activities
	for _, itin := range req.LichTrinhTours {
		itinInput := db.ItineraryWithActivitiesInput{
			NgayThu:        itin.NgayThu,
			TieuDe:         itin.TieuDe,
			MoTa:           stringPtrIfNotEmpty(itin.MoTa),
			GioBatDau:      stringPtrIfNotEmpty(itin.GioBatDau),
			GioKetThuc:     stringPtrIfNotEmpty(itin.GioKetThuc),
			DiaDiem:        stringPtrIfNotEmpty(itin.DiaDiem),
			ThongTinLuuTru: stringPtrIfNotEmpty(itin.ThongTinLuuTru),
			Activities:     make([]db.ActivityInput, 0),
		}

		// Convert activities
		for _, act := range itin.HoatDongTours {
			itinInput.Activities = append(itinInput.Activities, db.ActivityInput{
				Ten:        act.Ten,
				GioBatDau:  stringPtrIfNotEmpty(act.GioBatDau),
				GioKetThuc: stringPtrIfNotEmpty(act.GioKetThuc),
				MoTa:       stringPtrIfNotEmpty(act.MoTa),
				ThuTu:      &act.ThuTu,
			})
		}

		params.LichTrinhTours = append(params.LichTrinhTours, itinInput)
	}

	// Convert group config (if provided)
	if req.CauHinhNhomTours != nil {
		params.CauHinhNhomTours = &db.GroupConfigInput{
			SoNhoNhat: &req.CauHinhNhomTours.SoNhoNhat,
			SoLonNhat: &req.CauHinhNhomTours.SoLonNhat,
		}
	}

	// Convert departures (if provided)
	for _, dep := range req.LichKhoiHanhTours {
		depInput := db.DepartureInput{
			NgayKhoiHanh: dep.NgayKhoiHanh,
			NgayKetThuc:  dep.NgayKetThuc,
			SucChua:      dep.SucChua,
			TrangThai:    stringPtrIfNotEmpty(dep.TrangThai),
			GhiChu:       stringPtrIfNotEmpty(dep.GhiChu),
		}
		params.LichKhoiHanhTours = append(params.LichKhoiHanhTours, depInput)
	}

	// Execute transaction
	result, err := s.z.CreateTourWithDetails(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể tạo tour",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tạo tour thành công",
		"data": gin.H{
			"tour_id":            result.Tour.ID,
			"tour":               result.Tour,
			"images_count":       len(result.Images),
			"destinations_count": len(result.Destinations),
			"itineraries_count":  len(result.Itineraries),
			"departures_count":   len(result.Departures),
			"images":             result.Images,
			"itineraries":        result.Itineraries,
			"departures":         result.Departures,
		},
	})
}

// Helper functions
func stringPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// lấy danh sách tour theo filter
// @summary Lấy danh sách tour theo filter
// @description Lấy danh sách tour theo filter
// @tags tour
// @accept json
// @produce json
// @param filterTour query models.FilterToursRequest true "Filter tour"
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/filter [get]
func (s *Server) FilterTours(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	req := db.FilterToursParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// Chỉ set DanhMucID nếu có giá trị hợp lệ
	if danhMucIDStr := c.Query("danh_muc_id"); danhMucIDStr != "" {
		if danhMucID, err := strconv.Atoi(danhMucIDStr); err == nil && danhMucID > 0 {
			req.DanhMucID = helpers.Int32Ptr(int32(danhMucID))
		}
	}

	// Chỉ set GiaMin nếu có giá trị hợp lệ
	if giaMinStr := c.Query("gia_min"); giaMinStr != "" {
		if giaMin, err := strconv.ParseFloat(giaMinStr, 64); err == nil && giaMin > 0 {
			giamin := pgtype.Numeric{}
			_ = giamin.Scan(giaMin)
			req.GiaMin = giamin
		}
	}

	// Chỉ set GiaMax nếu có giá trị hợp lệ
	if giaMaxStr := c.Query("gia_max"); giaMaxStr != "" {
		if giaMax, err := strconv.ParseFloat(giaMaxStr, 64); err == nil && giaMax > 0 {
			giamax := pgtype.Numeric{}
			_ = giamax.Scan(giaMax)
			req.GiaMax = giamax
		}
	}

	// Chỉ set SoNgayMin nếu có giá trị hợp lệ
	if soNgayMinStr := c.Query("so_ngay_min"); soNgayMinStr != "" {
		if soNgayMin, err := strconv.Atoi(soNgayMinStr); err == nil && soNgayMin > 0 {
			req.SoNgayMin = helpers.Int32Ptr(int32(soNgayMin))
		}
	}

	// Chỉ set SoNgayMax nếu có giá trị hợp lệ
	if soNgayMaxStr := c.Query("so_ngay_max"); soNgayMaxStr != "" {
		if soNgayMax, err := strconv.Atoi(soNgayMaxStr); err == nil && soNgayMax > 0 {
			req.SoNgayMax = helpers.Int32Ptr(int32(soNgayMax))
		}
	}

	// Chỉ set RatingMin nếu có giá trị hợp lệ
	if ratingMinStr := c.Query("rating_min"); ratingMinStr != "" {
		if ratingMin, err := strconv.ParseFloat(ratingMinStr, 64); err == nil && ratingMin > 0 {
			req.RatingMin = &ratingMin
		}
	}

	// Chỉ set SortBy nếu không rỗng
	if sortBy := c.Query("sort_by"); sortBy != "" {
		req.SortBy = &sortBy
	}

	tours, err := s.z.FilterTours(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy danh sách tour",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh sách tour thành công",
		"data":    tours,
	})
}

// lấy danh sách tour theo tên
// @summary Lấy danh sách tour theo tên
// @description Lấy danh sách tour theo tên
// @tags tour
// @accept json
// @produce json
// @param keyword query string false "Keyword"
// @param diem_den_id query int false "Diem den id"
// @param diem_den_ten query string false "Diem den ten"
// @param so_ngay_min query int false "So ngay min"
// @param so_ngay_max query int false "So ngay max"
// @param so_dem_min query int false "So dem min"
// @param so_dem_max query int false "So dem max"
// @param limit query int false "Limit"
// @param offset query int false "Offset"
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/search [get]
func (s *Server) SearchTours(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	req := db.SearchToursParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// Chỉ set Keyword nếu query không rỗng
	if query := c.Query("query"); query != "" {
		req.Keyword = &query
	}

	// Chỉ set DiemDenTen nếu không rỗng
	if diemDenTen := c.Query("diem_den_ten"); diemDenTen != "" {
		req.DiemDenTen = &diemDenTen
	}

	// Chỉ set DiemDenID nếu có giá trị hợp lệ
	if diemDenIDStr := c.Query("diem_den_id"); diemDenIDStr != "" {
		if diemDenID, err := strconv.Atoi(diemDenIDStr); err == nil && diemDenID > 0 {
			req.DiemDenID = helpers.Int32Ptr(int32(diemDenID))
		}
	}

	// Chỉ set SoNgayMin nếu có giá trị hợp lệ
	if soNgayMinStr := c.Query("so_ngay_min"); soNgayMinStr != "" {
		if soNgayMin, err := strconv.Atoi(soNgayMinStr); err == nil && soNgayMin > 0 {
			req.SoNgayMin = helpers.Int32Ptr(int32(soNgayMin))
		}
	}

	// Chỉ set SoNgayMax nếu có giá trị hợp lệ
	if soNgayMaxStr := c.Query("so_ngay_max"); soNgayMaxStr != "" {
		if soNgayMax, err := strconv.Atoi(soNgayMaxStr); err == nil && soNgayMax > 0 {
			req.SoNgayMax = helpers.Int32Ptr(int32(soNgayMax))
		}
	}

	// Chỉ set SoDemMin nếu có giá trị hợp lệ
	if soDemMinStr := c.Query("so_dem_min"); soDemMinStr != "" {
		if soDemMin, err := strconv.Atoi(soDemMinStr); err == nil && soDemMin > 0 {
			req.SoDemMin = helpers.Int32Ptr(int32(soDemMin))
		}
	}

	// Chỉ set SoDemMax nếu có giá trị hợp lệ
	if soDemMaxStr := c.Query("so_dem_max"); soDemMaxStr != "" {
		if soDemMax, err := strconv.Atoi(soDemMaxStr); err == nil && soDemMax > 0 {
			req.SoDemMax = helpers.Int32Ptr(int32(soDemMax))
		}
	}

	tours, err := s.z.SearchTours(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy danh sách tour",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh sách tour thành công",
		"data":    tours,
	})
}

// tạo khuyến mãi tour
// @summary Tạo khuyến mãi tour
// @description Tạo khuyến mãi tour
// @tags tour
// @accept json
// @produce json
// @param request body models.CreateDiscountTourRequest true "Khuyến mãi tour"
// @security ApiKeyAuth
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/discount [post]
func (s *Server) CreateDiscountTour(c *gin.Context) {
	var req models.CreateDiscountTourRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	var phanTram pgtype.Numeric
	err := phanTram.Scan(fmt.Sprintf("%.2f", req.PhanTram))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Phần trăm không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	ngayBatDau, err := time.Parse(time.DateOnly, req.NgayBatDau)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Ngày bắt đầu không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	ngayKetThuc, err := time.Parse(time.DateOnly, req.NgayKetThuc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Ngày kết thúc không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	result, err := s.z.CreateDiscountTour(c.Request.Context(), db.CreateDiscountTourParams{
		TourID:      req.TourID,
		PhanTram:    phanTram,
		NgayBatDau:  pgtype.Date{Time: ngayBatDau, Valid: true},
		NgayKetThuc: pgtype.Date{Time: ngayKetThuc, Valid: true},
		NgayTao:     pgtype.Timestamp{Time: time.Now(), Valid: true},
		NgayCapNhat: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể tạo khuyến mãi tour",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Tạo khuyến mãi tour thành công",
		"data":    result,
	})
}

// cập nhật khuyến mãi tour
// @summary Cập nhật khuyến mãi tour
// @description Cập nhật khuyến mãi tour
// @tags tour
// @accept json
// @produce json
// @param request body models.UpdateDiscountTourRequest true "Khuyến mãi tour"
// @security ApiKeyAuth
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/discount [put]
func (s *Server) UpdateDiscountTour(c *gin.Context) {
	var req models.UpdateDiscountTourRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dữ liệu không hợp lệ",
		})
		return
	}
	var phanTram pgtype.Numeric
	_ = phanTram.Scan(fmt.Sprintf("%.2f", req.PhanTram))
	ngayBatDau, err := time.Parse(time.DateOnly, req.NgayBatDau)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Ngày bắt đầu không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	ngayKetThuc, err := time.Parse(time.DateOnly, req.NgayKetThuc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Ngày kết thúc không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	result, err := s.z.UpdateDiscountTour(c.Request.Context(), db.UpdateDiscountTourParams{
		ID:          req.ID,
		TourID:      req.TourID,
		PhanTram:    phanTram,
		NgayBatDau:  pgtype.Date{Time: ngayBatDau, Valid: true},
		NgayKetThuc: pgtype.Date{Time: ngayKetThuc, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể cập nhật khuyến mãi tour",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật khuyến mãi tour thành công",
		"data":    result,
	})
}

// xóa khuyến mãi tour
// @summary Xóa khuyến mãi tour
// @description Xóa khuyến mãi tour
// @tags tour
// @accept json
// @produce json
// @param id path int true "ID khuyến mãi tour"
// @ApiKeyAuth ApiKeyAuth
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/discount/{id} [delete]
func (s *Server) DeleteDiscountTour(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID không hợp lệ",
		})
		return
	}
	tourID, err := strconv.Atoi(c.Param("tour_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tour ID không hợp lệ",
		})
		return
	}
	err = s.z.DeleteDiscountTour(c.Request.Context(), db.DeleteDiscountTourParams{
		ID:     int32(id),
		TourID: int32(tourID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể xóa khuyến mãi tour",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Xóa khuyến mãi tour thành công",
	})
}

// lấy khuyến mãi tour theo tour id
// @summary Lấy khuyến mãi tour theo tour id
// @description Lấy khuyến mãi tour theo tour id
// @tags tour
// @accept json
// @produce json
// @param id path int true "ID tour"
// @security ApiKeyAuth
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/discount/{id} [get]
func (s *Server) GetDiscountsByTourID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID không hợp lệ",
		})
		return
	}
	discounts, err := s.z.GetDiscountsByTourID(c.Request.Context(), int32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy khuyến mãi tour",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy khuyến mãi tour thành công",
		"data":    discounts,
	})
}

// cập nhật tour
// @summary Cập nhật tour
// @description Cập nhật tour
// @tags tour
// @accept json
// @produce json
// @param request body models.UpdateTourRequest true "Tour"
// @security ApiKeyAuth
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/{id} [put]
func (s *Server) UpdateTour(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID không hợp lệ",
		})
		return
	}
	var req models.UpdateTourRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu không hợp lệ",
			"message": err.Error(),
		})
		return
	}
	var giaNguoiLon, giaTreEm pgtype.Numeric
	if err := giaNguoiLon.Scan(fmt.Sprintf("%.2f", req.GiaNguoiLon)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Giá không hợp lệ",
		})
		return
	}
	if err := giaTreEm.Scan(fmt.Sprintf("%.2f", req.GiaTreEm)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Giá không hợp lệ",
		})
		return
	}
	var nhaCungCapID pgtype.UUID
	if err := nhaCungCapID.Scan(req.NhaCungCapID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nha cung cap ID không hợp lệ",
		})
		return
	}
	result, err := s.z.UpdateTour(c.Request.Context(), db.UpdateTourParams{
		ID:           int32(id),
		TieuDe:       &req.TieuDe,
		MoTa:         &req.MoTa,
		DanhMucID:    &req.DanhMucID,
		SoNgay:       &req.SoNgay,
		SoDem:        &req.SoDem,
		GiaNguoiLon:  giaNguoiLon,
		GiaTreEm:     giaTreEm,
		DonViTienTe:  &req.DonViTienTe,
		TrangThai:    &req.TrangThai,
		NoiBat:       &req.NoiBat,
		NhaCungCapID: nhaCungCapID,
		DangHoatDong: helpers.NewBool(true),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể cập nhật tour",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật tour thành công",
		"data":    result,
	})
}
