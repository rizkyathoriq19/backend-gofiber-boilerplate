package helper

import (
	"context"
	"time"

	"boilerplate-be/internal/infrastructure/enum"
	"boilerplate-be/internal/infrastructure/errors"
	"boilerplate-be/internal/infrastructure/redis"
)

type RedisHelper struct {
	client *redis.Client
}

func NewRedisHelper(client *redis.Client) *RedisHelper {
	return &RedisHelper{client: client}
}

func (rh *RedisHelper) SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	err := rh.client.SetWithTTL(ctx, key, value, ttl)
	return rh.handleRedisError(err, errors.CacheStoreFailed)
}

func (rh *RedisHelper) Get(ctx context.Context, key string) (string, error) {
	value, err := rh.client.GetValue(ctx, key)
	if err != nil {
		return "", rh.handleRedisError(err, errors.CacheRetrieveFailed)
	}
	return value, nil
}

func (rh *RedisHelper) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := rh.client.Exists(ctx, key)
	if err != nil {
		return false, rh.handleRedisError(err, errors.CacheRetrieveFailed)
	}
	return exists, nil
}

func (rh *RedisHelper) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		if err := rh.client.DeleteKey(ctx, key); err != nil {
			return rh.handleRedisError(err, errors.CacheDeleteFailed)
		}
	}
	return nil
}

func (rh *RedisHelper) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := rh.client.Keys(ctx, pattern)
	if err != nil {
		return nil, rh.handleRedisError(err, errors.CacheRetrieveFailed)
	}
	return keys, nil
}

func (rh *RedisHelper) Increment(ctx context.Context, key string) (int64, error) {
	value, err := rh.client.Incr(ctx, key)
	if err != nil {
		return 0, rh.handleRedisError(err, errors.CacheStoreFailed)
	}
	return value, nil
}

func (rh *RedisHelper) handleRedisError(err error, errorCode enum.ErrorCode) error {
	if err == nil {
		return nil
	}
	
	// Handle Redis-specific errors if needed
	return errors.Wrap(err, errorCode)
}