package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/amirzre/autocomplete-system/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisCache represents a Redis-based cache with TTL support.
type RedisCache struct {
	client  *redis.Client
	ttl     time.Duration
	hits    int64
	misses  int64
	maxSize int
}

// CacheInterface defines the contract for cache operations.
type CacheInterface interface{}

// NewRedisCache creates a new Redis cache instance.
func NewRedisCache(config *config.CacheConfig) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client:  rdb,
		ttl:     config.TTL,
		maxSize: config.MaxSize,
	}, nil
}
