package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	СacheTTL = 15 * time.Minute
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	// Delete(ctx context.Context, key string) error
	Delete(ctx context.Context, pattern string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, bytes, expiration).Err()
}

// func (r *RedisCache) Delete(ctx context.Context, key string) error {
// 	return r.client.Del(ctx, key).Err()
// }

func (r *RedisCache) Delete(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		if err := r.client.Del(ctx, key).Err(); err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}

	return nil
}