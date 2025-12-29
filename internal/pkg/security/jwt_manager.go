package security

import (
	"context"
	"fmt"
	"time"

	"boilerplate-be/internal/database/redis"
	"boilerplate-be/internal/pkg/enum"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	secretKey     string
	expiry        time.Duration
	refreshExpiry time.Duration
}

type Claims struct {
	UserID    string        `json:"user_id"`
	Email     string        `json:"email"`
	Role      enum.UserRole `json:"role"`
	TokenType string        `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func NewJWTManager(secretKey string, expiry time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		expiry:        expiry,
		refreshExpiry: 168 * time.Hour, // 7 days default
	}
}

func (j *JWTManager) SetRefreshExpiry(expiry time.Duration) {
	j.refreshExpiry = expiry
}

func (j *JWTManager) GenerateTokenPair(userID string, email string, role enum.UserRole) (string, string, error) {
	// Generate access token
	accessToken, err := j.generateToken(userID, email, role, "access", j.expiry)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := j.generateToken(userID, email, role, "refresh", j.refreshExpiry)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (j *JWTManager) GenerateToken(userID string, email string, role enum.UserRole) (string, error) {
	return j.generateToken(userID, email, role, "access", j.expiry)
}

func (j *JWTManager) generateToken(userID string, email string, role enum.UserRole, tokenType string, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func (j *JWTManager) BlacklistToken(ctx context.Context, redisClient *redis.Client, tokenID string, expiry time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	return redisClient.SetWithTTL(ctx, key, "1", expiry)
}

func (j *JWTManager) IsTokenBlacklisted(ctx context.Context, redisClient *redis.Client, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	return redisClient.Exists(ctx, key)
}