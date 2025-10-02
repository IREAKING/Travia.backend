package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse là cấu trúc chuẩn cho response thành công
type SuccessResponse struct {
	Status string `json:"status"`         // luôn là "success"
	Code   int    `json:"code"`           // HTTP status code
	Data   any    `json:"data,omitempty"` // dữ liệu trả về
	Meta   any    `json:"meta,omitempty"` // thông tin phụ (paging, total, v.v.)
}

func JsonSuccess(c *gin.Context, code int, data any, meta any) {
	c.JSON(code, SuccessResponse{
		Status: "success",
		Code:   code,
		Data:   data,
		Meta:   meta,
	})
}

// Ok trả về response 200 OK với data
func Ok(c *gin.Context, data any) {
	JsonSuccess(c, http.StatusOK, data, nil)
}

// Created trả về response 201 Created
func Created(c *gin.Context, data any) {
	JsonSuccess(c, http.StatusCreated, data, nil)
}

// NoContent trả về 204 No Content
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
