package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "travia.backend/db/sqlc"
)

// GetSupplierOptions godoc
// @Summary Lấy danh sách nhà cung cấp
// @Description Lấy danh sách nhà cung cấp
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.SupplierOptionsRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/supplierOptions [get]
func (s *Server) GetSupplierOptions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	supplierOptions, err := s.z.SupplierOptions(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "lỗi khi lấy danh sách nhà cung cấp",
			"details": err.Error(),
		})
		return
	}
	// Convert pgtype.UUID to string for JSON response
	options := make([]gin.H, len(supplierOptions))
	for i, opt := range supplierOptions {
		options[i] = gin.H{
			"id":  opt.ID.String(),
			"ten": opt.Ten,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Supplier options fetched successfully",
		"data":    options,
	})
}

// GetDashboardOverview godoc
// @Summary Lấy tổng quan dashboard
// @Description Lấy tổng quan dashboard
// @Tags Admin
// @Accept json
// @Produce json
// @Param month query int false "Tháng (1-12)"
// @Param year query int false "Năm"
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

// GetDashboardOverview godoc
// @Summary Lấy tổng quan dashboard
// @Description Lấy tổng quan dashboard
// @Tags Admin
// @Accept json
// @Produce json
// @Param month query int false "Tháng (1-12)"
// @Param year query int false "Năm"
// @Success 200 {object} db.GetDashboardOverviewRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getDashboardOverviewByMonthAndYear [get]
func (s *Server) GetDashboardOverviewByMonthAndYear(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	overview, err := s.z.GetDashboardOverviewByMonthAndYear(ctx, db.GetDashboardOverviewByMonthAndYearParams{
		Nam:   int32(year),
		Thang: int32(month),
	})
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

// GetRevenueByDay godoc
// @Summary Lấy doanh thu theo ngày (30 ngày gần nhất)
// @Description Lấy doanh thu theo ngày (30 ngày gần nhất)
// @Tags Admin
// @Accept json
// @Produce json
// @Param year query int false "Năm"
// @Param month query int false "Tháng (1-12)"
// @Param nha_cung_cap_id query string false "ID Nhà cung cấp"
// @Success 200 {object} db.GetRevenueByDayRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getRevenueByDay [get]
func (s *Server) GetRevenueByDay(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	nhaCungCapIDStr := c.Query("nha_cung_cap_id")
	var nhaCungCapIDUUID pgtype.UUID
	if nhaCungCapIDStr != "" {
		if err := nhaCungCapIDUUID.Scan(nhaCungCapIDStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid nha_cung_cap_id format",
				"details": err.Error(),
			})
			return
		}
	}
	revenue, err := s.z.GetRevenueByDay(ctx, db.GetRevenueByDayParams{
		Nam:          int32(year),
		Thang:        int32(month),
		NhaCungCapID: nhaCungCapIDUUID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "lỗi khi lấy doanh thu theo ngày",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Revenue by day fetched successfully",
		"data":    revenue,
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

// GetTopDestination godoc
// @Summary Lấy top điểm đến phổ biến
// @Description Lấy top điểm đến phổ biến
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetTopDestinationsRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/getTopDestination [get]
func (s *Server) GetTopDestinations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	destination, err := s.z.GetTopDestinations(ctx, int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get top destination",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Top destination fetched successfully",
		"data":    destination,
	})
}

// AdminChartRevenueTrend godoc
// @Summary Lấy biểu đồ doanh thu theo ngày
// @Description Lấy biểu đồ doanh thu theo ngày
// @Tags Admin
// @Accept json
// @Produce json
// @Param year query int false "Năm"
// @Param month query int false "Tháng (1-12)"
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/chartRevenueTrend [get]
func (s *Server) AdminChartRevenueTrend(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	chart, err := s.z.AdminChartRevenueTrend(ctx, db.AdminChartRevenueTrendParams{
		Nam:   int32(year),
		Thang: int32(month),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get revenue trend chart",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Revenue trend chart fetched successfully",
		"data":    chart,
	})
}

// AdminChartCategoryDistribution godoc
// @Summary Lấy biểu đồ phân bố danh mục tour
// @Description Lấy biểu đồ phân bố danh mục tour
// @Tags Admin
// @Accept json
// @Produce json
// @Param year query int false "Năm"
// @Param month query int false "Tháng (1-12)"
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/chartCategoryDistribution [get]
func (s *Server) AdminChartCategoryDistribution(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	chart, err := s.z.AdminChartCategoryDistribution(ctx, db.AdminChartCategoryDistributionParams{
		Nam:   int32(year),
		Thang: int32(month),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get category distribution chart",
			"details": err.Error(),
		})
		return
	}
	// Convert pgtype.Numeric to string for JSON response
	result := make([]gin.H, len(chart))
	for i, item := range chart {
		var doanhThuStr string
		if item.TongDoanhThu.Valid {
			floatVal, _ := item.TongDoanhThu.Float64Value()
			if floatVal.Valid {
				doanhThuStr = fmt.Sprintf("%.2f", floatVal.Float64)
			} else {
				doanhThuStr = "0"
			}
		} else {
			doanhThuStr = "0"
		}
		result[i] = gin.H{
			"ten_danh_muc":   item.TenDanhMuc,
			"so_luong_dat":   item.SoLuongDat,
			"tong_doanh_thu": doanhThuStr,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Category distribution chart fetched successfully",
		"data":    result,
	})
}

// AdminChartTopSuppliers godoc
// @Summary Lấy biểu đồ top nhà cung cấp
// @Description Lấy biểu đồ top nhà cung cấp
// @Tags Admin
// @Accept json
// @Produce json
// @Param year query int false "Năm"
// @Param month query int false "Tháng (1-12)"
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/chartTopSuppliers [get]
func (s *Server) AdminChartTopSuppliers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	chart, err := s.z.AdminChartTopSuppliers(ctx, db.AdminChartTopSuppliersParams{
		Nam:   int32(year),
		Thang: int32(month),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get top suppliers chart",
			"details": err.Error(),
		})
		return
	}
	// Convert pgtype.Numeric to string for JSON response
	result := make([]gin.H, len(chart))
	for i, item := range chart {
		var doanhThuStr string
		if item.DoanhThuDatDuoc.Valid {
			// Use Float64Value to get the numeric value, then convert to string
			floatVal, _ := item.DoanhThuDatDuoc.Float64Value()
			if floatVal.Valid {
				doanhThuStr = fmt.Sprintf("%.2f", floatVal.Float64)
			} else {
				doanhThuStr = "0"
			}
		} else {
			doanhThuStr = "0"
		}
		result[i] = gin.H{
			"ten_nha_cung_cap":   item.TenNhaCungCap,
			"so_don_hang":        item.SoDonHang,
			"doanh_thu_dat_duoc": doanhThuStr,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Top suppliers chart fetched successfully",
		"data":    result,
	})
}

// AdminChartBookingStatusStats godoc
// @Summary Lấy biểu đồ trạng thái booking
// @Description Lấy biểu đồ trạng thái booking
// @Tags Admin
// @Accept json
// @Produce json
// @Param year query int false "Năm"
// @Param month query int false "Tháng (1-12)"
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/chartBookingStatusStats [get]
func (s *Server) AdminChartBookingStatusStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	chart, err := s.z.AdminChartBookingStatusStats(ctx, db.AdminChartBookingStatusStatsParams{
		Nam:   int32(year),
		Thang: int32(month),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get booking status stats chart",
			"details": err.Error(),
		})
		return
	}
	// Convert NullTrangThaiDatCho and pgtype.Numeric to strings for JSON response
	result := make([]gin.H, len(chart))
	for i, item := range chart {
		var trangThaiStr string
		if item.TrangThai.Valid {
			trangThaiStr = string(item.TrangThai.TrangThaiDatCho)
		} else {
			trangThaiStr = ""
		}

		var giaTriStr string
		if item.GiaTriUocTinh.Valid {
			floatVal, _ := item.GiaTriUocTinh.Float64Value()
			if floatVal.Valid {
				giaTriStr = fmt.Sprintf("%.2f", floatVal.Float64)
			} else {
				giaTriStr = "0"
			}
		} else {
			giaTriStr = "0"
		}

		result[i] = gin.H{
			"trang_thai":       trangThaiStr,
			"so_luong":         item.SoLuong,
			"gia_tri_uoc_tinh": giaTriStr,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Booking status stats chart fetched successfully",
		"data":    result,
	})
}

// =====================================Nhà cung cấp=====================================
// @Summary Lấy tất cả nhà cung cấp
// @Description Lấy tất cả nhà cung cấp
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} db.GetAllSuppliersRow
// @Param xac_thuc query bool false "Xác thực" default(true)
// @Param dang_hoat_dong query bool false "Hoạt động" default(false)
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/suppliers [get]
func (s *Server) GetAllSuppliers(c *gin.Context) {
	xacThucStr := c.Query("xac_thuc")
	dangHoatDongStr := c.Query("dang_hoat_dong")

	var xacThuc *bool
	if xacThucStr != "" {
		b := xacThucStr == "true"
		xacThuc = &b
	}

	var dangHoatDong *bool
	if dangHoatDongStr != "" {
		b := dangHoatDongStr == "true"
		dangHoatDong = &b
	}

	data, err := s.z.GetAllSuppliers(context.Background(), db.GetAllSuppliersParams{
		XacThuc:      xacThuc,
		DangHoatDong: dangHoatDong,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	// Convert pgtype.UUID to string for JSON response
	result := make([]gin.H, len(data))
	for i, item := range data {
		result[i] = gin.H{
			"id":                 item.ID.String(),
			"ten":                item.Ten,
			"dia_chi":            item.DiaChi,
			"website":            item.Website,
			"mo_ta":              item.MoTa,
			"logo":               item.Logo,
			"logo_url":           item.Logo,
			"nam_thanh_lap":      item.NamThanhLap,
			"thanh_pho":          item.ThanhPho,
			"quoc_gia":           item.QuocGia,
			"ma_so_thue":         item.MaSoThue,
			"so_nhan_vien":       item.SoNhanVien,
			"giay_to_kinh_doanh": item.GiayToKinhDoanh,
			"email":              item.Email,
			"ho_ten":             item.HoTen,
			"trang_thai":         "hoat_dong", // Default status
			"so_dien_thoai":      item.SoDienThoai,
			"ngay_tao":           item.NgayTao,
			"xac_thuc":           item.XacThuc,
			"dang_hoat_dong":     item.DangHoatDong,
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Suppliers fetched successfully", "data": result})
}

// Phê duyệt đối tác
// @Summary Phê duyệt đối tác
// @Description Phê duyệt đối tác
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/suppliers/approve/{id} [put]
func (s *Server) ApproveSupplier(c *gin.Context) {
	_id := c.Param("id")
	var id pgtype.UUID
	if err := id.Scan(_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}
	result, err := s.z.ApproveSupplier(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier approved successfully", "data": result})
}

// Từ chối đối tác
// @Summary Từ chối đối tác
// @Description Từ chối đối tác
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/suppliers/reject/{id} [put]
func (s *Server) RejectSupplier(c *gin.Context) {
	_id := c.Param("id")
	var id pgtype.UUID
	if err := id.Scan(_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}
	result, err := s.z.RejectSupplier(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier rejected successfully", "data": result})
}

// Xóa nhà cung cấp
// @Summary Xóa nhà cung cấp
// @Description Xóa nhà cung cấp
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/suppliers/soft-delete/{id} [delete]
func (s *Server) SoftDeleteSupplier(c *gin.Context) {
	_id := c.Param("id")
	var id pgtype.UUID
	if err := id.Scan(_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}
	err := s.z.SoftDeleteSupplier(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		log.Println("Error restoring supplier:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier soft deleted successfully"})
}

// Khôi phục nhà cung cấp
// @Summary Khôi phục nhà cung cấp
// @Description Khôi phục nhà cung cấp
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} db.NhaCungCap
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/suppliers/restore/{id} [put]
func (s *Server) RestoreSupplier(c *gin.Context) {
	_id := c.Param("id")
	var id pgtype.UUID
	if err := id.Scan(_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}
	data, err := s.z.RestoreSupplier(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier restored successfully", "data": data})
}

// Lấy nhà cung cấp theo ID
// @Summary Lấy nhà cung cấp theo ID
// @Description Lấy nhà cung cấp theo ID
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} db.NhaCungCap
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/suppliers/{id} [get]
func (s *Server) GetSupplierByID(c *gin.Context) {
	_id := c.Param("id")
	var id pgtype.UUID
	if err := id.Scan(_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}
	data, err := s.z.GetAdminSupplierByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Không tìm thấy nhà cung cấp với ID: " + _id,
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	// Convert pgtype.UUID and other types to JSON-friendly format
	result := gin.H{
		"id":                 data.ID.String(),
		"ten":                data.Ten,
		"dia_chi":            data.DiaChi,
		"website":            data.Website,
		"mo_ta":              data.MoTa,
		"logo":               data.Logo,
		"logo_url":           data.Logo,
		"nam_thanh_lap":      data.NamThanhLap,
		"thanh_pho":          data.ThanhPho,
		"quoc_gia":           data.QuocGia,
		"ma_so_thue":         data.MaSoThue,
		"so_nhan_vien":       data.SoNhanVien,
		"giay_to_kinh_doanh": data.GiayToKinhDoanh,
		"email":              data.Email,
		"ho_ten":             data.HoTen,
		"so_dien_thoai":      data.SoDienThoai,
		"ngay_tao":           data.NgayTao,
		"ngay_cap_nhat":      data.NgayCapNhat,
		"dang_hoat_dong":     data.DangHoatDong,
		"xac_thuc":           data.XacThuc,
		"trang_thai":         "hoat_dong", // Default status
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier fetched successfully", "data": result})
}

// =====================================Khách hàng=====================================
// GetTopActiveUsers godoc
// @Summary Lấy top người dùng hoạt động nhiều nhất (theo số booking)
// @Description Lấy top người dùng hoạt động nhiều nhất (theo số booking)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} db.GetTopActiveUsersRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/customers/getTopActiveUsers [get]
func (s *Server) GetTopActiveUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	defer cancel()
	users, err := s.z.GetTopActiveUsers(ctx, int32(limit))
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

// AdminCustomerGrowthMonthlyReport godoc
// @Summary Lấy báo cáo tăng trưởng khách hàng theo tháng
// @Description Lấy báo cáo tăng trưởng khách hàng theo tháng
// @Tags Admin
// @Accept json
// @Produce json
// @Param year query int false "Năm"
// @Success 200 {object} db.AdminCustomerGrowthMonthlyReportRow
// @Failure 500 {object} gin.H{}
// @Security ApiKeyAuth
// @Router /admin/customers/adminCustomerGrowthMonthlyReport [get]
func (s *Server) AdminCustomerGrowthMonthlyReport(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)
	defer cancel()
	report, err := s.z.AdminCustomerGrowthMonthlyReport(ctx, int32(year))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get admin customer growth monthly report",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin customer growth monthly report fetched successfully",
		"data":    report,
	})
}

// GetAdminBookingStatistics godoc
// @Summary Lấy thống kê đặt chỗ chi tiết cho admin
// @Description Lấy thống kê đặt chỗ với nhiều filter: thời gian, nhà cung cấp, trạng thái
// @Tags Admin
// @Accept json
// @Produce json
// @Param start_date query string false "Ngày bắt đầu (YYYY-MM-DD)"
// @Param end_date query string false "Ngày kết thúc (YYYY-MM-DD)"
// @Param supplier_id query string false "ID Nhà cung cấp (UUID)"
// @Param trang_thai query string false "Trạng thái booking (cho_xac_nhan, da_xac_nhan, da_thanh_toan, hoan_thanh, da_huy)"
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/bookings/statistics [get]
func (s *Server) GetAdminBookingStatistics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse query parameters
	var startDate, endDate pgtype.Timestamp
	var supplierID pgtype.UUID
	trangThai := c.Query("trang_thai")

	// Parse start_date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	// Parse end_date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	// Parse supplier_id
	if supplierIDStr := c.Query("supplier_id"); supplierIDStr != "" {
		if err := supplierID.Scan(supplierIDStr); err == nil {
			supplierID.Valid = true
		}
	}

	// Parse trang_thai (pointer to string)
	var trangThaiPtr *string
	if trangThai != "" {
		trangThaiPtr = &trangThai
	}

	// Get statistics
	stats, err := s.z.GetAdminBookingStatistics(ctx, db.GetAdminBookingStatisticsParams{
		StartDate:  startDate,
		EndDate:    endDate,
		SupplierID: supplierID,
		TrangThai:  trangThaiPtr,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get booking statistics",
			"details": err.Error(),
		})
		return
	}

	// Convert pgtype.Numeric to float64
	var tongTien, tongDoanhThu, tongTienDaHuy, giaTriTrungBinh, giaTriTrungBinhThanhCong float64

	if stats.TongTien.Valid {
		floatVal, _ := stats.TongTien.Float64Value()
		if floatVal.Valid {
			tongTien = floatVal.Float64
		}
	}

	if stats.TongDoanhThu.Valid {
		floatVal, _ := stats.TongDoanhThu.Float64Value()
		if floatVal.Valid {
			tongDoanhThu = floatVal.Float64
		}
	}

	if stats.TongTienDaHuy.Valid {
		floatVal, _ := stats.TongTienDaHuy.Float64Value()
		if floatVal.Valid {
			tongTienDaHuy = floatVal.Float64
		}
	}

	if stats.GiaTriTrungBinh.Valid {
		floatVal, _ := stats.GiaTriTrungBinh.Float64Value()
		if floatVal.Valid {
			giaTriTrungBinh = floatVal.Float64
		}
	}

	if stats.GiaTriTrungBinhThanhCong.Valid {
		floatVal, _ := stats.GiaTriTrungBinhThanhCong.Float64Value()
		if floatVal.Valid {
			giaTriTrungBinhThanhCong = floatVal.Float64
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking statistics fetched successfully",
		"data": gin.H{
			"tong_so_booking":              stats.TongSoBooking,
			"cho_xac_nhan":                 stats.ChoXacNhan,
			"da_xac_nhan":                  stats.DaXacNhan,
			"da_thanh_toan":                stats.DaThanhToan,
			"hoan_thanh":                   stats.HoanThanh,
			"da_huy":                       stats.DaHuy,
			"tong_tien":                    tongTien,
			"tong_doanh_thu":               tongDoanhThu,
			"tong_tien_da_huy":             tongTienDaHuy,
			"tong_so_khach_hang":           stats.TongSoKhachHang,
			"tong_so_tour":                 stats.TongSoTour,
			"tong_so_nha_cung_cap":         stats.TongSoNhaCungCap,
			"tong_so_khach":                stats.TongSoKhach,
			"tong_so_khach_thanh_cong":     stats.TongSoKhachThanhCong,
			"gia_tri_trung_binh":           giaTriTrungBinh,
			"gia_tri_trung_binh_thanh_cong": giaTriTrungBinhThanhCong,
		},
	})
}

// GetAllBookingsForAdmin godoc
// @Summary Lấy danh sách tất cả booking cho admin
// @Description Lấy danh sách booking với filter và pagination
// @Tags Admin
// @Accept json
// @Produce json
// @Param start_date query string false "Ngày bắt đầu (YYYY-MM-DD)"
// @Param end_date query string false "Ngày kết thúc (YYYY-MM-DD)"
// @Param supplier_id query string false "ID Nhà cung cấp (UUID)"
// @Param trang_thai query string false "Trạng thái booking"
// @Param search query string false "Tìm kiếm"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/bookings [get]
func (s *Server) GetAllBookingsForAdmin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse query parameters
	var startDate, endDate pgtype.Timestamp
	var supplierID pgtype.UUID
	trangThai := c.Query("trang_thai")
	searchKeyword := c.DefaultQuery("search", "")

	// Parse start_date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	// Parse end_date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	// Parse supplier_id
	if supplierIDStr := c.Query("supplier_id"); supplierIDStr != "" {
		if err := supplierID.Scan(supplierIDStr); err == nil {
			supplierID.Valid = true
		}
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Parse trang_thai (pointer to string)
	var trangThaiPtr *string
	if trangThai != "" {
		trangThaiPtr = &trangThai
	}

	// Parse search (pointer to string)
	var searchPtr *string
	if searchKeyword != "" {
		searchPtr = &searchKeyword
	}

	// Get bookings
	bookings, err := s.z.GetAllBookingsForAdmin(ctx, db.GetAllBookingsForAdminParams{
		StartDate:  startDate,
		EndDate:     endDate,
		SupplierID:  supplierID,
		TrangThai:   trangThaiPtr,
		Search:      searchPtr,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get bookings",
			"details": err.Error(),
		})
		return
	}

	// Get total count
	totalCount, err := s.z.CountAllBookingsForAdmin(ctx, db.CountAllBookingsForAdminParams{
		StartDate:  startDate,
		EndDate:    endDate,
		SupplierID: supplierID,
		TrangThai:  trangThaiPtr,
		Search:     searchPtr,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to count bookings",
			"details": err.Error(),
		})
		return
	}

	// Convert bookings to response format
	var bookingList []gin.H
	for _, booking := range bookings {
		var tongTien float64
		if booking.TongTien.Valid {
			floatVal, _ := booking.TongTien.Float64Value()
			if floatVal.Valid {
				tongTien = floatVal.Float64
			}
		}

		bookingList = append(bookingList, gin.H{
			"id":                 booking.ID,
			"booking_id":         booking.ID,
			"customer_id":        booking.CustomerID,
			"customer_name":      booking.CustomerName,
			"user_name":          booking.CustomerName,
			"customer_email":     booking.CustomerEmail,
			"customer_phone":     booking.CustomerPhone,
			"tour_id":            booking.TourID,
			"tour_title":         booking.TourTitle,
			"supplier_id":        booking.SupplierID,
			"supplier_name":      booking.SupplierName,
			"departure_id":       booking.DepartureID,
			"ngay_khoi_hanh":     booking.NgayKhoiHanh,
			"ngay_ket_thuc":      booking.NgayKetThuc,
			"departure_status":   booking.DepartureStatus,
			"so_nguoi_lon":       booking.SoNguoiLon,
			"so_tre_em":          booking.SoTreEm,
			"tong_tien":          tongTien,
			"tong_gia":           tongTien,
			"total_amount":       tongTien,
			"don_vi_tien_te":     booking.DonViTienTe,
			"trang_thai":         booking.TrangThai,
			"status":             booking.TrangThai,
			"phuong_thuc_thanh_toan": booking.PhuongThucThanhToan,
			"ngay_dat":           booking.NgayDat,
			"ngay_cap_nhat":      booking.NgayCapNhat,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Bookings fetched successfully",
		"data":       bookingList,
		"total":      totalCount,
		"page":       page,
		"limit":      limit,
		"total_pages": (int(totalCount) + limit - 1) / limit,
	})
}
