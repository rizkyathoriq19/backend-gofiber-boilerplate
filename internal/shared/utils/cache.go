package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gojson "github.com/goccy/go-json"

	"boilerplate-be/internal/database"
)

type CacheHelper struct {
	*database.RedisHelper
	defaultTTL time.Duration
}

func NewCacheHelper(client *database.RedisClient, defaultTTL time.Duration) *CacheHelper {
	if defaultTTL == 0 {
		defaultTTL = 1 * time.Hour // Default 1 hour
	}

	return &CacheHelper{
		RedisHelper: database.NewRedisHelper(client),
		defaultTTL:  defaultTTL,
	}
}

// CacheJSON caches data as JSON with optional TTL
func (ch *CacheHelper) CacheJSON(ctx context.Context, key string, data interface{}, ttl ...time.Duration) error {
	jsonData, err := gojson.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	cacheTTL := ch.defaultTTL
	if len(ttl) > 0 {
		cacheTTL = ttl[0]
	}

	return ch.SetWithTTL(ctx, key, string(jsonData), cacheTTL)
}

// GetJSON retrieves and unmarshals JSON data from cache
func (ch *CacheHelper) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := ch.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == "" {
		return fmt.Errorf("cache miss: key not found")
	}

	return json.Unmarshal([]byte(data), dest)
}

// GetOrSetTyped is a generic cache-aside pattern implementation
// It tries to get data from cache first, if not found, calls fetchFn and caches the result
func GetOrSetTyped[T any](ch *CacheHelper, ctx context.Context, key string, fetchFn func() (T, error), ttl ...time.Duration) (T, error) {
	var result T

	// Try to get from cache first
	data, err := ch.Get(ctx, key)
	if err == nil && data != "" {
		// Cache hit - unmarshal the data
		if err := gojson.Unmarshal([]byte(data), &result); err == nil {
			return result, nil
		}
		// If unmarshal fails, fall through to fetch
	}

	// Cache miss, fetch data
	result, err = fetchFn()
	if err != nil {
		return result, err
	}

	// Cache the result
	cacheTTL := ch.defaultTTL
	if len(ttl) > 0 {
		cacheTTL = ttl[0]
	}

	// Cache in background to not block the response
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = ch.CacheJSON(cacheCtx, key, result, cacheTTL)
	}()

	return result, nil
}

// GetOrSet maintains backward compatibility but uses interface{} return
// For type-safe operations, use GetOrSetTyped instead
func (ch *CacheHelper) GetOrSet(ctx context.Context, key string, fetchFn func() (interface{}, error), ttl ...time.Duration) (interface{}, error) {
	// Try to get from cache first
	data, err := ch.Get(ctx, key)
	if err == nil && data != "" {
		// Data found - return as raw JSON for later processing
		var result interface{}
		if err := gojson.Unmarshal([]byte(data), &result); err == nil {
			return result, nil
		}
	}

	// Cache miss, fetch data
	result, err := fetchFn()
	if err != nil {
		return nil, err
	}

	// Cache the result
	cacheTTL := ch.defaultTTL
	if len(ttl) > 0 {
		cacheTTL = ttl[0]
	}

	// Cache error is intentionally ignored - don't fail the request due to cache issues
	_ = ch.CacheJSON(ctx, key, result, cacheTTL)

	return result, nil
}

// InvalidatePattern removes all keys matching the pattern
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

// BuildUserCacheKey creates a standardized cache key for user data
func (ch *CacheHelper) BuildUserCacheKey(userID, keyType string) string {
	return fmt.Sprintf("user:%s:%s", userID, keyType)
}

// BuildEntityCacheKey creates a standardized cache key for any entity
func (ch *CacheHelper) BuildEntityCacheKey(entityType, entityID, keyType string) string {
	return fmt.Sprintf("%s:%s:%s", entityType, entityID, keyType)
}

// CacheUserData caches user-specific data
func (ch *CacheHelper) CacheUserData(ctx context.Context, userID, keyType string, data interface{}, ttl ...time.Duration) error {
	key := ch.BuildUserCacheKey(userID, keyType)
	return ch.CacheJSON(ctx, key, data, ttl...)
}

// GetUserData retrieves user-specific data from cache
func (ch *CacheHelper) GetUserData(ctx context.Context, userID, keyType string, dest interface{}) error {
	key := ch.BuildUserCacheKey(userID, keyType)
	return ch.GetJSON(ctx, key, dest)
}

// InvalidateUserCache removes all cached data for a user
func (ch *CacheHelper) InvalidateUserCache(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("user:%s:*", userID)
	return ch.InvalidatePattern(ctx, pattern)
}

// DeleteKey removes a specific key from cache
func (ch *CacheHelper) DeleteKey(ctx context.Context, key string) error {
	return ch.Delete(ctx, key)
}
