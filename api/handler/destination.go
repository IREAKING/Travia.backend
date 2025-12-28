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

	// Convert []*string to simple string array for frontend
	var countries []string
	for _, countryName := range countryNames {
		if countryName != nil {
			countries = append(countries, *countryName)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    countries,
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

	// Convert []*string to []db.DiemDen format for frontend
	var provinces []db.DiemDen
	for i, provinceName := range provinceNames {
		if provinceName != nil {
			provinces = append(provinces, db.DiemDen{
				ID:      int32(i + 1),  // Generate temporary ID
				Ten:     *provinceName, // Use province name as ten
				Tinh:    provinceName,  // Set tinh
				QuocGia: &country,      // Set quoc_gia
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    provinces,
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

	// Convert []string to []db.DiemDen format for frontend
	var cities []db.DiemDen
	for i, cityName := range cityNames {
		cities = append(cities, db.DiemDen{
			ID:   int32(i + 1), // Generate temporary ID
			Ten:  cityName,     // Use city name as ten
			Tinh: &province,    // Set tinh
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu thành công",
		"data":    cities,
	})
}

// Lấy điểm đến theo ID
// @Summary Lấy điểm đến theo ID
// @Description Lấy điểm đến theo ID
// @Tags Destination
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} db.DiemDen
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/{id} [get]
func (s *Server) GetDestinationByID(c *gin.Context) {
	_id := c.Param("id")
	id, err := strconv.Atoi(_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}
	data, err := s.z.GetDestinationByID(context.Background(), int32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Destination fetched successfully", "data": data})
}

// Lấy tất cả điểm đến
// @Summary Lấy tất cả điểm đến
// @Description Lấy tất cả điểm đến
// @Tags Destination
// @Accept json
// @Produce json
// @Success 200 {object} db.DiemDen
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/getAllDestination [get]
func (s *Server) GetAllDestinations(c *gin.Context) {
	data, err := s.z.GetAllDestinations(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All destinations fetched successfully", "data": data})
}

// Lấy tất cả điểm đến theo cấu trúc phân cấp
// @Summary Lấy tất cả điểm đến theo cấu trúc phân cấp
// @Description Lấy tất cả điểm đến được nhóm theo quốc gia, tỉnh thành và thành phố
// @Tags Destination
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/hierarchical [get]
func (s *Server) GetDestinationsHierarchical(c *gin.Context) {
	// Lấy tất cả destinations
	destinations, err := s.z.GetAllDestinations(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy dữ liệu điểm đến",
			"message": err.Error(),
		})
		return
	}

	// Tạo cấu trúc phân cấp
	hierarchicalData := make(map[string]map[string][]db.DiemDen)

	for _, dest := range destinations {
		country := ""
		province := ""

		if dest.QuocGia != nil {
			country = *dest.QuocGia
		}
		if dest.Tinh != nil {
			province = *dest.Tinh
		}

		// Khởi tạo country nếu chưa có
		if hierarchicalData[country] == nil {
			hierarchicalData[country] = make(map[string][]db.DiemDen)
		}

		// Thêm destination vào province
		hierarchicalData[country][province] = append(hierarchicalData[country][province], dest)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu điểm đến phân cấp thành công",
		"data":    hierarchicalData,
	})
}

// Lấy điểm đến với pagination và search cho create tour
// @Summary Lấy điểm đến với pagination và search cho create tour
// @Description Lấy điểm đến với pagination và search, tối ưu cho create tour
// @Tags Destination
// @Accept json
// @Produce json
// @Param limit query int false "Limit (default: 20)"
// @Param offset query int false "Offset (default: 0)"
// @Param search query string false "Search term"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /destination/for-tour-creation [get]
func (s *Server) GetDestinationsForTourCreation(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	search := c.Query("search")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	var destinations []db.DiemDen
	var total int32

	if search != "" && len(search) >= 2 {
		// Sử dụng search endpoint
		searchResults, err := s.z.SearchDestinations(context.Background(), &search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Không thể tìm kiếm điểm đến",
				"message": err.Error(),
			})
			return
		}
		destinations = searchResults
		total = int32(len(searchResults))
	} else {
		// Sử dụng pagination endpoint
		paginatedResults, err := s.z.GetDestinationsWithPagination(context.Background(), db.GetDestinationsWithPaginationParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Không thể lấy dữ liệu điểm đến",
				"message": err.Error(),
			})
			return
		}
		destinations = paginatedResults

		// Lấy total count
		countResult, err := s.z.CountDestinations(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Không thể đếm số lượng điểm đến",
				"message": err.Error(),
			})
			return
		}
		total = countResult
	}

	// Format response cho frontend
	response := gin.H{
		"destinations": destinations,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
		"has_more":     int32(offset+limit) < total,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy dữ liệu điểm đến thành công",
		"data":    response,
	})
}
