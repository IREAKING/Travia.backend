package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupCORS() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()

	// Cho phép nhiều origins cho development
	corsConfig.AllowOrigins = []string{
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:3000",
		"http://localhost:4173",
		"http://127.0.0.1:5173",
		"http://127.0.0.1:5174",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:4173",
		"https://travia-frontend.vercel.app",
		"https://travia-frontend.vercel.app/",
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-API-Key",
		"X-Requested-With",
		"Cache-Control",
	}
	corsConfig.AllowCredentials = true
	corsConfig.ExposeHeaders = []string{
		"Content-Length",
		"Content-Type",
		"X-Total-Count",
		"X-Page-Count",
	}
	corsConfig.MaxAge = 86400

	return cors.New(corsConfig)
}
