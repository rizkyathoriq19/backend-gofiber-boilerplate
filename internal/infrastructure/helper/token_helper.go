package helper

import (
	"context"
	"fmt"
	"time"

	"boilerplate-be/internal/infrastructure/redis"
)

type TokenManager struct {
	*RedisHelper
	keyPrefix string
	ttl       time.Duration
}

type TokenManagerConfig struct {
	KeyPrefix string
	TTL       time.Duration
}

func NewTokenManager(client *redis.Client) *TokenManager {
	return NewTokenManagerWithConfig(client, TokenManagerConfig{
		KeyPrefix: "refresh_token",
		TTL:       168 * time.Hour, // 7 days
	})
}

func NewTokenManagerWithConfig(client *redis.Client, config TokenManagerConfig) *TokenManager {
	return &TokenManager{
		RedisHelper: NewRedisHelper(client),
		keyPrefix:   config.KeyPrefix,
		ttl:         config.TTL,
	}
}

func (tm *TokenManager) StoreToken(userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := tm.buildTokenKey(userID, tokenID)
	return tm.SetWithTTL(ctx, key, "1", tm.ttl)
}

func (tm *TokenManager) ValidateToken(userID, tokenID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := tm.buildTokenKey(userID, tokenID)
	return tm.Exists(ctx, key)
}

func (tm *TokenManager) BlacklistToken(userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := tm.buildTokenKey(userID, tokenID)
	return tm.Delete(ctx, key)
}

func (tm *TokenManager) RevokeToken(userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := tm.buildTokenKey(userID, tokenID)
	return tm.Delete(ctx, key)
}

func (tm *TokenManager) RevokeAllUserTokens(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pattern := tm.buildTokenKey(userID, "*")
	keys, err := tm.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil // no tokens to revoke
	}

	return tm.Delete(ctx, keys...)
}

func (tm *TokenManager) GetUserTokenCount(userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pattern := tm.buildTokenKey(userID, "*")
	keys, err := tm.Keys(ctx, pattern)
	if err != nil {
		return 0, err
	}

	return len(keys), nil
}

func (tm *TokenManager) ExtendTokenTTL(userID, tokenID string, newTTL time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := tm.buildTokenKey(userID, tokenID)
	
	// Check if token exists first
	exists, err := tm.Exists(ctx, key)
	if err != nil {
		return err
	}
	
	if !exists {
		return fmt.Errorf("token not found")
	}

	// Extend TTL
	return tm.RedisHelper.client.Expire(ctx, key, newTTL)
}

func (tm *TokenManager) buildTokenKey(userID, tokenID string) string {
	return fmt.Sprintf("%s:%s:%s", tm.keyPrefix, userID, tokenID)
}

func (tm *TokenManager) SetTTL(ttl time.Duration) {
	tm.ttl = ttl
}

func (tm *TokenManager) GetTTL() time.Duration {
	return tm.ttl
}