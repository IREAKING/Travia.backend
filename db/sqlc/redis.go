package db

import (
	"time"

	"github.com/redis/go-redis/v9"
	"travia.backend/config"
)

func InitRedis(config *config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Network:      "tcp",
		Addr:         config.Address,
		ClientName:   "travia-redis",
		DB:           config.DB,
		Username:     config.Username,
		Password:     config.Password,
		PoolSize:     10,              // số connection tối đa
		MinIdleConns: 1,               // số connection tối thiểu
		MaxIdleConns: 10,              // số connection tối đa khi không có request
		DialTimeout:  5 * time.Second, // thời gian kết nối tối đa
		ReadTimeout:  3 * time.Second, // thời gian đọc tối đa
		WriteTimeout: 3 * time.Second, // thời gian ghi tối đa
		MaxRetries:   3,               // số lần retry khi lỗi
	})
}
