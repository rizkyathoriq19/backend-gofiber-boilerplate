package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"boilerplate-be/internal/infrastructure/redis"
)

type CacheHelper struct {
	*RedisHelper
	defaultTTL time.Duration
}

func NewCacheHelper(client *redis.Client, defaultTTL time.Duration) *CacheHelper {
	if defaultTTL == 0 {
		defaultTTL = 1 * time.Hour // Default 1 hour
	}
	
	return &CacheHelper{
		RedisHelper: NewRedisHelper(client),
		defaultTTL:  defaultTTL,
	}
}

func (ch *CacheHelper) CacheJSON(ctx context.Context, key string, data interface{}, ttl ...time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	cacheTTL := ch.defaultTTL
	if len(ttl) > 0 {
		cacheTTL = ttl[0]
	}

	return ch.SetWithTTL(ctx, key, string(jsonData), cacheTTL)
}

func (ch *CacheHelper) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := ch.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (ch *CacheHelper) GetOrSet(ctx context.Context, key string, fetchFn func() (interface{}, error), ttl ...time.Duration) (interface{}, error) {
	// Try to get from cache first
	var cachedData interface{}
	if err := ch.GetJSON(ctx, key, &cachedData); err == nil {
		return cachedData, nil
	}

	// Cache miss, fetch data
	data, err := fetchFn()
	if err != nil {
		return nil, err
	}

	// Cache the result
	cacheTTL := ch.defaultTTL
	if len(ttl) > 0 {
		cacheTTL = ttl[0]
	}

	if cacheErr := ch.CacheJSON(ctx, key, data, cacheTTL); cacheErr != nil {
		// Log error but don't fail the request
		// You can integrate with your logging system here
	}

	return data, nil
}

func (ch *CacheHelper) InvalidatePattern(ctx context.Context, pattern string) error {
	keys, err := ch.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return ch.Delete(ctx, keys...)
}

func (ch *CacheHelper) BuildUserCacheKey(userID, keyType string) string {
	return fmt.Sprintf("user:%s:%s", userID, keyType)
}

func (ch *CacheHelper) BuildEntityCacheKey(entityType, entityID, keyType string) string {
	return fmt.Sprintf("%s:%s:%s", entityType, entityID, keyType)
}

func (ch *CacheHelper) CacheUserData(ctx context.Context, userID, keyType string, data interface{}, ttl ...time.Duration) error {
	key := ch.BuildUserCacheKey(userID, keyType)
	return ch.CacheJSON(ctx, key, data, ttl...)
}

func (ch *CacheHelper) GetUserData(ctx context.Context, userID, keyType string, dest interface{}) error {
	key := ch.BuildUserCacheKey(userID, keyType)
	return ch.GetJSON(ctx, key, dest)
}

func (ch *CacheHelper) InvalidateUserCache(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("user:%s:*", userID)
	return ch.InvalidatePattern(ctx, pattern)
}

func (ch *CacheHelper) DeleteKey(ctx context.Context, key string) error {
	return ch.Delete(ctx, key)
}