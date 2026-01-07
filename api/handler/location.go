package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	db "travia.backend/db/sqlc"
)

// LocationResponse represents the location data response
type LocationResponse struct {
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	RegionCode  string  `json:"region_code"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	Currency    string  `json:"currency"`
	Languages   string  `json:"languages"`
	CachedAt    string  `json:"cached_at,omitempty"`
}

// IpApiResponse thể hiện phản hồi từ ip-api.com.
type IpApiResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"regionName"`
	RegionCode  string  `json:"region"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Currency    string  `json:"currency"`
	ISP         string  `json:"isp"`
	Query       string  `json:"query"` // IP address
	Message     string  `json:"message"`
}

// IpapiResponse represents the response from ipapi.co (backup)
type IpapiResponse struct {
	IP            string  `json:"ip"`
	City          string  `json:"city"`
	Region        string  `json:"region"`
	RegionCode    string  `json:"region_code"`
	Country       string  `json:"country_name"`
	CountryCode   string  `json:"country_code"`
	ContinentCode string  `json:"continent_code"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Timezone      string  `json:"timezone"`
	Currency      string  `json:"currency"`
	Languages     string  `json:"languages"`
	Error         bool    `json:"error"`
	Reason        string  `json:"reason"`
}

// GetClientIP extracts the real client IP from the request
// Handles X-Forwarded-For, X-Real-IP, and other proxy headers
// When behind VPN/proxy, prioritizes headers that contain the real client IP
func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (most common with proxies/load balancers)
	// Format: "client_ip, proxy1_ip, proxy2_ip"
	// For VPN: We want the LAST IP (the one closest to the client) if it's public,
	// otherwise the first non-private IP
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		// Try to find the first public IP (likely the real client behind VPN)
		for i := len(ips) - 1; i >= 0; i-- {
			clientIP := strings.TrimSpace(ips[i])
			if ip := net.ParseIP(clientIP); ip != nil {
				// If it's a public IP, use it (this is likely the VPN exit IP)
				if !isPrivateIP(clientIP) {
					return clientIP
				}
			}
		}
		// If no public IP found, use the first one
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if ip := net.ParseIP(clientIP); ip != nil {
				return clientIP
			}
		}
	}

	// Check X-Real-IP header
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		xRealIP = strings.TrimSpace(xRealIP)
		if ip := net.ParseIP(xRealIP); ip != nil {
			return xRealIP
		}
	}

	// Check CF-Connecting-IP (Cloudflare)
	cfConnectingIP := c.GetHeader("CF-Connecting-IP")
	if cfConnectingIP != "" {
		cfConnectingIP = strings.TrimSpace(cfConnectingIP)
		if ip := net.ParseIP(cfConnectingIP); ip != nil {
			return cfConnectingIP
		}
	}

	// Check True-Client-IP (used by some proxies)
	trueClientIP := c.GetHeader("True-Client-IP")
	if trueClientIP != "" {
		trueClientIP = strings.TrimSpace(trueClientIP)
		if ip := net.ParseIP(trueClientIP); ip != nil {
			return trueClientIP
		}
	}

	// Check X-Client-IP (some proxies use this)
	xClientIP := c.GetHeader("X-Client-IP")
	if xClientIP != "" {
		xClientIP = strings.TrimSpace(xClientIP)
		if ip := net.ParseIP(xClientIP); ip != nil {
			return xClientIP
		}
	}

	// Fallback to RemoteAddr
	// This will be the IP that directly connected to the server
	// If behind VPN, this should be the VPN server's IP
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		// If RemoteAddr doesn't have a port, try parsing it directly
		if parsedIP := net.ParseIP(c.Request.RemoteAddr); parsedIP != nil {
			return c.Request.RemoteAddr
		}
		return c.Request.RemoteAddr
	}
	return ip
}

// isPrivateIP checks if the IP is a private/local IP
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check for localhost
	if parsedIP.IsLoopback() {
		return true
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	}

	for _, cidr := range privateRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// fetchFromIpApi fetches location data from ip-api.com (primary, free unlimited)
func fetchFromIpApi(ip string) (*LocationResponse, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,countryCode,region,regionName,city,lat,lon,timezone,currency,isp,query", ip)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ip-api.com request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ip-api.com returned status code: %d", resp.StatusCode)
	}

	var apiResp IpApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode ip-api.com response: %w", err)
	}

	// Check if API returned an error
	if apiResp.Status == "fail" {
		return nil, fmt.Errorf("ip-api.com error: %s", apiResp.Message)
	}

	// Determine language based on country
	languages := "en"
	if apiResp.CountryCode == "VN" {
		languages = "vi"
	}

	// Convert to our response format
	location := &LocationResponse{
		IP:          apiResp.Query,
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		Region:      apiResp.Region,
		RegionCode:  apiResp.RegionCode,
		City:        apiResp.City,
		Latitude:    apiResp.Lat,
		Longitude:   apiResp.Lon,
		Timezone:    apiResp.Timezone,
		Currency:    apiResp.Currency,
		Languages:   languages,
	}

	return location, nil
}

// fetchFromIpApiCo fetches location data from ipapi.co (backup)
func fetchFromIpApiCo(ip string) (*LocationResponse, error) {
	url := fmt.Sprintf("https://ipapi.co/%s/json/", ip)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ipapi.co request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ipapi.co returned status code: %d", resp.StatusCode)
	}

	var apiResp IpapiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode ipapi.co response: %w", err)
	}

	// Check if API returned an error
	if apiResp.Error {
		return nil, fmt.Errorf("ipapi.co error: %s", apiResp.Reason)
	}

	// Convert to our response format
	location := &LocationResponse{
		IP:          apiResp.IP,
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		Region:      apiResp.Region,
		RegionCode:  apiResp.RegionCode,
		City:        apiResp.City,
		Latitude:    apiResp.Latitude,
		Longitude:   apiResp.Longitude,
		Timezone:    apiResp.Timezone,
		Currency:    apiResp.Currency,
		Languages:   apiResp.Languages,
	}

	return location, nil
}

// fetchLocationFromAPI fetches location data with multi-provider fallback
func fetchLocationFromAPI(ip string) (*LocationResponse, error) {
	// Try primary API: ip-api.com (free unlimited)
	location, err := fetchFromIpApi(ip)
	if err == nil {
		return location, nil
	}

	// Fallback to ipapi.co if primary fails
	location, err2 := fetchFromIpApiCo(ip)
	if err2 == nil {
		return location, nil
	}

	// Both APIs failed
	return nil, fmt.Errorf("all providers failed - primary: %v, backup: %v", err, err2)
}

// GetLocation godoc
// @Summary Lấy thông tin vị trí địa lý của người dùng
// @Description Phát hiện quốc gia và thông tin địa lý dựa trên địa chỉ IP của người dùng. Kết quả được cache trong 24 giờ.
// @Tags Location
// @Accept json
// @Produce json
// @Param ip query string false "IP address (tùy chọn, nếu không có sẽ tự động detect)"
// @Success 200 {object} LocationResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /location [get]
func (s *Server) GetLocation(c *gin.Context) {
	// Get IP from query param or auto-detect
	ip := c.Query("ip")
	if ip == "" {
		ip = GetClientIP(c)
	}

	// Validate IP
	if net.ParseIP(ip) == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid IP address",
		})
		return
	}

	// Handle private/local IPs
	if isPrivateIP(ip) {
		c.JSON(http.StatusOK, gin.H{
			"ip":           ip,
			"country":      "Vietnam",
			"country_code": "VN",
			"city":         "Local",
			"message":      "Private IP detected, returning default location",
		})
		return
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("location:%s", ip)

	// Try to get from Redis cache
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var location LocationResponse
		if err := json.Unmarshal([]byte(cachedData), &location); err == nil {
			location.CachedAt = "from_cache"
			c.JSON(http.StatusOK, location)
			return
		}
	}

	// Fetch from API
	location, err := fetchLocationFromAPI(ip)
	if err != nil {
		// Fallback: try to get country from Accept-Language header
		acceptLang := c.GetHeader("Accept-Language")
		fallbackCountry := "VN" // Default to Vietnam
		fallbackCountryName := "Vietnam"

		if strings.Contains(acceptLang, "en") {
			fallbackCountry = "US"
			fallbackCountryName = "United States"
		} else if strings.Contains(acceptLang, "vi") {
			fallbackCountry = "VN"
			fallbackCountryName = "Vietnam"
		}

		c.JSON(http.StatusOK, gin.H{
			"ip":           ip,
			"country":      fallbackCountryName,
			"country_code": fallbackCountry,
			"message":      "Location service unavailable, using fallback",
			"error":        err.Error(),
		})
		return
	}

	// Cache the result for 24 hours
	locationJSON, err := json.Marshal(location)
	if err == nil {
		s.redis.Set(ctx, cacheKey, locationJSON, 24*time.Hour)
	}

	c.JSON(http.StatusOK, location)
}

// GetLocationByIP godoc
// @Summary Lấy thông tin vị trí địa lý theo IP cụ thể
// @Description Phát hiện quốc gia và thông tin địa lý dựa trên địa chỉ IP được cung cấp
// @Tags Location
// @Accept json
// @Produce json
// @Param ip path string true "IP Address"
// @Success 200 {object} LocationResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /location/{ip} [get]
func (s *Server) GetLocationByIP(c *gin.Context) {
	ip := c.Param("ip")

	// Validate IP
	if net.ParseIP(ip) == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid IP address",
		})
		return
	}

	// Handle private/local IPs
	if isPrivateIP(ip) {
		c.JSON(http.StatusOK, gin.H{
			"ip":           ip,
			"country":      "Vietnam",
			"country_code": "VN",
			"city":         "Local",
			"message":      "Private IP detected, returning default location",
		})
		return
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("location:%s", ip)

	// Try to get from Redis cache
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var location LocationResponse
		if err := json.Unmarshal([]byte(cachedData), &location); err == nil {
			location.CachedAt = "from_cache"
			c.JSON(http.StatusOK, location)
			return
		}
	}

	// Fetch from API
	location, err := fetchLocationFromAPI(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch location data",
			"details": err.Error(),
		})
		return
	}

	// Cache the result for 24 hours
	locationJSON, err := json.Marshal(location)
	if err == nil {
		s.redis.Set(ctx, cacheKey, locationJSON, 24*time.Hour)
	}

	c.JSON(http.StatusOK, location)
}

// GetClientIPDebug godoc
// @Summary Debug endpoint để xem IP và headers được detect
// @Description Trả về thông tin chi tiết về IP và các headers liên quan để debug
// @Tags Location
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /location/debug [get]
func (s *Server) GetClientIPDebug(c *gin.Context) {
	detectedIP := GetClientIP(c)
	remoteAddr := c.Request.RemoteAddr

	// Get all relevant headers
	headers := gin.H{
		"X-Forwarded-For":  c.GetHeader("X-Forwarded-For"),
		"X-Real-IP":        c.GetHeader("X-Real-IP"),
		"CF-Connecting-IP": c.GetHeader("CF-Connecting-IP"),
		"True-Client-IP":   c.GetHeader("True-Client-IP"),
		"X-Client-IP":      c.GetHeader("X-Client-IP"),
		"RemoteAddr":       remoteAddr,
		"DetectedIP":       detectedIP,
		"IsPrivateIP":      isPrivateIP(detectedIP),
	}

	// Try to get location for detected IP
	var locationInfo gin.H
	if !isPrivateIP(detectedIP) {
		loc, err := fetchLocationFromAPI(detectedIP)
		if err == nil {
			locationInfo = gin.H{
				"country":      loc.Country,
				"country_code": loc.CountryCode,
				"city":         loc.City,
				"ip":           loc.IP,
			}
		} else {
			locationInfo = gin.H{
				"error": err.Error(),
			}
		}
	} else {
		locationInfo = gin.H{
			"message": "Private IP detected, cannot fetch location",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"headers":     headers,
		"location":    locationInfo,
		"all_headers": c.Request.Header,
	})
}

// GetToursByLocation godoc
// @Summary Lấy danh sách tour quốc nội và quốc tế dựa vào vị trí người dùng
// @Description Tự động phát hiện vị trí người dùng và trả về danh sách tour quốc nội (trong nước) và quốc tế (nước ngoài) sắp xếp theo số lượt đặt nhiều nhất
// @Tags Location
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng tour mỗi loại (mặc định: 10)"
// @Param offset query int false "Offset (mặc định: 0)"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /location/tours [get]
func (s *Server) GetToursByLocation(c *gin.Context) {
	// Get IP and detect location
	ip := GetClientIP(c)
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Không thể xác định địa chỉ IP",
		})
		return
	}

	// Get location info
	var countryCode string
	if isPrivateIP(ip) {
		// Default to Vietnam for private IPs
		countryCode = "VN"
	} else {
		location, err := fetchLocationFromAPI(ip)
		if err != nil {
			// Fallback to Vietnam if API fails
			countryCode = "VN"
		} else {
			countryCode = location.CountryCode
		}
	}

	// Get limit and offset
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	ctx := context.Background()

	// Get domestic tours (quốc nội)
	domesticTours, err := s.z.GetToursByCountryCode(ctx, db.GetToursByCountryCodeParams{
		Column1: "domestic",
		Iso2:    &countryCode,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy danh sách tour quốc nội",
			"message": err.Error(),
		})
		return
	}

	// Get international tours (quốc tế)
	internationalTours, err := s.z.GetToursByCountryCode(ctx, db.GetToursByCountryCodeParams{
		Column1: "international",
		Iso2:    &countryCode,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Không thể lấy danh sách tour quốc tế",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh sách tour thành công",
		"data": gin.H{
			"country_code":   countryCode,
			"tours_quoc_noi": domesticTours,
			"tours_quoc_te":  internationalTours,
		},
	})
}

// Handler godoc
// @Summary Handler endpoint để xem IP và headers được detect
// @Description Trả về thông tin chi tiết về IP và các headers liên quan để debug
// @Tags Location
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /location/test [get]
func (s *Server) Handler(c *gin.Context) {
	c.JSON(200, gin.H{
		"client_ip":       c.ClientIP(),
		"x_forwarded_for": c.GetHeader("X-Forwarded-For"),
		"x_real_ip":       c.GetHeader("X-Real-IP"),
	})
}
