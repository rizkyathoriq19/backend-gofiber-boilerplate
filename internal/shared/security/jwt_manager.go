package security

import (
	"time"

	"boilerplate-be/internal/shared/enum"

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
	SessionID string        `json:"session_id"`
	jwt.RegisteredClaims
}

func NewJWTManager(secretKey string, expiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		expiry:        expiry,
		refreshExpiry: refreshExpiry,
	}
}

func (j *JWTManager) generateTokenPair(userID string, email string, role enum.UserRole, sessionID string) (string, string, error) {
	// Generate access token
	accessToken, err := j.generateToken(userID, email, role, "access", sessionID, j.expiry)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := j.generateToken(userID, email, role, "refresh", sessionID, j.refreshExpiry)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (j *JWTManager) generateToken(userID string, email string, role enum.UserRole, tokenType, sessionID string, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		SessionID: sessionID,
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
