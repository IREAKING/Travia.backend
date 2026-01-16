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
			totalPeople := so_nguoi_lon + so_tre_em
			availability, availErr := s.z.CheckDepartureAvailability(ctx, db.CheckDepartureAvailabilityParams{
				ID:      int32(khoi_hanh_id),
				SucChua: int32(totalPeople),
			})
			if availErr == nil {
				c.JSON(http.StatusConflict, gin.H{
					"error":          "Không đủ chỗ trống",
					"details":        errMsg,
					"so_cho_trong":   availability.SoChoTrong,
					"so_cho_yeu_cau": totalPeople,
				})
			} else {
				c.JSON(http.StatusConflict, gin.H{
					"error":   "Không đủ chỗ trống",
					"details": errMsg,
				})
			}
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

// CancelBooking godoc
// @Summary Hủy đặt chỗ
// @Description Hủy đặt chỗ và tính số tiền hoàn lại theo chính sách hoàn tiền
// @Tags Booking
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Security BearerAuth
// @Router /booking/{id}/cancel [put]
func (s *Server) CancelBooking(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Kiểm tra authentication
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

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	// Kiểm tra booking có thuộc về user này không
	booking, err := s.z.GetBookingById(ctx, int32(bookingID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Kiểm tra quyền sở hữu
	if booking.NguoiDungID.String() != userUUID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền hủy booking này"})
		return
	}

	// Hủy booking và tính số tiền hoàn lại
	refundInfo, err := s.z.CancelBooking(ctx, int32(bookingID))
	if err != nil {
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "đã bị hủy") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Booking đã bị hủy trước đó"})
		} else if strings.Contains(errorMsg, "đã hoàn thành") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể hủy booking đã hoàn thành"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể hủy booking", "details": errorMsg})
		}
		return
	}

	// Convert pgtype.Numeric to float64
	var soTienHoan float64
	if refundInfo.SoTienHoan.Valid {
		floatVal, _ := refundInfo.SoTienHoan.Float64Value()
		if floatVal.Valid {
			soTienHoan = floatVal.Float64
		}
	}

	var phanTramHoan float64
	if refundInfo.PhanTramHoan.Valid {
		floatVal, _ := refundInfo.PhanTramHoan.Float64Value()
		if floatVal.Valid {
			phanTramHoan = floatVal.Float64
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hủy booking thành công",
		"data": gin.H{
			"so_tien_hoan":            soTienHoan,
			"phan_tram_hoan":          phanTramHoan,
			"so_ngay_truoc_khoi_hanh": refundInfo.SoNgayTruocKhoiHanh,
			"ly_do":                   refundInfo.LyDo,
		},
	})
}

// DeleteBooking godoc
// @Summary Xóa đặt chỗ
// @Description Xóa đặt chỗ và tất cả thông tin liên quan
// @Tags Booking
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Security BearerAuth
// @Router /booking/{id} [delete]
func (s *Server) DeleteBooking(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	err = s.z.DeleteBooking(ctx, int32(bookingID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete booking", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking deleted successfully"})
}

// DeleteBookings godoc
// @Summary Xóa nhiều đặt chỗ
// @Description Xóa nhiều đặt chỗ
// @Tags Booking
// @Accept json
// @Produce json
// @Param ids query []int true "Booking IDs"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Security BearerAuth
// @Router /booking/delete-bookings [delete]
func (s *Server) DeleteBookings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	idsStr := c.Query("ids")
	ids := strings.Split(idsStr, ",")
	idsInt := make([]int32, len(ids))
	for i, id := range ids {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking IDs"})
			return
		}
		idsInt[i] = int32(idInt)
	}
	err := s.z.DeleteBookings(ctx, idsInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bookings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bookings deleted successfully"})
}

// CalculateRefundAmount godoc
// @Summary Tính số tiền hoàn lại
// @Description Tính số tiền hoàn lại dựa trên chính sách hoàn tiền (không hủy booking)
// @Tags Booking
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Security BearerAuth
// @Router /booking/{id}/calculate-refund [get]
func (s *Server) CalculateRefundAmount(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Kiểm tra authentication
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

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	// Kiểm tra booking có thuộc về user này không
	booking, err := s.z.GetBookingById(ctx, int32(bookingID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Kiểm tra quyền sở hữu
	if booking.NguoiDungID.String() != userUUID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền xem booking này"})
		return
	}

	// Tính số tiền hoàn lại
	refundInfo, err := s.z.CalculateRefundAmount(ctx, int32(bookingID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tính số tiền hoàn lại", "details": err.Error()})
		return
	}

	// Convert pgtype.Numeric to float64
	var tongTien float64
	if refundInfo.TongTien.Valid {
		floatVal, _ := refundInfo.TongTien.Float64Value()
		if floatVal.Valid {
			tongTien = floatVal.Float64
		}
	}

	var soTienHoan float64
	if refundInfo.SoTienHoan.Valid {
		floatVal, _ := refundInfo.SoTienHoan.Float64Value()
		if floatVal.Valid {
			soTienHoan = floatVal.Float64
		}
	}

	var phanTramHoan float64
	if refundInfo.PhanTramHoan.Valid {
		floatVal, _ := refundInfo.PhanTramHoan.Float64Value()
		if floatVal.Valid {
			phanTramHoan = floatVal.Float64
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tính số tiền hoàn lại thành công",
		"data": gin.H{
			"tong_tien":               tongTien,
			"so_tien_hoan":            soTienHoan,
			"phan_tram_hoan":          phanTramHoan,
			"so_ngay_truoc_khoi_hanh": refundInfo.SoNgayTruocKhoiHanh,
			"ly_do":                   refundInfo.LyDo,
		},
	})
}

// ===========================================
// QUẢN LÝ HOÀN TIỀN (REFUND MANAGEMENT)
// ===========================================

// GetAllRefunds godoc
// @Summary Lấy tất cả refund cho admin
// @Description Lấy danh sách tất cả refund trong hệ thống với filter
// @Tags Booking
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param supplier_id query string false "Supplier ID (UUID)"
// @Param search query string false "Search keyword"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Security BearerAuth
// @Router /admin/refunds [get]
func (s *Server) GetAllRefunds(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Kiểm tra authentication và role admin
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

	// Kiểm tra role admin
	if jwtClaims.Vaitro != "quan_tri" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chỉ admin mới có quyền xem tất cả refund"})
		return
	}

	// Parse query parameters
	var startDate, endDate pgtype.Timestamp
	var supplierID pgtype.UUID
	searchKeyword := c.DefaultQuery("search", "")

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	if supplierIDStr := c.Query("supplier_id"); supplierIDStr != "" {
		if err := supplierID.Scan(supplierIDStr); err == nil {
			supplierID.Valid = true
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Lấy danh sách refund
	refunds, err := s.z.GetAllRefunds(ctx, db.GetAllRefundsParams{
		Column1: startDate,
		Column2: endDate,
		Column3: supplierID,
		Column4: searchKeyword,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách refund", "details": err.Error()})
		return
	}

	// Convert refunds to response format
	var refundList []gin.H
	for _, refund := range refunds {
		var tongTien, soTienHoan, phanTramHoan float64

		if refund.TongTien.Valid {
			floatVal, _ := refund.TongTien.Float64Value()
			if floatVal.Valid {
				tongTien = floatVal.Float64
			}
		}

		if refund.SoTienHoan.Valid {
			floatVal, _ := refund.SoTienHoan.Float64Value()
			if floatVal.Valid {
				soTienHoan = floatVal.Float64
			}
		}

		if refund.PhanTramHoan.Valid {
			floatVal, _ := refund.PhanTramHoan.Float64Value()
			if floatVal.Valid {
				phanTramHoan = floatVal.Float64
			}
		}

		refundList = append(refundList, gin.H{
			"booking_id":              refund.BookingID,
			"ngay_dat":                refund.NgayDat,
			"ngay_huy":                refund.NgayHuy,
			"tong_tien":               tongTien,
			"don_vi_tien_te":          refund.DonViTienTe,
			"so_nguoi_lon":            refund.SoNguoiLon,
			"so_tre_em":               refund.SoTreEm,
			"phuong_thuc_thanh_toan":  refund.PhuongThucThanhToan,
			"trang_thai":              refund.TrangThai,
			"customer_id":             refund.CustomerID,
			"customer_name":           refund.CustomerName,
			"customer_email":          refund.CustomerEmail,
			"customer_phone":          refund.CustomerPhone,
			"tour_id":                 refund.TourID,
			"tour_title":              refund.TourTitle,
			"supplier_id":             refund.SupplierID,
			"supplier_name":           refund.SupplierName,
			"departure_id":            refund.DepartureID,
			"ngay_khoi_hanh":          refund.NgayKhoiHanh,
			"ngay_ket_thuc":           refund.NgayKetThuc,
			"so_ngay_truoc_khoi_hanh": refund.SoNgayTruocKhoiHanh,
			"phan_tram_hoan":          phanTramHoan,
			"so_tien_hoan":            soTienHoan,
			"ly_do":                   refund.LyDo,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh sách refund thành công",
		"data":    refundList,
		"page":    page,
		"limit":   limit,
	})
}

// GetSupplierRefunds godoc
// @Summary Lấy refund cho supplier
// @Description Lấy danh sách refund cho các tour của supplier
// @Tags Booking
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param search query string false "Search keyword"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Security BearerAuth
// @Router /supplier/refunds [get]
func (s *Server) GetSupplierRefunds(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Kiểm tra authentication và role supplier
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

	// Kiểm tra role supplier
	if jwtClaims.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chỉ supplier mới có quyền xem refund"})
		return
	}

	supplierID := jwtClaims.Id

	// Parse query parameters
	var startDate, endDate pgtype.Timestamp
	searchKeyword := c.DefaultQuery("search", "")

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Lấy danh sách refund
	refunds, err := s.z.GetSupplierRefunds(ctx, db.GetSupplierRefundsParams{
		Column1: supplierID,
		Column2: startDate,
		Column3: endDate,
		Column4: searchKeyword,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách refund", "details": err.Error()})
		return
	}

	// Convert refunds to response format
	var refundList []gin.H
	for _, refund := range refunds {
		var tongTien, soTienHoan, phanTramHoan float64

		if refund.TongTien.Valid {
			floatVal, _ := refund.TongTien.Float64Value()
			if floatVal.Valid {
				tongTien = floatVal.Float64
			}
		}

		if refund.SoTienHoan.Valid {
			floatVal, _ := refund.SoTienHoan.Float64Value()
			if floatVal.Valid {
				soTienHoan = floatVal.Float64
			}
		}

		if refund.PhanTramHoan.Valid {
			floatVal, _ := refund.PhanTramHoan.Float64Value()
			if floatVal.Valid {
				phanTramHoan = floatVal.Float64
			}
		}

		refundList = append(refundList, gin.H{
			"booking_id":              refund.BookingID,
			"ngay_dat":                refund.NgayDat,
			"ngay_huy":                refund.NgayHuy,
			"tong_tien":               tongTien,
			"don_vi_tien_te":          refund.DonViTienTe,
			"so_nguoi_lon":            refund.SoNguoiLon,
			"so_tre_em":               refund.SoTreEm,
			"phuong_thuc_thanh_toan":  refund.PhuongThucThanhToan,
			"trang_thai":              refund.TrangThai,
			"customer_id":             refund.CustomerID,
			"customer_name":           refund.CustomerName,
			"customer_email":          refund.CustomerEmail,
			"customer_phone":          refund.CustomerPhone,
			"tour_id":                 refund.TourID,
			"tour_title":              refund.TourTitle,
			"departure_id":            refund.DepartureID,
			"ngay_khoi_hanh":          refund.NgayKhoiHanh,
			"ngay_ket_thuc":           refund.NgayKetThuc,
			"so_ngay_truoc_khoi_hanh": refund.SoNgayTruocKhoiHanh,
			"phan_tram_hoan":          phanTramHoan,
			"so_tien_hoan":            soTienHoan,
			"ly_do":                   refund.LyDo,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh sách refund thành công",
		"data":    refundList,
		"page":    page,
		"limit":   limit,
	})
}

// GetRefundStats godoc
// @Summary Thống kê refund cho admin
// @Description Lấy thống kê refund trong hệ thống
// @Tags Booking
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Security BearerAuth
// @Router /admin/refunds/stats [get]
func (s *Server) GetRefundStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Kiểm tra authentication và role admin
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

	// Kiểm tra role admin
	if jwtClaims.Vaitro != "quan_tri" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chỉ admin mới có quyền xem thống kê refund"})
		return
	}

	// Parse query parameters
	var startDate, endDate pgtype.Timestamp

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	// Lấy thống kê refund
	stats, err := s.z.GetRefundStats(ctx, db.GetRefundStatsParams{
		Column1: startDate,
		Column2: endDate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy thống kê refund", "details": err.Error()})
		return
	}

	// Convert stats to response format
	var tongTienGoc, tongTienHoan, tongTienPhat float64

	if stats.TongTienGoc.Valid {
		floatVal, _ := stats.TongTienGoc.Float64Value()
		if floatVal.Valid {
			tongTienGoc = floatVal.Float64
		}
	}

	if stats.TongTienHoan.Valid {
		floatVal, _ := stats.TongTienHoan.Float64Value()
		if floatVal.Valid {
			tongTienHoan = floatVal.Float64
		}
	}

	if stats.TongTienPhat.Valid {
		floatVal, _ := stats.TongTienPhat.Float64Value()
		if floatVal.Valid {
			tongTienPhat = floatVal.Float64
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy thống kê refund thành công",
		"data": gin.H{
			"tong_so_refund":   stats.TongSoRefund,
			"tong_tien_goc":    tongTienGoc,
			"tong_tien_hoan":   tongTienHoan,
			"tong_tien_phat":   tongTienPhat,
			"hoan_100_percent": stats.Hoan100Percent,
			"hoan_90_percent":  stats.Hoan90Percent,
			"hoan_70_percent":  stats.Hoan70Percent,
			"hoan_50_percent":  stats.Hoan50Percent,
			"khong_hoan":       stats.KhongHoan,
		},
	})
}

// GetSupplierRefundStats godoc
// @Summary Thống kê refund cho supplier
// @Description Lấy thống kê refund cho các tour của supplier
// @Tags Booking
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Security BearerAuth
// @Router /supplier/refunds/stats [get]
func (s *Server) GetSupplierRefundStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Kiểm tra authentication và role supplier
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

	// Kiểm tra role supplier
	if jwtClaims.Vaitro != "nha_cung_cap" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chỉ supplier mới có quyền xem thống kê refund"})
		return
	}

	supplierID := jwtClaims.Id

	// Parse query parameters
	var startDate, endDate pgtype.Timestamp

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	// Lấy thống kê refund
	stats, err := s.z.GetSupplierRefundStats(ctx, db.GetSupplierRefundStatsParams{
		Column1: supplierID,
		Column2: startDate,
		Column3: endDate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy thống kê refund", "details": err.Error()})
		return
	}

	// Convert stats to response format
	var tongTienGoc, tongTienHoan, tongTienPhat float64

	if stats.TongTienGoc.Valid {
		floatVal, _ := stats.TongTienGoc.Float64Value()
		if floatVal.Valid {
			tongTienGoc = floatVal.Float64
		}
	}

	if stats.TongTienHoan.Valid {
		floatVal, _ := stats.TongTienHoan.Float64Value()
		if floatVal.Valid {
			tongTienHoan = floatVal.Float64
		}
	}

	if stats.TongTienPhat.Valid {
		floatVal, _ := stats.TongTienPhat.Float64Value()
		if floatVal.Valid {
			tongTienPhat = floatVal.Float64
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy thống kê refund thành công",
		"data": gin.H{
			"tong_so_refund":   stats.TongSoRefund,
			"tong_tien_goc":    tongTienGoc,
			"tong_tien_hoan":   tongTienHoan,
			"tong_tien_phat":   tongTienPhat,
			"hoan_100_percent": stats.Hoan100Percent,
			"hoan_90_percent":  stats.Hoan90Percent,
			"hoan_70_percent":  stats.Hoan70Percent,
			"hoan_50_percent":  stats.Hoan50Percent,
			"khong_hoan":       stats.KhongHoan,
		},
	})
}
