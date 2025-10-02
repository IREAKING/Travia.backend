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
		auth.GET("/getUserById/:id", s.GetUserById)
		auth.POST("/createUserForm", s.CreateUserForm)
		auth.POST("/createUser", s.CreateUser)
		auth.POST("/login", s.Login)
		auth.POST("/logout", s.Logout)
		auth.PUT("/updateUserById/:id", s.UpdateUserById)
		oauth := auth.Group("/oauth")
		{
			oauth.GET("/:provider", s.AuthHandler())
			oauth.GET("/:provider/callback", s.AuthCallbackHandler())
		}
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
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", s.config.ServerConfig.Host, s.config.ServerConfig.Port)
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	docs.SwaggerInfo.BasePath = "/api"

	// Thêm thông tin contact
	docs.SwaggerInfo.InfoInstanceName = "swagger"

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
