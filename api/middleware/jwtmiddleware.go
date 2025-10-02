package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"travia.backend/api/utils"
)

func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Thiếu header Authorization",
			})
			c.Abort()
			return
		}

		// Kiểm tra định dạng Bearer
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Header Authorization không đúng định dạng Bearer",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateToken(token, secretKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token không hợp lệ",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Đưa claims vào context
		c.Set("claims", claims)

		c.Next()
	}
}
