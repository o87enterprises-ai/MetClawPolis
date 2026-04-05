package api

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

// InitRedis connects to Redis and verifies connectivity
func InitRedis() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return err
	}
	opts.PoolSize = 10
	opts.MinIdleConns = 2
	opts.ConnMaxIdleTime = 5 * time.Minute

	RDB = redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return err
	}

	log.Println("Redis connected at", redisURL)
	return nil
}

// PublishEvent publishes an event to a Redis channel for real-time broadcasting
func PublishEvent(channel string, payload string) error {
	if RDB == nil {
		return nil // Redis not available
	}
	return RDB.Publish(context.Background(), channel, payload).Err()
}

// Subscribe returns a Redis Pub/Sub subscription
func Subscribe(channels ...string) *redis.PubSub {
	if RDB == nil {
		return nil
	}
	return RDB.Subscribe(context.Background(), channels...)
}

// CacheSet stores a value in Redis with TTL
func CacheSet(key string, value string, ttl time.Duration) error {
	if RDB == nil {
		return nil
	}
	return RDB.Set(context.Background(), key, value, ttl).Err()
}

// CacheGet retrieves a value from Redis
func CacheGet(key string) (string, error) {
	if RDB == nil {
		return "", nil
	}
	return RDB.Get(context.Background(), key).Result()
}

// CacheDelete removes a key from Redis
func CacheDelete(key string) error {
	if RDB == nil {
		return nil
	}
	return RDB.Del(context.Background(), key).Err()
}
