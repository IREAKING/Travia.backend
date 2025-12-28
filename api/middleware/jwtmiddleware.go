package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"travia.backend/api/utils"
)

// AuthMiddleware chỉ hỗ trợ Bearer token (giữ nguyên để backward compatibility)
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Debug: Log incoming request path
		fmt.Printf("AuthMiddleware - Path: %s, Method: %s\n", c.Request.URL.Path, c.Request.Method)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Printf("AuthMiddleware - Missing Authorization header for path: %s\n", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Thiếu header Authorization",
				"path":  c.Request.URL.Path,
			})
			c.Abort()
			return
		}

		// Kiểm tra định dạng Bearer
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Header Authorization không đúng định dạng Bearer. Định dạng đúng: 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token không được để trống",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(token, secretKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token không hợp lệ hoặc đã hết hạn",
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

// RequireRoles cho phép truy cập nếu vai trò của người dùng thuộc một trong các roles được chỉ định
func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[strings.ToLower(r)] = struct{}{}
	}
	return func(c *gin.Context) {
		v, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiếu thông tin xác thực"})
			c.Abort()
			return
		}
		claims, ok := v.(*utils.JwtClams)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thông tin xác thực không hợp lệ"})
			c.Abort()
			return
		}
		if _, ok := allowed[strings.ToLower(claims.Vaitro)]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Không có quyền truy cập"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SelfOrRoles cho phép nếu là chính chủ theo param :id hoặc có vai trò trong danh sách
func SelfOrRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[strings.ToLower(r)] = struct{}{}
	}
	return func(c *gin.Context) {
		v, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiếu thông tin xác thực"})
			c.Abort()
			return
		}
		claims, ok := v.(*utils.JwtClams)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thông tin xác thực không hợp lệ"})
			c.Abort()
			return
		}
		// Cho phép nếu là admin (quan_tri) hoặc các vai trò khác trong danh sách
		if _, ok := allowed[strings.ToLower(claims.Vaitro)]; ok {
			c.Next()
			return
		}
		// Nếu không, chỉ cho phép nếu id khớp
		pathID := c.Param("id")
		if pathID != "" && strings.EqualFold(pathID, claims.Id.String()) {
			c.Next()
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "Không có quyền thực hiện hành động này"})
		c.Abort()
	}
}
