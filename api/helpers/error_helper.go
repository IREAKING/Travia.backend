package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

func JsonError(c *gin.Context, code int, message string, detail any) {
	c.AbortWithStatusJSON(code, ErrorResponse{
		Status:  "error",
		Code:    code,
		Message: message,
		Detail:  detail,
	})
}
func BadRequest(c *gin.Context, message string, detail any) {
	JsonError(c, http.StatusBadRequest, message, detail)
}
func Unauthorized(c *gin.Context, message string, detail any) {
	JsonError(c, http.StatusUnauthorized, message, detail)
}
func Forbidden(c *gin.Context, message string) {
	JsonError(c, http.StatusForbidden, message, nil)
}

func NotFound(c *gin.Context, message string) {
	JsonError(c, http.StatusNotFound, message, nil)
}

func InternalServerError(c *gin.Context, message string, detail any) {
	JsonError(c, http.StatusInternalServerError, message, detail)
}
