package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)
// Middleware giới hạn thời gian xử lý
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// tạo chanel done
		done := make(chan struct{})
		panicChan := make(chan interface{})
		go func ()  {
			defer func ()  {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			close(done)
		}()
		select {
		case p := <- panicChan:
			panic(p)
		case <- done:
			return 
		case <- time.After(timeout):
			c.JSON(http.StatusInternalServerError, gin.H {"error": "yêu cầu hết thời gian chờ"})
		}
	}
}