package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

// IpApiResponse represents the response from ip-api.com
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
func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (most common with proxies/load balancers)
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// We want the first one (original client)
		ips := strings.Split(xForwardedFor, ",")
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
		if ip := net.ParseIP(xRealIP); ip != nil {
			return xRealIP
		}
	}

	// Check CF-Connecting-IP (Cloudflare)
	cfConnectingIP := c.GetHeader("CF-Connecting-IP")
	if cfConnectingIP != "" {
		if ip := net.ParseIP(cfConnectingIP); ip != nil {
			return cfConnectingIP
		}
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
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
