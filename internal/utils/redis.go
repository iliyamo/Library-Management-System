package utils

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	RedisCtx    = context.Background()
)

// InitRedis initializes a Redis client connection using environment variables.
// Make sure REDIS_ADDR and REDIS_PASSWORD are set in .env
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),     // e.g. "localhost:6379"
		Password: os.Getenv("REDIS_PASSWORD"), // leave empty if no password
		DB:       0,
	})

	if err := RedisClient.Ping(RedisCtx).Err(); err != nil {
		log.Fatalf("❌ Redis connection failed: %v", err)
	}
	log.Println("✅ Redis connected")
}

// SetCache sets a key-value pair in Redis with expiration TTL.
func SetCache(key string, value string, ttl time.Duration) error {
	return RedisClient.Set(RedisCtx, key, value, ttl).Err()
}

// GetCache retrieves the value of a key from Redis.
func GetCache(key string) (string, error) {
	return RedisClient.Get(RedisCtx, key).Result()
}

// DeleteCache deletes a key from Redis (useful after data changes).
func DeleteCache(key string) error {
	return RedisClient.Del(RedisCtx, key).Err()
}
