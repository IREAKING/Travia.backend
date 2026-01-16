package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"travia.backend/api/middleware"
	"travia.backend/docs"
)

func (s *Server) SetupRoutes() {
	// Test route
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Travia API is running"})
	})

	api := s.router.Group("/api")
	auth := api.Group("/auth")
	{
		// Public - Registration & Login
		auth.POST("/createUserForm", middleware.RateLimitMiddleware(s.redis, 10, 1*time.Minute), s.CreateUserForm)
		auth.POST("/createUser", middleware.RateLimitMiddleware(s.redis, 10, 1*time.Minute), s.CreateUser)

		// Login endpoints với phân quyền
		auth.POST("/login/user", middleware.RateLimitMiddleware(s.redis, 15, 1*time.Minute), s.LoginUser)         // Đăng nhập cho khách hàng
		auth.POST("/login/admin", middleware.RateLimitMiddleware(s.redis, 15, 1*time.Minute), s.LoginAdmin)       // Đăng nhập cho admin
		auth.POST("/login/supplier", middleware.RateLimitMiddleware(s.redis, 15, 1*time.Minute), s.LoginSupplier) // Đăng nhập cho nhà cung cấp
		auth.POST("/login", middleware.RateLimitMiddleware(s.redis, 15, 1*time.Minute), s.Login)                  // Deprecated - giữ để backward compatibility
		auth.POST("/refresh", middleware.RateLimitMiddleware(s.redis, 30, 1*time.Minute), s.RefreshToken)         // Làm mới token
		// Password reset flow (3 steps)
		auth.POST("/forgot-password/request", middleware.RateLimitMiddleware(s.redis, 10, 1*time.Minute), s.RequestPasswordReset) // Step 1: Request OTP
		auth.POST("/forgot-password/verify", middleware.RateLimitMiddleware(s.redis, 10, 1*time.Minute), s.VerifyOTP)             // Step 2: Verify OTP
		auth.POST("/forgot-password/reset", middleware.RateLimitMiddleware(s.redis, 10, 1*time.Minute), s.ResetPassword)          // Step 3: Reset password
		// Protected
		authAuth := auth.Group("")
		authAuth.Use(middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret)) // Chỉ đọc từ Authorization header
		{
			authAuth.GET("/getUserById/:id", s.GetUserById)
			authAuth.PUT("/updateUser", s.UpdateUser)
			authAuth.POST("/logout", s.Logout)
			authAuth.PUT("/updateUserById/:id", middleware.SelfOrRoles("quan_tri"), s.UpdateUserById)
			authAuth.PUT("/changePassword", s.ChangePassword) // Cần xác thực để đổi mật khẩu
		}
		oauth := auth.Group("/oauth")
		{
			oauth.GET("/:provider", s.AuthHandler())
			oauth.GET("/:provider/callback", s.AuthCallbackHandler())
		}
	}

	storage := api.Group("/storage")
	{
		storage.Use(middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret)) // Chỉ đọc từ Authorization header
		{
			storage.POST("/upload", s.UploadImage)
			storage.POST("/upload-multiple", s.UploadMultipleImages)
			storage.POST("/upload-tour-images", s.UploadImagesForTour)
		}
		storage.GET("/get-signed-pdf/:filename", s.GetSignedPDF)
	}
	// ========== TOUR ROUTES (with Redis caching) ==========
	tour := api.Group("/tour")
	{
		// Public GET requests (cached)
		tour.GET("/categories",
			// middleware.CacheMiddleware(s.redis, 6*time.Hour),
			s.GetAllTourCategory,
		)
		tour.GET("/",
			// middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetAllTour,
		)
		tour.GET("/:id",
			// middleware.CacheMiddleware(s.redis, 2*time.Hour),
			s.GetTourDetailByID,
		)
		tour.GET("/:id/reviews",
			// middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetReviewByTourId,
		)
		tour.GET("/filter",
			// middleware.CacheMiddleware(s.redis, 10*time.Minute),
			s.FilterTours,
		)
		tour.GET("/search",
			// middleware.CacheMiddleware(s.redis, 10*time.Minute),
			s.SearchTours,
		)
		tour.POST("/",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.InvalidateCacheMiddleware(s.redis, "cache:http:GET:/tour*"),
			s.CreateTour,
		)
		tour.GET("/discount/:id", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.GetDiscountsByTourID)
		tour.POST("/discount", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.CreateDiscountTour)
		tour.PUT("/discount", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.UpdateDiscountTour)
		tour.DELETE("/discount/:id", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.DeleteDiscountTour)
		tour.PUT("/:id",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.InvalidateCacheMiddleware(s.redis, "cache:http:GET:/tour*"),
			s.UpdateTour,
		)
	}
	// ========== ADMIN ROUTES (with short cache for fresh stats) ==========
	admin := api.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
		middleware.RequireRoles("quan_tri"),
	)
	{
		admin.GET("/supplierOptions",
			s.GetSupplierOptions,
		)
		admin.GET("/getDashboardOverview",
			s.GetDashboardOverview,
		)
		admin.GET("/getDashboardOverviewByMonthAndYear",
			s.GetDashboardOverviewByMonthAndYear,
		)
		admin.GET("/getUserStatsByRole",
			s.GetUserStatsByRole,
		)

		admin.GET("/getTopBookedTours",
			s.GetTopBookedTours,
		)
		admin.GET("/getTourPriceDistribution",
			s.GetTourPriceDistribution,
		)
		admin.GET("/getRevenueByDay",
			s.GetRevenueByDay,
		)
		admin.GET("/getBookingsByDayOfWeek",
			s.GetBookingsByDayOfWeek,
		)
		admin.GET("/getRecentBookings",
			s.GetRecentBookings,
		)
		admin.GET("/transactions",
			s.GetTransactions,
		)
		admin.GET("/chartRevenueTrend",
			s.AdminChartRevenueTrend,
		)
		admin.GET("/chartCategoryDistribution",
			s.AdminChartCategoryDistribution,
		)
		admin.GET("/chartTopSuppliers",
			s.AdminChartTopSuppliers,
		)
		admin.GET("/chartBookingStatusStats",
			s.AdminChartBookingStatusStats,
		)

		//=====================================Nhà cung cấp=====================================
		admin.GET("/suppliers",
			s.GetAllSuppliers,
		)
		admin.PUT("/suppliers/approve/:id",
			s.ApproveSupplier,
		)
		admin.PUT("/suppliers/reject/:id",
			s.RejectSupplier,
		)
		admin.DELETE("/suppliers/soft-delete/:id",
			s.SoftDeleteSupplier,
		)
		admin.PUT("/suppliers/restore/:id",
			s.RestoreSupplier,
		)
		admin.GET("/suppliers/:id",
			s.GetSupplierByID,
		)
		//=====================================Khách hàng=====================================
		admin.GET("/customers/getTopActiveUsers",
			s.GetTopActiveUsers,
		)
		admin.GET("/customers/adminCustomerGrowthMonthlyReport",
			s.AdminCustomerGrowthMonthlyReport,
		)
		//=====================================Booking Management=====================================
		admin.GET("/bookings",
			s.GetAllBookingsForAdmin,
		)
		admin.GET("/bookings/statistics",
			s.GetAdminBookingStatistics,
		)
		//=====================================Refund Management=====================================
		admin.GET("/refunds",
			s.GetAllRefunds,
		)
		admin.GET("/refunds/stats",
			s.GetRefundStats,
		)
	}
	// ========== DESTINATION ROUTES (with Redis caching) ==========
	destination := api.Group("/destination")
	{
		destination.GET("/country",
			// middleware.CacheMiddleware(s.redis, 6*time.Hour),
			s.GetCountry,
		)
		destination.GET("/province/:country",
			//middleware.CacheMiddleware(s.redis, 6*time.Hour),
			s.GetProvinceByCountry,
		)
		destination.GET("/city/:province",
			// middleware.CacheMiddleware(s.redis, 6*time.Hour),
			s.GetCityByProvince,
		)
		destination.GET("/popular",
			// middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetPopularDestinations,
		)
		destination.GET("/top",
			// middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetTopPopularDestinations,
		)
		destination.GET("/:id",
			// middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetDestinationByID,
		)
		destination.GET("/:id/tours",
			// middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetToursByDestination,
		)
		// Write operations - Invalidate cache on success
		destWrite := destination.Group("")
		destWrite.Use(middleware.InvalidateCacheMiddleware(s.redis,
			"cache:http:*destination*",
		))
		{
			destWrite.POST("/createDestination", s.CreateDestination)
			destWrite.PUT("/:id/image",
				middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
				middleware.RequireRoles("quan_tri"),
				s.UpdateDestinationImage,
			)
		}
	}
	// ========== SUPPLIER ROUTES (with Redis caching) ==========
	supplier := api.Group("/supplier")
	{
		// Đăng ký đối tác - công khai, không cần auth
		supplier.POST("/register", middleware.RateLimitMiddleware(s.redis, 5, 1*time.Minute), s.RegisterPartner)

		// Create supplier - Admin only
		supplier.POST("/createSupplier",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.CreateSupplier,
		)

		// Specific routes must be defined before parameterized routes
		supplier.GET("/tours/my",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetMyTours,
		)
		supplier.GET("/info",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetInfoSupplier,
		)

		// Dashboard routes - must be before parameterized routes
		supplier.GET("/dashboard/overview",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierDashboardOverview,
		)
		supplier.GET("/dashboard/revenue-by-time",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierRevenueByTimeRange,
		)
		supplier.GET("/dashboard/top-tours",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierTopTours,
		)
		supplier.GET("/dashboard/booking-stats",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierBookingStatsByStatus,
		)
		supplier.GET("/dashboard/tour-stats",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierTourStatsByStatus,
		)
		supplier.GET("/dashboard/tour-stats-by-category",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierTourStatsByCategory,
		)
		supplier.GET("/dashboard/revenue-chart",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierRevenueChart,
		)
		supplier.GET("/dashboard/customer-stats",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierCustomerStats,
		)

		supplier.GET("/dashboard/rating-analysis",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierRatingAnalysis,
		)
		supplier.GET("/dashboard/upcoming-departures",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierUpcomingDepartures,
		)
		supplier.GET("/dashboard/recent-bookings",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierRecentBookings,
		)
		supplier.GET("/revenue/statistics",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierRevenueStatistics,
		)
		supplier.GET("/revenue/transactions",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierTransactions,
		)
		// Advanced bookings query - must be before parameterized routes
		supplier.GET("/bookings/advanced",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierBookingsByStatusAdvanced,
		)
		supplier.PUT("/tours/update-status/:id",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.UpdateTourStatus,
		)
		supplier.GET("/search/:keyword",
			middleware.CacheMiddleware(s.redis, 10*time.Minute),
			s.SearchSuppliers,
		)
		supplier.GET("/count",
			middleware.CacheMiddleware(s.redis, 10*time.Minute),
			s.CountSuppliers,
		)
		supplier.GET("/active",
			middleware.CacheMiddleware(s.redis, 10*time.Minute),
			s.GetActiveSuppliers,
		)

		supplier.GET("/dashboard/review-statistics",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierReviewStatistics,
		)
		supplier.GET("/dashboard/reviews",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetDetailedSupplierReviews,
		)
		supplier.GET("/dashboard/options-tour",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetOptionTour,
		)
		//=====================================Refund Management=====================================
		supplier.GET("/refunds",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierRefunds,
		)
		supplier.GET("/refunds/stats",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierRefundStats,
		)

		// Write operations - require authentication and invalidate cache
		supplierWrite := supplier.Group("")
		supplierWrite.Use(
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.InvalidateCacheMiddleware(s.redis, "cache:http:*supplier*"),
		)
	}
	// Location - IP Geolocation detection
	location := api.Group("/location")
	{
		location.GET("", s.GetLocation)              // Auto-detect IP or use ?ip=xxx.xxx.xxx.xxx
		location.GET("/:ip", s.GetLocationByIP)      // Get location by specific IP
		location.GET("/tours", s.GetToursByLocation) // Get domestic and international tours by user location
		location.GET("/debug", s.GetClientIPDebug)   // Debug endpoint to see detected IP and headers
		location.GET("/test", s.Handler)             // Test endpoint to see detected IP and headers
	}

	// Recommendation - AI Tour Recommendation
	recommendation := api.Group("/recommendation")
	{
		recommendation.POST("/track-view", middleware.RateLimitMiddleware(s.redis, 60, 1*time.Minute), s.TrackTourView) // Lưu lịch sử xem tour
		recommendation.GET("/tours", s.GetRecommendedTours)                                                             // Lấy tour gợi ý
		recommendation.GET("/similar/:tour_id", s.GetSimilarTours)                                                      // Lấy tour tương tự
	}

	// ========== PAYMENT ROUTES (with rate limiting) ==========
	payment := api.Group("/payment")
	{
		vnpay := payment.Group("/vnpay")
		{
			vnpay.POST("/create",
				middleware.RateLimitMiddleware(s.redis, 10, 1*time.Minute),
				middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
				s.CreateVNPayPaymentURL,
			)
			vnpay.GET("/return", s.VNPayCallback)
			vnpay.GET("/verify", s.VNPayVerifyCallback)
			vnpay.POST("/ipn", s.VNPayIPN)
		}
	}

	// ========== CONTACT ROUTES ==========
	contact := api.Group("/contact")
	{
		contact.POST("", middleware.RateLimitMiddleware(s.redis, 10, 1*time.Minute), s.CreateContact)
		contactAdmin := contact.Group("")
		contactAdmin.Use(
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
		)
		{
			contactAdmin.GET("", s.GetAllContacts)
			contactAdmin.GET("/unread", s.GetUnreadContacts)
			contactAdmin.GET("/status/:status", s.GetContactsByStatus)
			contactAdmin.GET("/:id", s.GetContactByID)
			contactAdmin.PUT("/:id/status", s.UpdateContactStatus)
			contactAdmin.PUT("/:id/read", s.MarkContactAsRead)
			contactAdmin.POST("/:id/response", s.CreateContactResponse)
			contactAdmin.GET("/:id/responses", s.GetContactResponses)
		}
	}

	// ========== NOTIFICATION ROUTES ==========
	notification := api.Group("/notifications")
	{
		notificationAuth := notification.Group("")
		notificationAuth.Use(middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret))
		{
			notificationAuth.GET("", s.GetMyNotifications)
			notificationAuth.GET("/unread", s.GetUnreadNotifications)
			notificationAuth.GET("/count", s.GetNotificationCount)
			notificationAuth.PUT("/:id/read", s.MarkNotificationAsRead)
			notificationAuth.PUT("/read-all", s.MarkAllNotificationsAsRead)
		}
	}

	// ========== BOOKING ROUTES (Core functionality) ==========
	booking := api.Group("/booking")
	{
		// Protected endpoints (require authentication) - Đăng ký trước để tránh conflict
		bookingAuth := booking.Group("")
		bookingAuth.Use(middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret))
		{
			bookingAuth.POST("/create", middleware.RateLimitMiddleware(s.redis, 20, 1*time.Minute), s.CreateBooking)
			bookingAuth.POST("/add-passengers", s.AddPassengers)
			bookingAuth.GET("/:id", s.GetBookingById)
			bookingAuth.GET("/:id/calculate-refund", s.CalculateRefundAmount)
			bookingAuth.PUT("/:id/cancel", s.CancelBooking)
			bookingAuth.GET("/my-bookings",
				middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
				middleware.RequireRoles("khach_hang"),
				s.GetMyBookings)
			bookingAuth.DELETE("/:id", s.DeleteBooking)
			bookingAuth.DELETE("/delete-bookings", s.DeleteBookings)
		}

		// Public endpoints (no auth required) - Đăng ký sau
		booking.POST("/hold-seat/:khoi_hanh_id/:so_nguoi_lon/:so_tre_em", middleware.RateLimitMiddleware(s.redis, 30, 1*time.Minute), s.HoldSeat)
	}

	// ========== DEPARTURE ROUTES (Tour schedule management) ==========
	departure := api.Group("/departure")
	{
		// Public GET requests (cached)
		departure.GET("/:id",
			middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetDepartureByID,
		)
		departure.GET("/tour/:tour_id",
			middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetDeparturesByTour,
		)
		departure.GET("/upcoming",
			middleware.CacheMiddleware(s.redis, 15*time.Minute),
			s.GetUpcomingDeparturesPublic,
		)

		// Write operations - require authentication and invalidate cache
		departureWrite := departure.Group("")
		departureWrite.Use(
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.InvalidateCacheMiddleware(s.redis, "cache:http:*departure*"),
		)
		{
			// Admin/Supplier only
			departureWrite.POST("/create",
				middleware.RequireRoles("nha_cung_cap"),
				s.CreateDeparture,
			)
			departureWrite.PUT("/:id",
				middleware.RequireRoles("nha_cung_cap"),
				s.UpdateDeparture,
			)
			departureWrite.DELETE("/:id",
				middleware.RequireRoles("nha_cung_cap"),
				s.DeleteDeparture,
			)
			departureWrite.PUT("/:id/cancel",
				middleware.RequireRoles("nha_cung_cap"),
				s.CancelDeparture,
			)
			departureWrite.PUT("/lich-trinh/:id",
				middleware.RequireRoles("nha_cung_cap"),
				s.UpdateLichTrinh,
			)
			departureWrite.PUT("/hoat-dong-trong-ngay/:id",
				middleware.RequireRoles("nha_cung_cap"),
				s.UpdateHoatDongTrongNgay,
			)
			departureWrite.POST("/add-hinh-anh/:id",
				middleware.RequireRoles("nha_cung_cap"),
				s.AddHinhAnhTour,
			)
			departureWrite.DELETE("/delete-hinh-anh/:id",
				middleware.RequireRoles("nha_cung_cap"),
				s.DeleteHinhAnhTour,
			)
			departureWrite.POST("/add-tour-destination/:id",
				middleware.RequireRoles("nha_cung_cap"),
				s.AddTourDestination,
			)
			departureWrite.DELETE("/delete-tour-destination/:tour_id/:diem_den_id",
				middleware.RequireRoles("nha_cung_cap"),
				s.DeleteTourDestination,
			)
		}
	}

	// ========== REVIEW ROUTES (Customer feedback) ==========
	review := api.Group("/review")
	{
		review.GET("/tour/:id",
			middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetReviewByTourId,
		)

		// Protected endpoints - require authentication
		reviewAuth := review.Group("")
		reviewAuth.Use(
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.InvalidateCacheMiddleware(s.redis, "cache:http:*review*"),
		)
		{
			reviewAuth.POST("/create", middleware.RateLimitMiddleware(s.redis, 30, 1*time.Minute), s.CreateReview)
			reviewAuth.GET("/check/:dat_cho_id", s.CheckReviewStatus)
		}

	}

	// ========== IMAGES ROUTES (Image management) ==========
	images := api.Group("/images")
	{
		// Admin only - Update destination images from Pexels
		imagesAuth := images.Group("")
		imagesAuth.Use(
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			middleware.InvalidateCacheMiddleware(s.redis, "cache:http:*destination*"),
		)
	}
	// ========== FAVORITE ROUTES (Favorite management) ==========
	favorite := api.Group("/favorite")
	{
		favorite.POST("/", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.CreateFavoriteTour)
		favorite.DELETE("/", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.DeleteFavoriteTour)
		favorite.GET("/", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.GetFavoriteTours)
	}
	ticket := api.Group("/ticket")
	{
		ticket.GET("/:dat_cho_id",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			s.PrintTicket,
		)
	}

	// ========== BLOG ROUTES ==========
	blog := api.Group("/blog")
	{
		// Public endpoints
		blog.GET("/posts", s.GetPublishedBlogs)
		blog.GET("/posts/:slug", s.GetBlogBySlug)
		blog.GET("/featured", s.GetFeaturedBlogs)
		blog.GET("/search", s.SearchBlogs)
		blog.GET("/category/:category", s.GetBlogsByCategory)
		blog.POST("/posts/:id/view", s.IncrementBlogViews)

		// Admin endpoints
		blogAdmin := blog.Group("")
		blogAdmin.Use(
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
		)
		{
			blogAdmin.GET("/admin", s.GetAllBlogsForAdmin)
			blogAdmin.GET("/admin/:id", s.GetBlogByIDForAdmin)
			blogAdmin.POST("/admin", s.CreateBlog)
			blogAdmin.PUT("/admin/:id", s.UpdateBlog)
			blogAdmin.DELETE("/admin/:id", s.DeleteBlog)
			blogAdmin.GET("/admin/stats", s.GetBlogStats)
			blogAdmin.POST("/admin/ai/generate", s.GenerateBlogContent)
			blogAdmin.POST("/admin/ai/titles", s.GenerateBlogTitleSuggestions)
			blogAdmin.POST("/admin/ai/create", s.CreateBlogWithAI)
			blogAdmin.GET("/admin/:id/ai-history", s.GetBlogAIHistory)
		}
	}

	// ========== AI ROUTES ==========
	ai := api.Group("/ai")
	{
		ai.POST("/chatbot", middleware.RateLimitMiddleware(s.redis, 20, 1*time.Minute), s.Chatbot) // Public endpoint for chatbot
	}

}

func (s *Server) SetupMiddlewares() {
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.RequestID())
	// s.router.Use(gzip.Gzip(gzip.DefaultCompression))
	s.router.Use(middleware.TimeoutMiddleware(10 * time.Second))
	s.router.Use(middleware.SetupCORS())
}
func (s *Server) SetupSwagger() {
	// Cập nhật thông tin Swagger với API key authentication
	docs.SwaggerInfo.Title = "Travia API"
	docs.SwaggerInfo.Description = "Travia Travel Management API with Google OAuth"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", "localhost", s.config.ServerConfig.Port)
	//docs.SwaggerInfo.Host = "travia-backend-363518914287.asia-southeast1.run.app"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	docs.SwaggerInfo.BasePath = "/api"

	// Thêm thông tin contact
	docs.SwaggerInfo.InfoInstanceName = "swagger"

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
