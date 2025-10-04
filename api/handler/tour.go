package handler

import (
	"context"
	"net/http"
	"strconv"

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

// lấy danh sách tour
// @summary Lấy danh sách tour
// @description Lấy danh sách tour
// @tags tour
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/getAllTour [get]
func (s *Server) GetAllTour(c *gin.Context) {
	tours, err := s.z.GetAllTourWithRelations(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy danh sách tour",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh sách tour thành công",
		"data":    tours,
	})
}

// lấy chi tiết tour
// @summary Lấy chi tiết tour
// @description Lấy chi tiết tour
// @tags tour
// @accept json
// @produce json
// @success 200 {object} gin.H "Thành công"
// @failure 500 {object} gin.H "Lỗi server"
// @router /tour/getTourDetailByID/:id [get]
func (s *Server) GetTourDetailByID(c *gin.Context) {
	_id := c.Param("id")
	id, err := strconv.ParseInt(_id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "ID không hợp lệ",
		})
		return
	}
	tour, err := s.z.GetTourDetailByID(context.Background(), int32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy chi tiết tour thành công",
		"data":    tour,
	})
}
