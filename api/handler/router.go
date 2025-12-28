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
		auth.POST("/createUserForm", s.CreateUserForm)
		auth.POST("/createUser", s.CreateUser)

		// Login endpoints với phân quyền
		auth.POST("/login/user", s.LoginUser)         // Đăng nhập cho khách hàng
		auth.POST("/login/admin", s.LoginAdmin)       // Đăng nhập cho admin
		auth.POST("/login/supplier", s.LoginSupplier) // Đăng nhập cho nhà cung cấp
		auth.POST("/login", s.Login)                  // Deprecated - giữ để backward compatibility
		auth.POST("/refresh", s.RefreshToken)         // Làm mới token
		auth.PUT("/resetPassword/:email", s.ResetPassword)
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
	}
	// ========== TOUR ROUTES (with Redis caching) ==========
	tour := api.Group("/tour")
	{
		// Public GET requests (cached)
		tour.GET("/categories",
			//middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetAllTourCategory,
		)
		tour.GET("/",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetAllTour,
		)
		tour.GET("/:id",
			//middleware.CacheMiddleware(s.redis, 2*time.Hour),
			s.GetTourDetailByID,
		)
		tour.GET("/:id/reviews",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetReviewByTourId,
		)
		tour.GET("/filter",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.FilterTours,
		)
		tour.GET("/search",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.SearchTours,
		)
		tour.POST("/", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.CreateTour)
		tour.GET("/discount/:id", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.GetDiscountsByTourID)
		tour.POST("/discount", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.CreateDiscountTour)
		tour.PUT("/discount", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.UpdateDiscountTour)
		tour.DELETE("/discount/:id", middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret), s.DeleteDiscountTour)
	}
	// ========== ADMIN ROUTES (with short cache for fresh stats) ==========
	admin := api.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
		middleware.RequireRoles("quan_tri"),
	)
	{
		admin.GET("/getDashboardOverview",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetDashboardOverview,
		)
		admin.GET("/getDashboardOverviewWithComparison",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetDashboardOverviewWithComparison,
		)
		admin.GET("/getUserStatsByRole",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetUserStatsByRole,
		)
		admin.GET("/getUserGrowthByMonth",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetUserGrowthByMonth,
		)
		admin.GET("/getUserGrowthByDay",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetUserGrowthByDay,
		)
		admin.GET("/getNewUsersToday",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetNewUsersToday,
		)
		admin.GET("/getTopActiveUsers",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetTopActiveUsers,
		)
		admin.GET("/getTopBookedTours",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetTopBookedTours,
		)
		admin.GET("/getToursCreatedByMonth",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetToursCreatedByMonth,
		)
		admin.GET("/getTourPriceDistribution",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetTourPriceDistribution,
		)
		admin.GET("/getRevenueByDay",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetRevenueByDay,
		)
		admin.GET("/getRevenueByMonth",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetRevenueByMonth,
		)
		admin.GET("/getRevenueByYear",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetRevenueByYear,
		)
		admin.GET("/getBookingsByDayOfWeek",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetBookingsByDayOfWeek,
		)
		admin.GET("/getRecentBookings",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetRecentBookings,
		)
		admin.GET("/getBookingsByStatus",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetBookingStatsByStatus,
		)
		admin.GET("/transactions",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.GetTransactions,
		)
	}
	// ========== DESTINATION ROUTES (with Redis caching) ==========
	destination := api.Group("/destination")
	{
		// GET routes with caching
		destination.GET("/getDestinationByID/:id",
			//middleware.CacheMiddleware(s.redis, 2*time.Hour),
			s.GetDestinationByID,
		)
		destination.GET("/country",
			//middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetCountry,
		)
		destination.GET("/province/:country",
			//middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetProvinceByCountry,
		)
		destination.GET("/city/:province",
			//middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetCityByProvince,
		)
		destination.GET("/getAllDestination",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetAllDestinations,
		)
		destination.GET("/hierarchical",
			//middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetDestinationsHierarchical,
		)
		destination.GET("/for-tour-creation",
			middleware.CacheMiddleware(s.redis, 15*time.Minute),
			s.GetDestinationsForTourCreation,
		)

		// Write operations - Invalidate cache on success
		destWrite := destination.Group("")
		destWrite.Use(middleware.InvalidateCacheMiddleware(s.redis,
			"cache:http:*destination*",
		))
		{
			destWrite.POST("/createDestination", s.CreateDestination)
		}
	}
	// ========== SUPPLIER ROUTES (with Redis caching) ==========
	supplier := api.Group("/supplier")
	{
		// Đăng ký đối tác - công khai, không cần auth
		supplier.POST("/register", s.RegisterPartner)

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
		supplier.GET("/dashboard/cancellation-analysis",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierCancellationAnalysis,
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
		supplier.GET("/dashboard/monthly-comparison",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierMonthlyComparison,
		)

		// Advanced bookings query - must be before parameterized routes
		supplier.GET("/bookings/advanced",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.GetSupplierBookingsByStatusAdvanced,
		)

		supplier.GET("/pending",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetPendingSuppliers,
		)

		// GET routes with caching
		supplier.GET("/:id",
			//middleware.CacheMiddleware(s.redis, 2*time.Hour),
			s.GetSupplierByID,
		)
		supplier.PUT("/approve/:id",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.ApproveSupplier,
		)
		supplier.PUT("/reject/:id",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("quan_tri"),
			s.RejectSupplier,
		)
		supplier.PUT("/tours/update-status/:id",
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.RequireRoles("nha_cung_cap"),
			s.UpdateTourStatus,
		)
		supplier.GET("",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetAllSuppliers,
		)
		supplier.GET("/search/:keyword",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.SearchSuppliers,
		)
		supplier.GET("/count",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.CountSuppliers,
		)
		supplier.GET("/active",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetActiveSuppliers,
		)

		// Write operations - require authentication and invalidate cache
		supplierWrite := supplier.Group("")
		supplierWrite.Use(
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.InvalidateCacheMiddleware(s.redis, "cache:http:*supplier*"),
		)
		{

			supplierWrite.DELETE("/soft-delete/:id",
				middleware.RequireRoles("quan_tri"),
				s.SoftDeleteSupplier,
			)
			supplierWrite.PUT("/restore/:id",
				middleware.RequireRoles("quan_tri"),
				s.RestoreSupplier,
			)
			supplierWrite.DELETE("/delete/:id",
				middleware.RequireRoles("quan_tri"),
				s.DeleteSupplier,
			)
		}
	}
	// Location - IP Geolocation detection
	location := api.Group("/location")
	{
		location.GET("", s.GetLocation)         // Auto-detect IP or use ?ip=xxx.xxx.xxx.xxx
		location.GET("/:ip", s.GetLocationByIP) // Get location by specific IP
	}

	// ========== PAYMENT ROUTES (with rate limiting) ==========
	payment := api.Group("/payment")
	{
		// VNPay routes
		vnpay := payment.Group("/vnpay")
		{
			// Create payment URL (requires authentication)
			vnpay.POST("/create",
				middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
				s.CreateVNPayPaymentURL,
			)
			// Return URL callback (public - VNPay redirects here)
			vnpay.GET("/return", s.VNPayCallback)
			// IPN callback (public - VNPay server calls this)
			vnpay.POST("/ipn", s.VNPayIPN)
		}
	}

	// ========== BOOKING ROUTES (Core functionality) ==========
	booking := api.Group("/booking")
	{
		// Protected endpoints (require authentication) - Đăng ký trước để tránh conflict
		bookingAuth := booking.Group("")
		bookingAuth.Use(middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret))
		{
			bookingAuth.POST("/create", s.CreateBooking)
			bookingAuth.POST("/add-passengers", s.AddPassengers)
			fmt.Println("✅ Route registered: POST /api/booking/add-passengers")
			bookingAuth.GET("/:id", s.GetBookingById)
			fmt.Println("✅ Route registered: GET /api/booking/:id")
			bookingAuth.GET("/my-bookings",
				middleware.RequireRoles("khach_hang"),
				s.GetMyBookings)
		}

		// Public endpoints (no auth required) - Đăng ký sau
		booking.POST("/hold-seat/:khoi_hanh_id/:so_nguoi_lon/:so_tre_em", s.HoldSeat)
	}

	// ========== DEPARTURE ROUTES (Tour schedule management) ==========
	departure := api.Group("/departure")
	{
		// Public GET requests (cached)
		departure.GET("/:id",
			//middleware.CacheMiddleware(s.redis, 1*time.Hour),
			s.GetDepartureByID,
		)
		departure.GET("/tour/:tour_id",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
			s.GetDeparturesByTour,
		)
		departure.GET("/upcoming",
			//middleware.CacheMiddleware(s.redis, 30*time.Minute),
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
				middleware.RequireRoles("quan_tri", "nha_cung_cap"),
				s.CreateDeparture,
			)
			departureWrite.PUT("/:id",
				middleware.RequireRoles("quan_tri", "nha_cung_cap"),
				s.UpdateDeparture,
			)
			departureWrite.DELETE("/:id",
				middleware.RequireRoles("quan_tri", "nha_cung_cap"),
				s.DeleteDeparture,
			)
			departureWrite.PUT("/:id/cancel",
				middleware.RequireRoles("quan_tri", "nha_cung_cap"),
				s.CancelDeparture,
			)
		}
	}

	// ========== REVIEW ROUTES (Customer feedback) ==========
	review := api.Group("/review")
	{
		review.GET("/tour/:id", s.GetReviewByTourId)

		// Protected endpoints - require authentication
		reviewAuth := review.Group("")
		reviewAuth.Use(
			middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
			middleware.InvalidateCacheMiddleware(s.redis, "cache:http:*review*"),
		)

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
	//docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", "localhost", s.config.ServerConfig.Port)
	docs.SwaggerInfo.Host = "travia-backend-363518914287.asia-southeast1.run.app"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	docs.SwaggerInfo.BasePath = "/api"

	// Thêm thông tin contact
	docs.SwaggerInfo.InfoInstanceName = "swagger"

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
