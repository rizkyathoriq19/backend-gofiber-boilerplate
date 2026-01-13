package security

import (
	"testing"
	"time"

	"boilerplate-be/internal/shared/enum"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-for-testing-purposes", 24*time.Hour)

	tests := []struct {
		name    string
		userID  string
		email   string
		role    enum.UserRole
		wantErr bool
	}{
		{
			name:    "Generate valid token",
			userID:  "user-123",
			email:   "test@example.com",
			role:    enum.UserRoleUser,
			wantErr: false,
		},
		{
			name:    "Generate token with admin role",
			userID:  "admin-123",
			email:   "admin@example.com",
			role:    enum.UserRoleAdmin,
			wantErr: false,
		},
		{
			name:    "Empty user ID",
			userID:  "",
			email:   "test@example.com",
			role:    enum.UserRoleUser,
			wantErr: false, // JWT allows empty claims
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtManager.GenerateToken(tt.userID, tt.email, tt.role)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-for-testing-purposes", 24*time.Hour)

	accessToken, refreshToken, err := jwtManager.GenerateTokenPair("user-123", "test@example.com", enum.UserRoleUser)

	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	if accessToken == "" {
		t.Error("GenerateTokenPair() returned empty access token")
	}

	if refreshToken == "" {
		t.Error("GenerateTokenPair() returned empty refresh token")
	}

	if accessToken == refreshToken {
		t.Error("GenerateTokenPair() access and refresh tokens should be different")
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-for-testing-purposes", 24*time.Hour)

	// Generate a valid token first
	validToken, err := jwtManager.GenerateToken("user-123", "test@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "Invalid token",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Malformed token",
			token:   "not-a-jwt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && claims == nil {
				t.Error("ValidateToken() returned nil claims for valid token")
			}

			if !tt.wantErr && claims.UserID != "user-123" {
				t.Errorf("ValidateToken() claims.UserID = %v, want %v", claims.UserID, "user-123")
			}
		})
	}
}

func TestJWTManager_TokenExpiry(t *testing.T) {
	// Create JWT manager with very short expiry
	jwtManager := NewJWTManager("test-secret", 1*time.Millisecond)

	token, err := jwtManager.GenerateToken("user-123", "test@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Token should be expired
	_, err = jwtManager.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should return error for expired token")
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Hash valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Hash empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "Hash long password",
			password: "this-is-a-very-long-password-that-should-still-work-correctly",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}

			if !tt.wantErr && hash == tt.password {
				t.Error("HashPassword() returned unhashed password")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "password123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongpassword",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password,
			hash:     "invalid-hash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.hash, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
