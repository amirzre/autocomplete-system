package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/amirzre/autocomplete-system/internal/config"
	"github.com/amirzre/autocomplete-system/internal/model"
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

// CacheStats represents cache statistics.
type CacheStats struct {
	Hits     int64   `json:"hits"`
	Misses   int64   `json:"misses"`
	HitRatio float64 `json:"hit_ratio"`
	Size     int     `json:"size"`
	MaxSize  int     `json:"max_size"`
}

// CacheInterface defines the contract for cache operations.
type CacheInterface interface {
	Get(key string) ([]model.Suggestion, bool)
	Set(key string, value []model.Suggestion)
	Delete(key string)
	Clear()
	Close() error
	GetStats() CacheStats
}

// NewRedisCache creates a new Redis cache instance.
func NewRedisCache(config *config.Config) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Cache.Host, config.Cache.Port),
		Password:     config.Cache.Password,
		DB:           config.Cache.DB,
		PoolSize:     config.Cache.PoolSize,
		MinIdleConns: config.Cache.MinIdleConns,
		MaxRetries:   config.Cache.MaxRetries,
		DialTimeout:  config.Cache.DialTimeout,
		ReadTimeout:  config.Cache.ReadTimeout,
		WriteTimeout: config.Cache.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client:  rdb,
		ttl:     config.Cache.TTL,
		maxSize: config.Cache.MaxSize,
	}, nil
}

// Get retrieves a value from the cache.
func (rc *RedisCache) Get(key string) ([]model.Suggestion, bool) {
	ctx := context.Background()

	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			atomic.AddInt64(&rc.misses, 1)
			return nil, false
		}

		atomic.AddInt64(&rc.misses, 1)
		return nil, false
	}

	var suggestions []model.Suggestion
	if err := json.Unmarshal([]byte(val), &suggestions); err != nil {
		atomic.AddInt64(&rc.misses, 1)
		return nil, false
	}

	atomic.AddInt64(&rc.hits, 1)
	return suggestions, true
}

// Set stores a value in the cache.
func (rc *RedisCache) Set(key string, value []model.Suggestion) {
	ctx := context.Background()

	if rc.maxSize > 0 {
		size, _ := rc.client.DBSize(ctx).Result()
		if size >= int64(rc.maxSize) {
			keys, _ := rc.client.RandomKey(ctx).Result()
			if keys != "" {
				rc.client.Del(ctx, keys)
			}
		}
	}

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	rc.client.Set(ctx, key, data, rc.ttl)
}

// Delete removes a value from the cache.
func (rc *RedisCache) Delete(key string) {
	ctx := context.Background()
	rc.client.Del(ctx, key)
}

// Clear removes all items from the cache.
func (rc *RedisCache) Clear() {
	ctx := context.Background()
	rc.client.FlushDB(ctx)
	atomic.StoreInt64(&rc.hits, 0)
	atomic.StoreInt64(&rc.misses, 0)
}

// Close closes the Redis connection.
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// GetStats returns cache statistics.
func (rc *RedisCache) GetStats() CacheStats {
	ctx := context.Background()

	hits := atomic.LoadInt64(&rc.hits)
	misses := atomic.LoadInt64(&rc.misses)
	total := hits + misses

	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(hits) / float64(total)
	}

	size := 0
	if dbSize, err := rc.client.DBSize(ctx).Result(); err == nil {
		size = int(dbSize)
	}

	return CacheStats{
		Hits:     hits,
		Misses:   misses,
		HitRatio: hitRatio,
		Size:     size,
		MaxSize:  rc.maxSize,
	}
}
