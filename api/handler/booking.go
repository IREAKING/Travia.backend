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

// HoldSeat godoc
// @Summary Hold seat
// @Description Hold seat
// @Tags Booking
// @Accept json
// @Produce json
// @Param khoi_hanh_id path int true "Khoi Hanh ID"
// @Param so_nguoi_lon path int true "So Nguoi Lon"
// @Param so_tre_em path int true "So Tre Em"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /booking/hold-seat/{khoi_hanh_id}/{so_nguoi_lon}/{so_tre_em} [post]
func (s *Server) HoldSeat(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	khoi_hanh_id, err := strconv.Atoi(c.Param("khoi_hanh_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid khoi_hanh_id"})
		return
	}

	so_nguoi_lon, err := strconv.Atoi(c.Param("so_nguoi_lon"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid so_nguoi_lon"})
		return
	}
	so_tre_em, err := strconv.Atoi(c.Param("so_tre_em"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid so_tre_em"})
		return
	}

	// Validation: Kiểm tra số lượng người hợp lệ
	if so_nguoi_lon < 0 || so_tre_em < 0 || (so_nguoi_lon+so_tre_em) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Số lượng người lớn và trẻ em phải không âm và tổng số người phải lớn hơn 0",
		})
		return
	}

	err = s.z.HoldSeat(ctx, db.HoldSeatParams{
		KhoiHanhID: int32(khoi_hanh_id),
		SoNguoiLon: int32(so_nguoi_lon),
		SoTreEm:    int32(so_tre_em),
	})
	if err != nil {
		// Phân loại lỗi để trả về status code phù hợp
		errMsg := err.Error()
		if strings.Contains(errMsg, "không tồn tại") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Không tìm thấy lịch khởi hành",
				"details": errMsg,
			})
		} else if strings.Contains(errMsg, "Không đủ chỗ") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Không đủ chỗ trống",
				"details": errMsg,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to hold seat",
				"details": errMsg,
			})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Seat held successfully"})
}

// CreateBooking godoc
// @Summary Create booking
// @Description Create booking
// @Tags Booking
// @Accept json
// @Produce json
// @Param booking body db.CreateBookingParams true "Booking"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /booking/create [post]
func (s *Server) CreateBooking(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

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
	var req db.CreateBookingParams
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation: Kiểm tra số lượng người
	if req.SoNguoiLon <= 0 && req.SoTreEm <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số lượng người phải lớn hơn 0"})
		return
	}

	// Validation: Kiểm tra departure tồn tại và hợp lệ trước
	departure, err := s.z.GetDepartureByID(ctx, req.KhoiHanhID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy lịch khởi hành"})
		fmt.Printf("Error getting departure: %v\n", err)
		return
	}

	// Kiểm tra trạng thái departure
	if !departure.TrangThai.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lịch khởi hành không hợp lệ"})
		fmt.Printf("Error getting departure: %v\n", err)
		return
	}
	trangThai := departure.TrangThai.TrangThaiKhoiHanh

	// Kiểm tra trạng thái hợp lệ: len_lich, xac_nhan, hoặc con_cho
	// Không chấp nhận het_cho vì đã hết chỗ
	isValidStatus := trangThai == db.TrangThaiKhoiHanhLenLich ||
		trangThai == db.TrangThaiKhoiHanhXacNhan ||
		trangThai == db.TrangThaiKhoiHanhConCho

	if !isValidStatus {
		if trangThai == db.TrangThaiKhoiHanhHetCho {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Lịch khởi hành đã hết chỗ",
				"trang_thai": string(trangThai),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "Lịch khởi hành không ở trạng thái hợp lệ để đặt tour",
				"trang_thai":     string(trangThai),
				"valid_statuses": []string{"len_lich", "xac_nhan", "con_cho"},
			})
		}
		return
	}

	// Validation: Kiểm tra số chỗ còn lại trước khi tạo booking
	totalPeople := req.SoNguoiLon + req.SoTreEm
	availability, err := s.z.CheckDepartureAvailability(ctx, db.CheckDepartureAvailabilityParams{
		ID:      req.KhoiHanhID,
		SucChua: int32(totalPeople),
	})
	if err != nil {
		// Lỗi có thể là "no rows" nếu departure không thỏa mãn điều kiện
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Không thể kiểm tra số chỗ. Lịch khởi hành có thể không hợp lệ hoặc đã hết chỗ",
		})
		fmt.Printf("Error checking availability: %v\n", err)
		return
	}
	if !availability.ConCho {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        "Không đủ chỗ trống",
			"so_cho_trong": availability.SoChoTrong,
		})
		return
	}

	_, err = s.z.CreateBooking(ctx, db.CreateBookingParams{
		NguoiDungID:         userUUID,
		KhoiHanhID:          req.KhoiHanhID,
		SoNguoiLon:          req.SoNguoiLon,
		SoTreEm:             req.SoTreEm,
		PhuongThucThanhToan: req.PhuongThucThanhToan,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking", "details": err.Error()})
		return
	}

	// CreateBookingRow is empty (sqlc issue with composite types), so we need to get the booking
	// Get the most recent booking for this user and departure
	bookings, err := s.z.GetBookingsByUserId(ctx, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve created booking", "details": err.Error()})
		return
	}

	// Find the booking that matches this departure (should be the most recent one)
	var createdBooking *db.GetBookingsByUserIdRow
	for i := range bookings {
		if bookings[i].KhoiHanhID == req.KhoiHanhID {
			// Take the first match (should be the most recent)
			createdBooking = &bookings[i]
			break
		}
	}

	if createdBooking == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find created booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking created successfully", "booking": createdBooking})
}

// AddPassengers godoc
// @Summary Add passengers
// @Description Add passengers
// @Tags Booking
// @Accept json
// @Produce json
// @Param passengers body []models.AddPassengersParams true "Passengers"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /booking/add-passengers [post]
func (s *Server) AddPassengers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Debug: Log incoming request
	fmt.Printf("AddPassengers called - Method: %s, Path: %s\n", c.Request.Method, c.Request.URL.Path)

	var req []models.AddPassengersParams
	err := c.ShouldBindJSON(&req)
	if err != nil {
		fmt.Printf("AddPassengers - Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("AddPassengers - Received %d passengers\n", len(req))

	if len(req) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Danh sách hành khách không được để trống"})
		return
	}

	// Validation: Kiểm tra tất cả passengers thuộc cùng một booking
	bookingID := req[0].DatChoID
	for _, p := range req {
		if p.DatChoID != bookingID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tất cả hành khách phải thuộc cùng một booking"})
			return
		}
	}

	// Validation: Kiểm tra booking tồn tại và lấy thông tin
	booking, err := s.z.GetBookingById(ctx, int32(bookingID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Validation: Kiểm tra số lượng hành khách khớp với booking
	var soNguoiLon, soTreEm int32
	if booking.SoNguoiLon != nil {
		soNguoiLon = *booking.SoNguoiLon
	}
	if booking.SoTreEm != nil {
		soTreEm = *booking.SoTreEm
	}
	expectedTotal := soNguoiLon + soTreEm
	if len(req) != int(expectedTotal) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":               "Số lượng hành khách không khớp với booking",
			"expected_passengers": expectedTotal,
			"provided_passengers": len(req),
		})
		return
	}

	// Validation: Kiểm tra số lượng người lớn và trẻ em
	var nguoiLonCount, treEmCount int32
	for _, p := range req {
		if p.LoaiKhach != nil {
			if *p.LoaiKhach == "nguoi_lon" {
				nguoiLonCount++
			} else if *p.LoaiKhach == "tre_em" {
				treEmCount++
			}
		}
	}

	if nguoiLonCount != soNguoiLon || treEmCount != soTreEm {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Số lượng người lớn và trẻ em không khớp với booking",
			"expected": gin.H{
				"so_nguoi_lon": soNguoiLon,
				"so_tre_em":    soTreEm,
			},
			"provided": gin.H{
				"so_nguoi_lon": nguoiLonCount,
				"so_tre_em":    treEmCount,
			},
		})
		return
	}
	var dbReq []db.AddPassengersParams
	for _, p := range req {
		dbReq = append(dbReq, db.AddPassengersParams{
			DatChoID:         int32(bookingID),
			HoTen:            p.HoTen,
			NgaySinh:         pgtype.Date{Time: utils.StringToDate(p.NgaySinh), Valid: true},
			LoaiKhach:        p.LoaiKhach,
			GioiTinh:         p.GioiTinh,
			SoGiayToTuyThanh: p.SoGiayToTuyThanh,
			QuocTich:         p.QuocTich,
			GhiChu:           p.GhiChu,
		})
	}

	_, err = s.z.AddPassengers(ctx, dbReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add passengers", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Passengers added successfully"})
}

// Get my bookings
// @Summary Get my bookings
// @Description Get my bookings with pagination
// @Tags Booking
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param trang_thai_dat_cho query string false "Trạng thái đặt chỗ"
// @Param trang_thai_khoi_hanh query string false "Trạng thái khởi hành"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /booking/my-bookings [get]
func (s *Server) GetMyBookings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var trangThaiDatCho *string
	var trangThaiKhoiHanh *string

	if c.Query("trang_thai_dat_cho") != "" {
		status := c.Query("trang_thai_dat_cho")
		trangThaiDatCho = &status
	}
	if c.Query("trang_thai_khoi_hanh") != "" {
		status := c.Query("trang_thai_khoi_hanh")
		// Nếu có dấu phẩy, đó là nhiều giá trị - giữ nguyên để SQL xử lý
		trangThaiKhoiHanh = &status
	}

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

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get bookings with pagination
	params := db.GetBookingsByUserParams{
		NguoiDungID: userUUID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	}
	if trangThaiDatCho != nil {
		params.Column4 = *trangThaiDatCho
	}
	if trangThaiKhoiHanh != nil {
		params.Column5 = *trangThaiKhoiHanh
	}
	bookings, err := s.z.GetBookingsByUser(ctx, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bookings", "message": err.Error()})
		return
	}

	// Get total count with same filters
	countParams := db.CountBookingsByUserParams{
		NguoiDungID: userUUID,
	}
	if trangThaiDatCho != nil {
		countParams.Column2 = *trangThaiDatCho
	}
	if trangThaiKhoiHanh != nil {
		countParams.Column3 = *trangThaiKhoiHanh
	}
	totalCount, err := s.z.CountBookingsByUser(ctx, countParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Bookings fetched successfully",
		"data":     bookings,
		"total":    totalCount,
		"limit":    limit,
		"offset":   offset,
		"has_more": (offset + limit) < int(totalCount),
	})
}

// GetBookingById godoc
// @Summary Get booking by ID
// @Description Get booking details by ID
// @Tags Booking
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /booking/{id} [get]
func (s *Server) GetBookingById(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	booking, err := s.z.GetBookingById(ctx, int32(bookingID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking fetched successfully",
		"data":    booking,
	})
}
