package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepo interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expire time.Duration) error
	Delete(ctx context.Context, key string) error
	Update(ctx context.Context, key string, value string, expire time.Duration) error
}
type cacheRepo struct {
	redis *redis.Client
}

func NewCacheRepo(redis *redis.Client) CacheRepo {
	return &cacheRepo{redis: redis}
}

// Keyが存在しない場合は値がnilを返す
func (r *cacheRepo) Get(ctx context.Context, key string) (string, error) {
	value, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (r *cacheRepo) Set(ctx context.Context, key string, value string, expire time.Duration) error {
	err := r.redis.Set(ctx, key, value, expire).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *cacheRepo) Delete(ctx context.Context, key string) error {
	err := r.redis.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *cacheRepo) Update(ctx context.Context, key string, value string, expire time.Duration) error {
	return r.Set(ctx, key, value, expire)
}
