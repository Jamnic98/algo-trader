package redis

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
			addr = "redis:6379"
	}

	rdb := redis.NewClient(&redis.Options{
			Addr: addr,
			DB:   0, // default DB
	})

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
			log.Fatal("Redis connection failed:", err)
	}

	log.Println("Connected to Redis!")
	return rdb
}
