package database

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

func InitializeRedis(url string) *redis.Client {
	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(fmt.Sprintf("redis connection failed: %v", err))
	}

	return redis.NewClient(opt)
}
