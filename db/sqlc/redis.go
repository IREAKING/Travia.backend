package db

import (
	"github.com/redis/go-redis/v9"
	"travia.backend/config"
)

func InitRedis(config *config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Network:    "tcp",
		Addr:       config.Address,
		ClientName: "travia-redis",
		DB:         config.DB,
		Username:   config.Username,
		Password:   config.Password,
	})
}
