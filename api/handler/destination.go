package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	db "travia.backend/db/sqlc"
)

// Tạo điểm đến
// @Summary Tạo điểm đến
// @Description Tạo điểm đến
// @Tags Destination
// @Accept json
// @Produce json
// @Param destination body db.CreateDestinationParams true "Destination"
// @Success 200 {object} db.CreateDestinationParams
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination [post]
func (s *Server) CreateDestination(c *gin.Context) {
	var input db.CreateDestinationParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := s.z.CreateDestination(context.Background(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Destination created successfully", "data": data})
}

// Lấy danh sách quốc gia
// @Summary Lấy danh sách quốc gia
// @Description Lấy danh sách quốc gia
// @Tags Destination
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/country [get]
func (s *Server) GetCountry(c *gin.Context) {
	countryNames, err := s.z.GetCountry(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "không thể lấy dữ liệu",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    countryNames,
	})
}

// Lấy danh sách tỉnh thành theo quốc gia
// @Summary Lấy danh sách tỉnh thành theo quốc gia
// @Description Lấy danh sách tỉnh thành theo quốc gia
// @Tags Destination
// @Accept json
// @Produce json
// @Param country path string true "Country"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/province/{country} [get]
func (s *Server) GetProvinceByCountry(c *gin.Context) {
	country := c.Param("country")
	provinceNames, err := s.z.GetProvinceByCountry(context.Background(), &country)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "không thể lấy dữ liệu",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    provinceNames,
	})
}

// Lấy danh sách thành phố theo tỉnh thành
// @Summary Lấy danh sách thành phố theo tỉnh thành
// @Description Lấy danh sách thành phố theo tỉnh thành
// @Tags Destination
// @Accept json
// @Produce json
// @Param province path string true "Province"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/city/{province} [get]
func (s *Server) GetCityByProvince(c *gin.Context) {
	province := c.Param("province")
	cityNames, err := s.z.GetCityByProvince(context.Background(), &province)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "không thể lấy dữ liệu",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    cityNames,
	})
}

// Lấy danh sách điểm đến phổ biến nhất
// @Summary Lấy danh sách điểm đến phổ biến nhất
// @Description Lấy danh sách điểm đến phổ biến nhất (được nhiều tour sử dụng)
// @Tags Destination
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng điểm đến" default(10)
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/popular [get]
func (s *Server) GetPopularDestinations(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	destinations, err := s.z.GetPopularDestinations(context.Background(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "không thể lấy dữ liệu",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    destinations,
	})
}

// Lấy top điểm đến phổ biến nhất với thông tin chi tiết
// @Summary Lấy top điểm đến phổ biến nhất
// @Description Lấy top điểm đến phổ biến nhất với số tour nổi bật
// @Tags Destination
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng điểm đến" default(10)
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/top [get]
func (s *Server) GetTopPopularDestinations(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	destinations, err := s.z.GetTopPopularDestinations(context.Background(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "không thể lấy dữ liệu",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    destinations,
	})
}

// Lấy thông tin điểm đến theo ID
// @Summary Lấy thông tin điểm đến theo ID
// @Description Lấy thông tin chi tiết điểm đến theo ID
// @Tags Destination
// @Accept json
// @Produce json
// @Param id path int true "Destination ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/:id [get]
func (s *Server) GetDestinationByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	destination, err := s.z.GetDestinationByID(context.Background(), int32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Không tìm thấy điểm đến",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    destination,
	})
}

// Lấy danh sách tours theo điểm đến
// @Summary Lấy danh sách tours theo điểm đến
// @Description Lấy danh sách tours của một điểm đến cụ thể
// @Tags Destination
// @Accept json
// @Produce json
// @Param id path int true "Destination ID"
// @Param limit query int false "Số lượng tours" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/:id/tours [get]
func (s *Server) GetToursByDestination(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Sử dụng SearchTours với diem_den_id
	diemDenID := int32(id)
	req := db.SearchToursParams{
		Limit:     int32(limit),
		Offset:    int32(offset),
		DiemDenID: &diemDenID,
	}

	tours, err := s.z.SearchTours(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "không thể lấy dữ liệu",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    tours,
	})
}

// Cập nhật hình ảnh cho điểm đến
// @Summary Cập nhật hình ảnh cho điểm đến
// @Description Cập nhật hình ảnh cho điểm đến (cần quyền admin)
// @Tags Destination
// @Accept json
// @Produce json
// @Param id path int true "Destination ID"
// @Param image body object true "Image URL"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/:id/image [put]
func (s *Server) UpdateDestinationImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	var req struct {
		Anh string `json:"anh" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ", "details": err.Error()})
		return
	}

	destination, err := s.z.UpdateDestinationImage(context.Background(), db.UpdateDestinationImageParams{
		ID:  int32(id),
		Anh: &req.Anh,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "không thể cập nhật hình ảnh",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật hình ảnh thành công",
		"data":    destination,
	})
}
