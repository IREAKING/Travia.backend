package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Middleware sinh RequestID cho mỗi request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.New().String()
		c.Writer.Header().Set("X-Request-ID", id)
		c.Set("requestID", id)
		c.Next()
	}
}
