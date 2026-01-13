package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"boilerplate-be/internal/config"
	"boilerplate-be/internal/shared/enum"
	"boilerplate-be/internal/shared/errors"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the go-redis client with custom methods
type RedisClient struct {
	*redis.Client
}

// NewRedis creates a new Redis client connection
func NewRedis(cfg config.RedisConfig) (*RedisClient, error) {
	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{Client: rdb}, nil
}

func (c *RedisClient) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisClient) GetValue(ctx context.Context, key string) (string, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *RedisClient) DeleteKey(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.Client.Exists(ctx, key).Result()
	return result > 0, err
}

func (c *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

func (c *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.Client.Expire(ctx, key, expiration).Err()
}

func (c *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}

func (c *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.Client.Keys(ctx, pattern).Result()
}

func (c *RedisClient) Close() error {
	return c.Client.Close()
}

// RedisHelper provides high-level Redis operations with error handling
type RedisHelper struct {
	client *RedisClient
}

// NewRedisHelper creates a new RedisHelper instance
func NewRedisHelper(client *RedisClient) *RedisHelper {
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

func (rh *RedisHelper) Start(ctx context.Context, key string, expiration time.Duration) error {
	return rh.Expire(ctx, key, expiration)
}

func (rh *RedisHelper) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := rh.client.Expire(ctx, key, expiration)
	return rh.handleRedisError(err, errors.CacheStoreFailed)
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

	return errors.Wrap(err, errorCode)
}
