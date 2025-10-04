package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "travia.backend/db/sqlc"
)

// lấy tổng quan admin
// @summary Lấy tổng quan admin
// @description Lấy tổng quan admin
// @tags admin
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /admin/getAdminSummary [get]
func (s *Server) GetAdminSummary(c *gin.Context) {
	summary, err := s.z.GetAdminSummary(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy tổng quan admin thành công",
		"data":    summary,
	})
}

// lấy doanh thu theo tháng
// @summary Lấy doanh thu theo tháng
// @description Lấy doanh thu theo tháng
// @tags admin
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /admin/getRevenueByMonth [get]
func (s *Server) GetRevenueByMonth(c *gin.Context) {
	_year := c.Query("year")
	year, err := strconv.ParseInt(_year, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Năm không hợp lệ",
		})
		return
	}
	revenue, err := s.z.GetRevenueByMonth(context.Background(), pgtype.Timestamp{Time: time.Unix(year, 0)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy doanh thu theo tháng thành công",
		"data":    revenue,
	})
}

// lấy doanh thu theo năm
// @summary Lấy doanh thu theo năm
// @description Lấy doanh thu theo năm
// @tags admin
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /admin/getRevenueByYear [get]
func (s *Server) GetRevenueByYear(c *gin.Context) {
	_year := c.Query("year")
	year, err := strconv.ParseInt(_year, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Năm không hợp lệ",
		})
		return
	}
	revenue, err := s.z.GetRevenueByYear(context.Background(), pgtype.Timestamp{Time: time.Unix(year, 0)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy doanh thu theo năm thành công",
		"data":    revenue,
	})
}

// lấy đơn đặt chỗ theo trạng thái
// @summary Lấy đơn đặt chỗ theo trạng thái
// @description Lấy đơn đặt chỗ theo trạng thái
// @tags admin
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /admin/getBookingsByStatus [get]
func (s *Server) GetBookingsByStatus(c *gin.Context) {
	bookings, err := s.z.GetBookingsByStatus(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy đơn đặt chỗ theo trạng thái thành công",
		"data":    bookings,
	})
}

// lấy tour theo số đơn đặt chỗ
// @summary Lấy tour theo số đơn đặt chỗ
// @description Lấy tour theo số đơn đặt chỗ
// @tags admin
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /admin/getTopToursByBookings [get]
func (s *Server) GetTopToursByBookings(c *gin.Context) {
	_limit := c.DefaultQuery("limit", "10")
	limit, err := strconv.ParseInt(_limit, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Limit không hợp lệ",
		})
		return
	}
	bookings, err := s.z.GetTopToursByBookings(context.Background(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy tour theo số đơn đặt chỗ thành công",
		"data":    bookings,
	})
}

// @summary Lấy người dùng mới theo tháng
// @description Thống kê số người dùng đăng ký mới theo tháng trong năm
// @tags admin
// @accept json
// @produce json
// @param year query int true "Năm"
// @success 200 {object} gin.H
// @failure 400 {object} gin.H
// @router /admin/getNewUsersByMonth [get]
func (s *Server) GetNewUsersByMonth(c *gin.Context) {
	_year := c.Query("year")
	year, err := strconv.Atoi(_year)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Năm không hợp lệ"})
		return
	}
	data, err := s.z.GetNewUsersByMonth(context.Background(), int32(year))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}

// @summary Lấy đặt chỗ theo tháng
// @description Thống kê số lượng đặt chỗ theo tháng trong năm
// @tags admin
// @accept json
// @produce json
// @param year query int true "Năm"
// @success 200 {object} gin.H
// @router /admin/getBookingsByMonth [get]
func (s *Server) GetBookingsByMonth(c *gin.Context) {
	_year := c.Query("year")
	year, err := strconv.Atoi(_year)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Năm không hợp lệ"})
		return
	}
	data, err := s.z.GetBookingsByMonth(context.Background(), int32(year))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}

// @summary Top nhà cung cấp
// @description Lấy danh sách nhà cung cấp hàng đầu theo số tour và booking
// @tags admin
// @param limit query int false "Giới hạn kết quả" default(10)
// @success 200 {object} gin.H
// @router /admin/getTopSuppliers [get]
func (s *Server) GetTopSuppliers(c *gin.Context) {
	_limit := c.DefaultQuery("limit", "10")
	limit, _ := strconv.ParseInt(_limit, 10, 32)
	data, err := s.z.GetTopSuppliers(context.Background(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}

// @summary Doanh thu theo nhà cung cấp
// @description Thống kê doanh thu của từng nhà cung cấp
// @tags admin
// @param limit query int false "Giới hạn" default(10)
// @success 200 {object} gin.H
// @router /admin/getRevenueBySupplier [get]
func (s *Server) GetRevenueBySupplier(c *gin.Context) {
	_limit := c.DefaultQuery("limit", "10")
	limit, _ := strconv.ParseInt(_limit, 10, 32)
	data, err := s.z.GetRevenueBySupplier(context.Background(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}

// @summary Tour theo danh mục
// @description Thống kê số lượng tour và booking theo danh mục
// @tags admin
// @success 200 {object} gin.H
// @router /admin/getToursByCategory [get]
func (s *Server) GetToursByCategory(c *gin.Context) {
	data, err := s.z.GetToursByCategory(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}

// @summary Lịch khởi hành sắp tới
// @description Danh sách chuyến khởi hành sắp tới với thông tin chỗ trống
// @tags admin
// @param limit query int false "Giới hạn" default(20)
// @success 200 {object} gin.H
// @router /admin/getUpcomingDepartures [get]
func (s *Server) GetUpcomingDepartures(c *gin.Context) {
	_limit := c.DefaultQuery("limit", "20")
	limit, _ := strconv.ParseInt(_limit, 10, 32)
	data, err := s.z.GetUpcomingDepartures(context.Background(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}

// @summary Top khách hàng
// @description Khách hàng chi tiêu nhiều nhất
// @tags admin
// @param limit query int false "Giới hạn" default(10)
// @success 200 {object} gin.H
// @router /admin/getTopCustomers [get]
func (s *Server) GetTopCustomers(c *gin.Context) {
	_limit := c.DefaultQuery("limit", "10")
	limit, _ := strconv.ParseInt(_limit, 10, 32)
	data, err := s.z.GetTopCustomers(context.Background(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}

// @summary Thống kê đánh giá theo tour
// @description Chi tiết phân bố rating cho các tour
// @tags admin
// @param limit query int false "Giới hạn" default(10)
// @success 200 {object} gin.H
// @router /admin/getReviewStatsByTour [get]
func (s *Server) GetReviewStatsByTour(c *gin.Context) {
	_limit := c.DefaultQuery("limit", "10")
	limit, _ := strconv.ParseInt(_limit, 10, 32)
	data, err := s.z.GetReviewStatsByTour(context.Background(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}

// @summary Doanh thu theo khoảng thời gian
// @description Tổng doanh thu trong khoảng thời gian chỉ định
// @tags admin
// @param from query string true "Ngày bắt đầu (YYYY-MM-DD)"
// @param to query string true "Ngày kết thúc (YYYY-MM-DD)"
// @success 200 {object} gin.H
// @router /admin/getRevenueByDateRange [get]
func (s *Server) GetRevenueByDateRange(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Ngày bắt đầu không hợp lệ"})
		return
	}
	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Ngày kết thúc không hợp lệ"})
		return
	}
	params := db.GetRevenueByDateRangeParams{
		NgayThanhToan:   pgtype.Timestamp{Time: fromDate, Valid: true},
		NgayThanhToan_2: pgtype.Timestamp{Time: toDate, Valid: true},
	}
	data, err := s.z.GetRevenueByDateRange(context.Background(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}

// @summary Tăng trưởng người dùng
// @description Số người dùng mới theo ngày trong khoảng thời gian
// @tags admin
// @param from query string true "Ngày bắt đầu (YYYY-MM-DD)"
// @param to query string true "Ngày kết thúc (YYYY-MM-DD)"
// @success 200 {object} gin.H
// @router /admin/getUserGrowth [get]
func (s *Server) GetUserGrowth(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Ngày bắt đầu không hợp lệ"})
		return
	}
	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Ngày kết thúc không hợp lệ"})
		return
	}
	params := db.GetUserGrowthParams{
		NgayTao:   pgtype.Timestamp{Time: fromDate, Valid: true},
		NgayTao_2: pgtype.Timestamp{Time: toDate, Valid: true},
	}
	data, err := s.z.GetUserGrowth(context.Background(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": data})
}
