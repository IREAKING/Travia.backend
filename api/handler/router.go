package handler

import (
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
		// Public
		auth.POST("/createUserForm", s.CreateUserForm)
		auth.POST("/createUser", s.CreateUser)
		auth.POST("/login", s.Login)

		// Protected
		authAuth := auth.Group("")
		authAuth.Use(middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret))
		{
			authAuth.GET("/getUserById/:id", s.GetUserById)
			authAuth.POST("/logout", s.Logout)
			authAuth.PUT("/updateUserById/:id", middleware.SelfOrRoles("quan_tri"), s.UpdateUserById)
		}
		oauth := auth.Group("/oauth")
		{
			oauth.GET("/:provider", s.AuthHandler())
			oauth.GET("/:provider/callback", s.AuthCallbackHandler())
		}
	}
	tour := api.Group("/tour")
	{
		tour.GET("/getAllTourCategory", s.GetAllTourCategory)
		tour.GET("/getAllTour", s.GetAllTour)
		tour.GET("/getTourDetailByID/:id", s.GetTourDetailByID)
	}
	// Admin routes - Only for quan_tri role
	admin := api.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
		middleware.RequireRoles("quan_tri"),
	)
	{
		// Dashboard & Summary
		admin.GET("/getAdminSummary", s.GetAdminSummary)

		// Revenue analytics
		admin.GET("/getRevenueByMonth", s.GetRevenueByMonth)
		admin.GET("/getRevenueByYear", s.GetRevenueByYear)
		admin.GET("/getRevenueByDateRange", s.GetRevenueByDateRange)
		admin.GET("/getRevenueBySupplier", s.GetRevenueBySupplier)

		// Bookings & Tours
		admin.GET("/getBookingsByStatus", s.GetBookingsByStatus)
		admin.GET("/getBookingsByMonth", s.GetBookingsByMonth)
		admin.GET("/getTopToursByBookings", s.GetTopToursByBookings)
		admin.GET("/getToursByCategory", s.GetToursByCategory)
		admin.GET("/getUpcomingDepartures", s.GetUpcomingDepartures)

		// Users & Customers
		admin.GET("/getNewUsersByMonth", s.GetNewUsersByMonth)
		admin.GET("/getUserGrowth", s.GetUserGrowth)
		admin.GET("/getTopCustomers", s.GetTopCustomers)

		// Suppliers & Reviews
		admin.GET("/getTopSuppliers", s.GetTopSuppliers)
		admin.GET("/getReviewStatsByTour", s.GetReviewStatsByTour)
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
	// docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", s.config.ServerConfig.Host, s.config.ServerConfig.Port)
	docs.SwaggerInfo.Host = "https://travia-363518914287.europe-west1.run.app"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	docs.SwaggerInfo.BasePath = "/api"

	// Thêm thông tin contact
	docs.SwaggerInfo.InfoInstanceName = "swagger"

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
