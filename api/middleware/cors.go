package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupCORS() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()

	corsConfig.AllowAllOrigins = true
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
