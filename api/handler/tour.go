package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)
// lấy danh mục tour
// @summary Lấy danh mục tour
// @description Lấy danh mục tour
// @tags tour
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/getAllTourCategory [get]
func (s *Server) GetAllTourCategory(c *gin.Context) {
	tourCategories, err := s.z.GetAllTourCategory(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể lấy danh mục tour",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh mục tour thành công",
		"data":    tourCategories,
	})
}
