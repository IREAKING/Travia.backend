package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
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
