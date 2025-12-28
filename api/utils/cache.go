package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheHelper provides utility functions for Redis caching
type CacheHelper struct {
	redis *redis.Client
}

// NewCacheHelper creates a new cache helper
func NewCacheHelper(redisClient *redis.Client) *CacheHelper {
	return &CacheHelper{
		redis: redisClient,
	}
}

// Get retrieves a value from cache and unmarshals it into result
func (c *CacheHelper) Get(ctx context.Context, key string, result interface{}) error {
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), result)
}

// Set stores a value in cache with expiration
func (c *CacheHelper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, expiration).Err()
}

// Delete removes a key from cache
func (c *CacheHelper) Delete(ctx context.Context, keys ...string) error {
	return c.redis.Del(ctx, keys...).Err()
}

// DeletePattern deletes all keys matching a pattern
func (c *CacheHelper) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.redis.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.redis.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *CacheHelper) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.redis.Exists(ctx, key).Result()
	return count > 0, err
}

// TTL returns the remaining time to live of a key
func (c *CacheHelper) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.redis.TTL(ctx, key).Result()
}

// CacheKey generates a consistent cache key
func CacheKey(prefix string, parts ...interface{}) string {
	key := prefix
	for _, part := range parts {
		key += fmt.Sprintf(":%v", part)
	}
	return key
}

// Cache key prefixes
const (
	CachePrefixTour         = "tour"
	CachePrefixTourList     = "tour:list"
	CachePrefixTourDetail   = "tour:detail"
	CachePrefixTourCategory = "tour:category"

	CachePrefixDestination       = "destination"
	CachePrefixDestinationList   = "destination:list"
	CachePrefixDestinationDetail = "destination:detail"

	CachePrefixSupplier       = "supplier"
	CachePrefixSupplierList   = "supplier:list"
	CachePrefixSupplierDetail = "supplier:detail"

	CachePrefixAdmin      = "admin"
	CachePrefixAdminStats = "admin:stats"

	CachePrefixPayment  = "payment"
	CachePrefixLocation = "location"
)

// Thời gian cache mặc định
const (
	// Cache ngắn hạn (1 giờ)
	CacheDurationShort = 1 * time.Hour

	// Cache trung hạn (6 giờ)
	CacheDurationMedium = 6 * time.Hour

	// Cache dài hạn (24 giờ)
	CacheDurationLong = 24 * time.Hour

	// Cache rất dài hạn (7 ngày) cho dữ liệu ít thay đổi
	CacheDurationVeryLong = 7 * 24 * time.Hour

	// Cache admin stats (15 phút)
	CacheDurationAdminStats = 15 * time.Minute

	// Cache list (30 phút)
	CacheDurationList = 30 * time.Minute

	// Cache detail (2 giờ)
	CacheDurationDetail = 2 * time.Hour
)

// GetOrSet lấy từ cache hoặc tính toán và lưu vào cache
func (c *CacheHelper) GetOrSet(
	ctx context.Context,
	key string,
	expiration time.Duration,
	result interface{},
	fetchFn func() (interface{}, error),
) error {
	// Thử lấy từ cache
	err := c.Get(ctx, key, result)
	if err == nil {
		// Cache hit
		return nil
	}

	// Lỗi khi lấy từ cache
	data, err := fetchFn()
	if err != nil {
		return err
	}

	// Lưu vào cache
	if err := c.Set(ctx, key, data, expiration); err != nil {
		// Log error nhưng không fail request
		fmt.Printf("Failed to cache data: %v\n", err)
	}

	// Marshal data vào result
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, result)
}

// InvalidateTourCache hủy tất cả cache tour-related
func (c *CacheHelper) InvalidateTourCache(ctx context.Context, tourID ...int) error {
	patterns := []string{
		CachePrefixTourList + ":*",
		CachePrefixTourCategory + ":*",
	}

	// Nếu có Id tour cụ thể, hủy cache cho chúng
	for _, id := range tourID {
		patterns = append(patterns, CacheKey(CachePrefixTourDetail, id))
	}

	for _, pattern := range patterns {
		if err := c.DeletePattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}

// 
func (c *CacheHelper) InvalidateDestinationCache(ctx context.Context, destinationID ...int) error {
	patterns := []string{
		CachePrefixDestinationList + ":*",
	}

	// Nếu có Id destination cụ thể, hủy cache cho chúng
	for _, id := range destinationID {
		patterns = append(patterns, CacheKey(CachePrefixDestinationDetail, id))
	}

	for _, pattern := range patterns {
		if err := c.DeletePattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}

// InvalidateSupplierCache hủy tất cả cache supplier-related
func (c *CacheHelper) InvalidateSupplierCache(ctx context.Context, supplierID ...int) error {
	patterns := []string{
		CachePrefixSupplierList + ":*",
	}

	// Nếu có Id supplier cụ thể, hủy cache cho chúng
	for _, id := range supplierID {
		patterns = append(patterns, CacheKey(CachePrefixSupplierDetail, id))
	}

	for _, pattern := range patterns {
		if err := c.DeletePattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}

// InvalidateAdminStatsCache hủy tất cả cache admin statistics
func (c *CacheHelper) InvalidateAdminStatsCache(ctx context.Context) error {
	return c.DeletePattern(ctx, CachePrefixAdminStats+":*")
}

// BatchGet lấy nhiều keys cùng lúc
func (c *CacheHelper) BatchGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	pipe := c.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	results := make(map[string]interface{})
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err == nil {
			var data interface{}
			if err := json.Unmarshal([]byte(val), &data); err == nil {
				results[keys[i]] = data
			}
		}
	}

	return results, nil
}

// BatchSet lưu nhiều key-value cùng lúc
func (c *CacheHelper) BatchSet(ctx context.Context, items map[string]interface{}, expiration time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := c.redis.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		pipe.Set(ctx, key, data, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// IncrementCounter tăng counter trong cache
func (c *CacheHelper) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	count, err := c.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiration chỉ khi đây là key mới
	if count == 1 {
		c.redis.Expire(ctx, key, expiration)
	}

	return count, nil
}

// GetCacheStats lấy thống kê cache
func (c *CacheHelper) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.redis.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	dbSize, err := c.redis.DBSize(ctx).Result()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"db_size": dbSize,
		"info":    info,
	}, nil
}
