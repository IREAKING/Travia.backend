package middleware

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"travia.backend/api/utils"
)

// CacheMiddleware caches GET request responses
func CacheMiddleware(redis *redis.Client, duration time.Duration) gin.HandlerFunc {
	cache := utils.NewCacheHelper(redis)

	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Generate cache key from URL and query params
		cacheKey := generateCacheKey(c)

		// Try to get from cache
		var cachedResponse CachedResponse
		err := cache.Get(c.Request.Context(), cacheKey, &cachedResponse)
		if err == nil {
			// Cache hit - return cached response
			c.Header("X-Cache", "HIT")
			c.Header("X-Cache-Key", cacheKey)
			c.Data(cachedResponse.StatusCode, cachedResponse.ContentType, cachedResponse.Body)
			c.Abort()
			return
		}

		// Cache miss - capture response
		c.Header("X-Cache", "MISS")
		c.Header("X-Cache-Key", cacheKey)

		// Create response writer wrapper
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Only cache successful responses
		if c.Writer.Status() == http.StatusOK {
			response := CachedResponse{
				StatusCode:  c.Writer.Status(),
				ContentType: c.Writer.Header().Get("Content-Type"),
				Body:        writer.body.Bytes(),
				CachedAt:    time.Now(),
			}

			// Store in cache (fire and forget) using background context to avoid cancellation
			go func(key string, resp CachedResponse, ttl time.Duration) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = cache.Set(ctx, key, resp, ttl)
			}(cacheKey, response, duration)
		}
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode  int       `json:"status_code"`
	ContentType string    `json:"content_type"`
	Body        []byte    `json:"body"`
	CachedAt    time.Time `json:"cached_at"`
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// generateCacheKey creates a unique cache key from request
func generateCacheKey(c *gin.Context) string {
	// Include method, path, and query params
	key := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL.Path)

	// Add query params if present
	if c.Request.URL.RawQuery != "" {
		key += "?" + c.Request.URL.RawQuery
	}

	// Hash the key if it's too long
	if len(key) > 200 {
		hash := md5.Sum([]byte(key))
		return "cache:http:" + hex.EncodeToString(hash[:])
	}

	return "cache:http:" + key
}

// CacheBustMiddleware adds cache busting headers for no-cache endpoints
func CacheBustMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// ConditionalCacheMiddleware caches only if certain conditions are met
func ConditionalCacheMiddleware(redis *redis.Client, duration time.Duration, condition func(*gin.Context) bool) gin.HandlerFunc {
	cacheMiddleware := CacheMiddleware(redis, duration)

	return func(c *gin.Context) {
		if condition(c) {
			cacheMiddleware(c)
		} else {
			c.Next()
		}
	}
}

// InvalidateCacheMiddleware provides cache invalidation for write operations
func InvalidateCacheMiddleware(redis *redis.Client, patterns ...string) gin.HandlerFunc {
	cache := utils.NewCacheHelper(redis)

	return func(c *gin.Context) {
		// Process request first
		c.Next()

		// Only invalidate on successful write operations
		if c.Request.Method != http.MethodGet && c.Writer.Status() < 400 {
			for _, pattern := range patterns {
				go func(p string) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					_ = cache.DeletePattern(ctx, p)
				}(pattern)
			}
		}
	}
}

// RateLimitMiddleware provides rate limiting using Redis
func RateLimitMiddleware(redis *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", clientIP)

		// Increment counter
		count, err := redis.Incr(c.Request.Context(), key).Result()
		if err != nil {
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			redis.Expire(c.Request.Context(), key, window)
		}

		// Check if limit exceeded
		if count > int64(limit) {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-int(count)))

		c.Next()
	}
}

// CacheWarmupMiddleware pre-warms cache for common queries
func CacheWarmupMiddleware(redis *redis.Client, urls []string, duration time.Duration) error {
	client := &http.Client{Timeout: 10 * time.Second}
	cache := utils.NewCacheHelper(redis)

	for _, url := range urls {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			cacheKey := "cache:warmup:" + url
			response := CachedResponse{
				StatusCode:  resp.StatusCode,
				ContentType: resp.Header.Get("Content-Type"),
				Body:        body,
				CachedAt:    time.Now(),
			}

			// Use background context for cache warmup
			cache.Set(context.Background(), cacheKey, response, duration)
		}
	}

	return nil
}
