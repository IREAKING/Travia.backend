package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"travia.backend/api/models"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

// ==================== DEPARTURE MANAGEMENT ====================

type CreateDepartureRequest struct {
	TourID       int32   `json:"tour_id" binding:"required"`
	NgayKhoiHanh string  `json:"ngay_khoi_hanh" binding:"required"`
	NgayKetThuc  string  `json:"ngay_ket_thuc" binding:"required"`
	SucChua      int32   `json:"suc_chua" binding:"required,gt=0"`
	TrangThai    string  `json:"trang_thai"`
	GhiChu       *string `json:"ghi_chu"`
}

type UpdateDepartureRequest struct {
	NgayKhoiHanh *string `json:"ngay_khoi_hanh"`
	NgayKetThuc  *string `json:"ngay_ket_thuc"`
	SucChua      *int32  `json:"suc_chua"`
	TrangThai    *string `json:"trang_thai"`
	GhiChu       *string `json:"ghi_chu"`
}

// CreateDeparture godoc
// @Summary      Tạo lịch khởi hành
// @Description  Tạo lịch khởi hành mới cho tour
// @Tags         departure
// @Accept       json
// @Produce      json
// @Param        request body CreateDepartureRequest true "Departure data"
// @Success      201 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/create [post]
func (s *Server) CreateDeparture(c *gin.Context) {
	var req CreateDepartureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse dates
	ngayKhoiHanh, _ := time.Parse("2006-01-02", req.NgayKhoiHanh)
	ngayKetThuc, _ := time.Parse("2006-01-02", req.NgayKetThuc)

	var dateKhoiHanh, dateKetThuc pgtype.Date
	dateKhoiHanh.Scan(ngayKhoiHanh)
	dateKetThuc.Scan(ngayKetThuc)

	trangThai := "len_lich"
	if req.TrangThai != "" {
		trangThai = req.TrangThai
	}

	departure, err := s.z.CreateDeparture(context.Background(), db.CreateDepartureParams{
		TourID:       req.TourID,
		NgayKhoiHanh: dateKhoiHanh,
		NgayKetThuc:  dateKetThuc,
		SucChua:      req.SucChua,
		TrangThai:    db.NullTrangThaiKhoiHanh{TrangThaiKhoiHanh: db.TrangThaiKhoiHanh(trangThai), Valid: true},
		GhiChu:       req.GhiChu,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo lịch khởi hành"})
		fmt.Println("error", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tạo lịch khởi hành thành công",
		"data":    departure,
	})
}

// UpdateLichTrinh godoc
// @Summary      Cập nhật lịch trình
// @Description  Cập nhật lịch trình
// @Tags         departure
// @Accept       json
// @Produce      json
// @Param        id path int true "Lịch trình ID"
// @Param        request body models.UpdateLichTrinhRequest true "Update data"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/lich-trinh/{id} [put]
func (s *Server) UpdateLichTrinh(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID không hợp lệ",
		})
		return
	}
	var req models.UpdateLichTrinhRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dữ liệu không hợp lệ",
		})
		return
	}
	// Parse and convert time strings to pgtype.Time
	var gioBatDau, gioKetThuc pgtype.Time
	if req.GioBatDau != "" {
		if err := gioBatDau.Scan(utils.OnlyTine(req.GioBatDau)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Giờ bắt đầu không hợp lệ: " + err.Error(),
			})
			return
		}
	}
	if req.GioKetThuc != "" {
		if err := gioKetThuc.Scan(utils.OnlyTine(req.GioKetThuc)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Giờ kết thúc không hợp lệ: " + err.Error(),
			})
			return
		}
	}
	result, err := s.z.UpdateLichTrinh(c.Request.Context(), db.UpdateLichTrinhParams{
		ID:             int32(id),
		NgayThu:        &req.NgayThu,
		TieuDe:         &req.TieuDe,
		MoTa:           &req.MoTa,
		GioBatDau:      gioBatDau,
		GioKetThuc:     gioKetThuc,
		DiaDiem:        &req.DiaDiem,
		ThongTinLuuTru: &req.ThongTinLuuTru,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể cập nhật lịch trình",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật lịch trình thành công",
		"data":    result,
	})
}

// UpdateHoatDongTrongNgay godoc
// @Summary      Cập nhật hoạt động trong ngày
// @Description  Cập nhật hoạt động trong ngày
// @Tags         departure
// @Accept       json
// @Produce      json
// @Param        id path int true "Hoạt động trong ngày ID"
// @Param        request body models.UpdateHoatDongTrongNgayRequest true "Update data"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/hoat-dong-trong-ngay/{id} [put]
func (s *Server) UpdateHoatDongTrongNgay(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID không hợp lệ",
		})
		return
	}
	var req models.UpdateHoatDongTrongNgayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dữ liệu không hợp lệ",
		})
		return
	}
	var gioBatDau, gioKetThuc pgtype.Time
	if req.GioBatDau != "" {
		if err := gioBatDau.Scan(utils.OnlyTine(req.GioBatDau)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Giờ bắt đầu không hợp lệ: " + err.Error(),
			})
			return
		}
	}
	if req.GioKetThuc != "" {
		if err := gioKetThuc.Scan(utils.OnlyTine(req.GioKetThuc)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Giờ kết thúc không hợp lệ: " + err.Error(),
			})
			return
		}
	}
	result, err := s.z.UpdateHoatDongTrongNgay(c.Request.Context(), db.UpdateHoatDongTrongNgayParams{
		ID:         int32(id),
		Ten:        &req.Ten,
		GioBatDau:  gioBatDau,
		GioKetThuc: gioKetThuc,
		MoTa:       &req.MoTa,
		ThuTu:      &req.ThuTu,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể cập nhật hoạt động trong ngày",
			"details": err.Error(),
		})
		fmt.Println("error", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật hoạt động trong ngày thành công",
		"data":    result,
	})
}

// GetDepartureByID godoc
// @Summary      Lấy chi tiết lịch khởi hành
// @Description  Lấy thông tin chi tiết của một lịch khởi hành
// @Tags         departure
// @Produce      json
// @Param        id path int true "Departure ID"
// @Success      200 {object} map[string]interface{}
// @Router       /departure/{id} [get]
func (s *Server) GetDepartureByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	departure, err := s.z.GetDepartureByID(context.Background(), int32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy lịch khởi hành"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thành công",
		"data":    departure,
	})
}

// GetDeparturesByTour godoc
// @Summary      Lấy lịch khởi hành theo tour
// @Description  Lấy tất cả lịch khởi hành của một tour
// @Tags         departure
// @Produce      json
// @Param        tour_id path int true "Tour ID"
// @Success      200 {object} map[string]interface{}
// @Router       /departure/tour/{tour_id} [get]
func (s *Server) GetDeparturesByTour(c *gin.Context) {
	tourIDStr := c.Param("tour_id")
	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour ID không hợp lệ"})
		return
	}

	departures, err := s.z.GetDeparturesByTour(context.Background(), int32(tourID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách khởi hành"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thành công",
		"data":    departures,
	})
}

// GetUpcomingDeparturesPublic godoc
// @Summary      Lấy lịch khởi hành sắp tới
// @Description  Lấy danh sách lịch khởi hành trong 30 ngày tới
// @Tags         departure
// @Produce      json
// @Param        limit query int false "Limit" default(20)
// @Success      200 {object} map[string]interface{}
// @Router       /departure/upcoming [get]
func (s *Server) GetUpcomingDeparturesPublic(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	departures, err := s.z.GetUpcomingDeparturesList(context.Background(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách khởi hành"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thành công",
		"data":    departures,
	})
}

// UpdateDeparture godoc
// @Summary      Cập nhật lịch khởi hành
// @Description  Cập nhật thông tin lịch khởi hành
// @Tags         departure
// @Accept       json
// @Produce      json
// @Param        id path int true "Departure ID"
// @Param        request body UpdateDepartureRequest true "Update data"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/{id} [put]
func (s *Server) UpdateDeparture(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	var req UpdateDepartureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ngayKhoiHanh, _ := time.Parse(time.DateOnly, *req.NgayKhoiHanh)
	ngayKetThuc, _ := time.Parse(time.DateOnly, *req.NgayKetThuc)
	var dateKhoiHanh, dateKetThuc pgtype.Date
	dateKhoiHanh.Scan(ngayKhoiHanh)
	dateKetThuc.Scan(ngayKetThuc)
	// TODO: Parse and convert all fields properly
	params := db.UpdateDepartureParams{
		ID:           int32(id),
		NgayKhoiHanh: dateKhoiHanh,
		NgayKetThuc:  dateKetThuc,
		SucChua:      req.SucChua,
		TrangThai:    db.NullTrangThaiKhoiHanh{TrangThaiKhoiHanh: db.TrangThaiKhoiHanh(*req.TrangThai), Valid: true},
		GhiChu:       req.GhiChu,
	}

	departure, err := s.z.UpdateDeparture(context.Background(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể cập nhật lịch khởi hành",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật thành công",
		"data":    departure,
	})
}

// DeleteDeparture godoc
// @Summary      Xóa lịch khởi hành
// @Description  Xóa một lịch khởi hành
// @Tags         departure
// @Produce      json
// @Param        id path int true "Departure ID"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/{id} [delete]
func (s *Server) DeleteDeparture(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	err = s.z.DeleteDeparture(context.Background(), int32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa lịch khởi hành"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xóa lịch khởi hành thành công",
	})
}

// CancelDeparture godoc
// @Summary      Hủy lịch khởi hành
// @Description  Hủy một lịch khởi hành (đặt status = huy)
// @Tags         departure
// @Produce      json
// @Param        id path int true "Departure ID"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/{id}/cancel [put]
func (s *Server) CancelDeparture(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	departure, err := s.z.CancelDeparture(context.Background(), int32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể hủy lịch khởi hành"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hủy lịch khởi hành thành công",
		"data":    departure,
	})
}

// AddHinhAnhTour godoc
// @Summary      Thêm hình ảnh tour
// @Description  Thêm hình ảnh tour
// @Tags         departure
// @Accept       json
// @Produce      json
// @Param        id path int true "Tour ID"
// @Param        request body models.AddHinhAnhTourRequest true "Add data"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/add-hinh-anh/{id} [post]
func (s *Server) AddHinhAnhTour(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}
	var req models.AddHinhAnhTourRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}
	result, err := s.z.AddHinhAnhTour(context.Background(), db.AddHinhAnhTourParams{
		TourID:       int32(id),
		DuongDan:     req.DuongDan,
		MoTa:         &req.MoTa,
		LaAnhChinh:   &req.LaAnhChinh,
		ThuTuHienThi: &req.ThuTuHienThi,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể thêm hình ảnh"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Thêm hình ảnh thành công",
		"data":    result,
	})
}

// DeleteHinhAnhTour godoc
// @Summary      Xóa hình ảnh tour
// @Description  Xóa một hình ảnh tour (cả trong database và Supabase storage)
// @Tags         departure
// @Produce      json
// @Param        id path int true "Hình ảnh tour ID"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/delete-hinh-anh/{id} [delete]
func (s *Server) DeleteHinhAnhTour(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Lấy thông tin hình ảnh trước khi xóa để có duong_dan và tour_id
	imageInfo, err := s.z.GetHinhAnhTourByID(ctx, int32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy hình ảnh"})
		return
	}

	// Xóa file từ Supabase storage nếu có duong_dan
	if imageInfo.DuongDan != "" {
		// Extract file path from URL if it's a full URL
		filePath := imageInfo.DuongDan

		// Nếu là full URL, extract path sau /storage/v1/object/public/images/ hoặc /storage/v1/object/images/
		// Ví dụ: https://xxx.supabase.co/storage/v1/object/public/images/tours/upload/file.jpg
		// -> tours/upload/file.jpg
		if strings.Contains(filePath, "/storage/v1/object/public/") {
			parts := strings.Split(filePath, "/storage/v1/object/public/")
			if len(parts) > 1 {
				// Remove bucket name if present
				pathParts := strings.SplitN(parts[1], "/", 2)
				if len(pathParts) > 1 {
					filePath = pathParts[1]
				} else {
					filePath = parts[1]
				}
			}
		} else if strings.Contains(filePath, "/storage/v1/object/") {
			parts := strings.Split(filePath, "/storage/v1/object/")
			if len(parts) > 1 {
				// Remove bucket name if present
				pathParts := strings.SplitN(parts[1], "/", 2)
				if len(pathParts) > 1 {
					filePath = pathParts[1]
				} else {
					filePath = parts[1]
				}
			}
		}

		// Xóa file từ Supabase storage
		if err := s.DeleteImage(ctx, filePath); err != nil {
			// Log error nhưng vẫn tiếp tục xóa record trong database
			fmt.Printf("Warning: Failed to delete image from Supabase storage: %v\n", err)
		}
	}

	// Xóa record trong database
	err = s.z.DeleteHinhAnhTour(ctx, db.DeleteHinhAnhTourParams{
		ID:     int32(id),
		TourID: imageInfo.TourID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa hình ảnh từ database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xóa hình ảnh thành công",
	})
}

// AddTourDestination godoc
// @Summary      Thêm điểm đến tour
// @Description  Thêm một điểm đến tour
// @Tags         departure
// @Accept       json
// @Produce      json
// @Param        id path int true "Tour ID"
// @Param        request body models.AddTourDestinationRequest true "Add data"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/add-tour-destination/{id} [post]
func (s *Server) AddTourDestination(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}
	var req models.AddTourDestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}
	err = s.z.AddTourDestination(context.Background(), db.AddTourDestinationParams{
		TourID:        int32(id),
		DiemDenID:     req.DiemDenID,
		ThuTuThamQuan: &req.ThuTuThamQuan,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể thêm điểm đến"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Thêm điểm đến thành công",
	})
}

// DeleteTourDestination godoc
// @Summary      Xóa điểm đến tour
// @Description  Xóa một điểm đến tour
// @Tags         departure
// @Produce      json
// @Param        tour_id path int true "Tour ID"
// @Param        diem_den_id path int true "Diem den ID"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /departure/delete-tour-destination/{tour_id}/{diem_den_id} [delete]
func (s *Server) DeleteTourDestination(c *gin.Context) {
	idStr := c.Param("tour_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}
	diemDenIDStr := c.Param("diem_den_id")
	diemDenID, err := strconv.Atoi(diemDenIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Diem den ID không hợp lệ"})
		return
	}
	err = s.z.DeleteTourDestination(context.Background(), db.DeleteTourDestinationParams{
		TourID:    int32(id),
		DiemDenID: int32(diemDenID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa điểm đến"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Xóa điểm đến thành công",
	})
}
