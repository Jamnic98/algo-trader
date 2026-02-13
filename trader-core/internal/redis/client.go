package redis

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func NewClient() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379/0"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Invalid REDIS_URL:", err)
	}

	rdb := redis.NewClient(opt)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	log.Println("Connected to Redis!")
	return rdb
}

func (c *Client) Set(ctx context.Context, key string, value any) error {
	return c.rdb.Set(ctx, key, value, 0).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}
