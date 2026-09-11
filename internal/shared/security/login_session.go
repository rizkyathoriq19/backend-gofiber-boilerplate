package security

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"boilerplate-be/internal/database"
	"boilerplate-be/internal/shared/enum"

	"github.com/google/uuid"
)

var (
	ErrInvalidSession          = errors.New("invalid login session")
	ErrSessionStoreUnavailable = errors.New("login session store unavailable")
)

type SessionStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	CompareAndSwap(ctx context.Context, key, expected, replacement string, ttl time.Duration) (bool, error)
}

type redisSessionStore struct {
	client *database.RedisClient
}

func NewRedisSessionStore(client *database.RedisClient) SessionStore {
	return &redisSessionStore{client: client}
}

func (s *redisSessionStore) Get(ctx context.Context, key string) (string, error) {
	return s.client.GetValue(ctx, key)
}

func (s *redisSessionStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.client.SetWithTTL(ctx, key, value, ttl)
}

func (s *redisSessionStore) Delete(ctx context.Context, key string) error {
	return s.client.DeleteKey(ctx, key)
}

func (s *redisSessionStore) CompareAndSwap(ctx context.Context, key, expected, replacement string, ttl time.Duration) (bool, error) {
	result, err := s.client.Client.Eval(ctx, `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			redis.call("SET", KEYS[1], ARGV[2], "PX", ARGV[3])
			return 1
		end
		return 0
	`, []string{key}, expected, replacement, strconv.FormatInt(ttl.Milliseconds(), 10)).Int()
	return result == 1, err
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type LoginSessionManager struct {
	jwt           *JWTManager
	store         SessionStore
	refreshExpiry time.Duration
}

func NewLoginSessionManager(jwtManager *JWTManager, store SessionStore) *LoginSessionManager {
	return &LoginSessionManager{jwt: jwtManager, store: store, refreshExpiry: jwtManager.refreshExpiry}
}

func (m *LoginSessionManager) Issue(userID string, email string, role enum.UserRole) (TokenPair, error) {
	sessionID := uuid.New().String()
	accessToken, refreshToken, err := m.jwt.generateTokenPair(userID, email, role, sessionID)
	if err != nil {
		return TokenPair{}, err
	}

	refreshClaims, err := m.jwt.ValidateToken(refreshToken)
	if err != nil {
		return TokenPair{}, err
	}
	if err := m.withStore(func(ctx context.Context) error {
		return m.store.Set(ctx, m.sessionKey(sessionID), refreshClaims.ID, m.refreshExpiry)
	}); err != nil {
		return TokenPair{}, fmt.Errorf("%w: %v", ErrSessionStoreUnavailable, err)
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (m *LoginSessionManager) ValidateAccess(token string) (*Claims, error) {
	return m.validateSessionToken(token, "access")
}

func (m *LoginSessionManager) ValidateRefresh(token string) (*Claims, error) {
	claims, err := m.jwt.ValidateToken(token)
	if err != nil {
		return nil, ErrInvalidSession
	}
	return m.validateSessionClaims(claims, "refresh")
}

func (m *LoginSessionManager) validateSessionToken(token, tokenType string) (*Claims, error) {
	claims, err := m.jwt.ValidateToken(token)
	if err != nil {
		return nil, ErrInvalidSession
	}
	return m.validateSessionClaims(claims, tokenType)
}

func (m *LoginSessionManager) validateSessionClaims(claims *Claims, tokenType string) (*Claims, error) {
	if claims.TokenType != tokenType || claims.SessionID == "" {
		return nil, ErrInvalidSession
	}

	if err := m.withStore(func(ctx context.Context) error {
		value, err := m.store.Get(ctx, m.sessionKey(claims.SessionID))
		if err != nil {
			return fmt.Errorf("%w: %v", ErrSessionStoreUnavailable, err)
		}
		if value == "" || (tokenType == "refresh" && value != claims.ID) {
			return ErrInvalidSession
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return claims, nil
}

func (m *LoginSessionManager) Renew(token, userID, email string, role enum.UserRole) (TokenPair, error) {
	claims, err := m.ValidateRefresh(token)
	if err != nil || claims.UserID != userID {
		return TokenPair{}, ErrInvalidSession
	}

	accessToken, refreshToken, err := m.jwt.generateTokenPair(userID, email, role, claims.SessionID)
	if err != nil {
		return TokenPair{}, err
	}
	newClaims, err := m.jwt.ValidateToken(refreshToken)
	if err != nil {
		return TokenPair{}, err
	}

	var swapped bool
	if err := m.withStore(func(ctx context.Context) error {
		var err error
		swapped, err = m.store.CompareAndSwap(ctx, m.sessionKey(claims.SessionID), claims.ID, newClaims.ID, m.refreshExpiry)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrSessionStoreUnavailable, err)
		}
		return nil
	}); err != nil {
		return TokenPair{}, err
	}
	if !swapped {
		return TokenPair{}, ErrInvalidSession
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (m *LoginSessionManager) Logout(sessionID string) error {
	if sessionID == "" {
		return ErrInvalidSession
	}
	return m.withStore(func(ctx context.Context) error {
		if err := m.store.Delete(ctx, m.sessionKey(sessionID)); err != nil {
			return fmt.Errorf("%w: %v", ErrSessionStoreUnavailable, err)
		}
		return nil
	})
}

func (m *LoginSessionManager) withStore(fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return fn(ctx)
}

func (m *LoginSessionManager) sessionKey(sessionID string) string {
	return "login_session:" + sessionID
}
