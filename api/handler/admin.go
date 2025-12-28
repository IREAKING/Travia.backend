package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	db "travia.backend/db/sqlc"
)

// GetDashboardOverview godoc
// @Summary Lấy tổng quan dashboard
// @Description Lấy tổng quan dashboard
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetDashboardOverviewRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getDashboardOverview [get]
func (s *Server) GetDashboardOverview(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	overview, err := s.z.GetDashboardOverview(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get dashboard overview",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Dashboard overview fetched successfully",
		"data":    overview,
	})
}

// GetDashboardOverviewWithComparison godoc
// @Summary Lấy tổng quan dashboard với so sánh tháng trước
// @Description Lấy tổng quan dashboard với so sánh tháng trước
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetDashboardOverviewWithComparisonRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getDashboardOverviewWithComparison [get]
func (s *Server) GetDashboardOverviewWithComparison(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	overview, err := s.z.GetDashboardOverviewWithComparison(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get dashboard overview with comparison",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Dashboard overview with comparison fetched successfully",
		"data":    overview,
	})
}

// GetUserStatsByRole godoc
// @Summary Lấy thống kê người dùng theo vai trò
// @Description Lấy thống kê người dùng theo vai trò
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetUserStatsByRoleRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getUserStatsByRole [get]
func (s *Server) GetUserStatsByRole(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	stats, err := s.z.GetUserStatsByRole(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user stats by role",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User stats by role fetched successfully",
		"data":    stats,
	})
}

// GetUserGrowthByMonth godoc
// @Summary Lấy tăng trưởng người dùng theo tháng (12 tháng gần nhất)
// @Description Lấy tăng trưởng người dùng theo tháng (12 tháng gần nhất)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetUserGrowthByMonthRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getUserGrowthByMonth [get]
func (s *Server) GetUserGrowthByMonth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	growth, err := s.z.GetUserGrowthByMonth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user growth by month",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User growth by month fetched successfully",
		"data":    growth,
	})
}

// GetUserGrowthByDay godoc
// @Summary Lấy tăng trưởng người dùng theo ngày (30 ngày gần nhất)
// @Description Lấy tăng trưởng người dùng theo ngày (30 ngày gần nhất)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetUserGrowthByDayRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getUserGrowthByDay [get]
func (s *Server) GetUserGrowthByDay(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	growth, err := s.z.GetUserGrowthByDay(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user growth by day",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User growth by day fetched successfully",
		"data":    growth,
	})
}

// GetNewUsersToday godoc
// @Summary Lấy số người dùng mới hôm nay
// @Description Lấy số người dùng mới hôm nay
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetNewUsersTodayRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getNewUsersToday [get]
func (s *Server) GetNewUsersToday(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	growth, err := s.z.GetNewUsersToday(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get new users today",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "New users today fetched successfully",
		"data":    growth,
	})
}

// GetTopActiveUsers godoc
// @Summary Lấy top người dùng hoạt động nhiều nhất (theo số booking)
// @Description Lấy top người dùng hoạt động nhiều nhất (theo số booking)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetTopActiveUsersRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getTopActiveUsers [get]
func (s *Server) GetTopActiveUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	users, err := s.z.GetTopActiveUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get top active users",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Top active users fetched successfully",
		"data":    users,
	})
}

// GetTopBookedTours godoc
// @Summary Lấy top tours được đặt nhiều nhất
// @Description Lấy top tours được đặt nhiều nhất
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetTopBookedToursRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getTopBookedTours [get]
func (s *Server) GetTopBookedTours(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	tours, err := s.z.GetTopBookedTours(ctx, int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get top booked tours",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Top booked tours fetched successfully",
		"data":    tours,
	})
}

// GetToursCreatedByMonth godoc
// @Summary Lấy số tour mới theo tháng
// @Description Lấy số tour mới theo tháng
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetToursCreatedByMonthRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getToursCreatedByMonth [get]
func (s *Server) GetToursCreatedByMonth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	months, err := s.z.GetToursCreatedByMonth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get tours created by month",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Tours created by month fetched successfully",
		"data":    months,
	})
}

// GetTourPriceDistribution godoc
// @Summary Lấy phân bố giá tour
// @Description Lấy phân bố giá tour
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetTourPriceDistributionRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getTourPriceDistribution [get]
func (s *Server) GetTourPriceDistribution(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	distribution, err := s.z.GetTourPriceDistribution(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get tour price distribution",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Tour price distribution fetched successfully",
		"data":    distribution,
	})
}

// GetRevenueByDay godoc
// @Summary Lấy doanh thu theo ngày (30 ngày gần nhất)
// @Description Lấy doanh thu theo ngày (30 ngày gần nhất)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetRevenueByDayRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getRevenueByDay [get]
func (s *Server) GetRevenueByDay(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	revenue, err := s.z.GetRevenueByDay(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get revenue by day",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Revenue by day fetched successfully",
		"data":    revenue,
	})
}

// GetRevenueByMonth godoc
// @Summary Lấy doanh thu theo tháng (12 tháng gần nhất)
// @Description Lấy doanh thu theo tháng (12 tháng gần nhất)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetRevenueByMonthRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getRevenueByMonth [get]
func (s *Server) GetRevenueByMonth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	revenue, err := s.z.GetRevenueByMonth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get revenue by month",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Revenue by month fetched successfully",
		"data":    revenue,
	})
}

// GetRevenueByYear godoc
// @Summary Lấy doanh thu theo năm
// @Description Lấy doanh thu theo năm
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetRevenueByYearRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getRevenueByYear [get]
func (s *Server) GetRevenueByYear(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	revenue, err := s.z.GetRevenueByYear(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get revenue by year",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Revenue by year fetched successfully",
		"data":    revenue,
	})
}

// GetBookingsByDayOfWeek godoc
// @Summary Lấy số booking theo ngày trong tuần
// @Description Lấy số booking theo ngày trong tuần
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetBookingsByDayOfWeekRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getBookingsByDayOfWeek [get]
func (s *Server) GetBookingsByDayOfWeek(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	bookings, err := s.z.GetBookingsByDayOfWeek(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get bookings by day of week",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Bookings by day of week fetched successfully",
		"data":    bookings,
	})
}

// GetRecentBookings godoc
// @Summary Lấy booking gần đây
// @Description Lấy booking gần đây
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetRecentBookingsRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getRecentBookings [get]
func (s *Server) GetRecentBookings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	bookings, err := s.z.GetRecentBookings(ctx, int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get recent bookings",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Recent bookings fetched successfully",
		"data":    bookings,
	})
}

// GetBookingStatsByStatus godoc
// @Summary Lấy thống kê booking theo trạng thái
// @Description Lấy thống kê booking theo trạng thái
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetBookingStatsByStatusRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getBookingsByStatus [get]
func (s *Server) GetBookingStatsByStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	stats, err := s.z.GetBookingStatsByStatus(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get booking stats by status",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Booking stats by status fetched successfully",
		"data":    stats,
	})
}

// GetTransactions godoc
// @Summary Lấy danh sách giao dịch thanh toán
// @Description Lấy danh sách tất cả giao dịch thanh toán với phân trang
// @Tags Admin
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng giao dịch mỗi trang (default: 20)"
// @Param offset query int false "Số lượng giao dịch bỏ qua (default: 0)"
// @Param status query string false "Lọc theo trạng thái (cho_thanh_toan, dang_cho_thanh_toan, thanh_cong, that_bai)"
// @Param gateway query string false "Lọc theo cổng thanh toán"
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/transactions [get]
func (s *Server) GetTransactions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	if offset < 0 {
		offset = 0
	}

	// Get status filter
	statusFilter := c.Query("status")
	gatewayFilter := c.Query("gateway")

	var totalCount int64
	var err error

	// Get all transactions (we'll filter in memory for now, or add proper SQL filters later)
	allTransactions, err := s.z.GetAllTransactions(ctx, db.GetAllTransactionsParams{
		Limit:  int32(limit * 10), // Get more to filter
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get transactions",
			"details": err.Error(),
		})
		return
	}

	// Filter by status if provided
	var filteredTransactions []db.GetAllTransactionsRow
	if statusFilter != "" {
		for _, tx := range allTransactions {
			if tx.TrangThai.Valid && string(tx.TrangThai.TrangThaiThanhToan) == statusFilter {
				filteredTransactions = append(filteredTransactions, tx)
			}
		}
		allTransactions = filteredTransactions[:min(len(filteredTransactions), limit)]
	} else if gatewayFilter != "" {
		for _, tx := range allTransactions {
			if tx.CongThanhToanID != nil && *tx.CongThanhToanID == gatewayFilter {
				filteredTransactions = append(filteredTransactions, tx)
			}
		}
		allTransactions = filteredTransactions[:min(len(filteredTransactions), limit)]
	} else {
		// Limit to requested limit
		if len(allTransactions) > limit {
			allTransactions = allTransactions[:limit]
		}
	}

	// Get total count
	totalCount, err = s.z.CountAllTransactions(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to count transactions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Transactions fetched successfully",
		"data":     allTransactions,
		"total":    totalCount,
		"limit":    limit,
		"offset":   offset,
		"has_more": (offset + limit) < int(totalCount),
	})
}
