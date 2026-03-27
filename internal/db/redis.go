package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// MustConnectRedis returns a Redis client from REDIS_URL or exits.
func MustConnectRedis() *redis.Client {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(fmt.Sprintf("redis url: %v", err))
	}
	c := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("redis ping: %v", err))
	}
	return c
}
