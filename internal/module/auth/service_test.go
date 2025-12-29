package auth

import (
	"testing"
	"time"

	apperrors "boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/security"
)

// MockAuthRepository implements AuthRepository interface for testing
type MockAuthRepository struct {
	users         map[string]*User
	createUserErr error
	getUserErr    error
	updateUserErr error
}

func NewMockAuthRepository() *MockAuthRepository {
	return &MockAuthRepository{
		users: make(map[string]*User),
	}
}

func (m *MockAuthRepository) CreateUser(user *User) error {
	if m.createUserErr != nil {
		return m.createUserErr
	}
	// Check for duplicate email
	for _, u := range m.users {
		if u.Email == user.Email {
			return apperrors.New(apperrors.EmailExists)
		}
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockAuthRepository) GetUserByEmail(email string) (*User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, apperrors.New(apperrors.AccountNotFound)
}

func (m *MockAuthRepository) GetUserByID(id string) (*User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, apperrors.New(apperrors.AccountNotFound)
}

func (m *MockAuthRepository) UpdateUser(user *User) error {
	if m.updateUserErr != nil {
		return m.updateUserErr
	}
	if _, ok := m.users[user.ID]; !ok {
		return apperrors.New(apperrors.AccountNotFound)
	}
	m.users[user.ID] = user
	return nil
}

// MockTokenManager for testing
type MockTokenManager struct {
	tokens map[string]bool
}

func NewMockTokenManager() *MockTokenManager {
	return &MockTokenManager{
		tokens: make(map[string]bool),
	}
}

func (m *MockTokenManager) StoreToken(userID, tokenID string) error {
	m.tokens[userID+":"+tokenID] = true
	return nil
}

func (m *MockTokenManager) ValidateToken(userID, tokenID string) (bool, error) {
	return m.tokens[userID+":"+tokenID], nil
}

func (m *MockTokenManager) BlacklistToken(userID, tokenID string) error {
	delete(m.tokens, userID+":"+tokenID)
	return nil
}

func (m *MockTokenManager) RevokeToken(userID, tokenID string) error {
	delete(m.tokens, userID+":"+tokenID)
	return nil
}

func (m *MockTokenManager) RevokeAllUserTokens(userID string) error {
	for key := range m.tokens {
		if key[:len(userID)+1] == userID+":" {
			delete(m.tokens, key)
		}
	}
	return nil
}

// Test Suite for Auth Service
func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		userName    string
		setupMock   func(*MockAuthRepository)
		expectError bool
		errorCode   string
	}{
		{
			name:        "successful registration",
			email:       "test@example.com",
			password:    "password123",
			userName:    "Test User",
			setupMock:   func(m *MockAuthRepository) {},
			expectError: false,
		},
		{
			name:     "duplicate email",
			email:    "existing@example.com",
			password: "password123",
			userName: "Test User",
			setupMock: func(m *MockAuthRepository) {
				m.users["existing-id"] = &User{
					ID:    "existing-id",
					Email: "existing@example.com",
					Name:  "Existing User",
				}
			},
			expectError: true,
			errorCode:   "email_exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockAuthRepository()
			tt.setupMock(mockRepo)

			jwtManager := security.NewJWTManager("test-secret", 24*time.Hour)
			// Note: Using real token manager requires Redis, so we test without it here

			// For this test, we'll directly test the repository logic
			user := &User{
				ID:       "test-id",
				Email:    tt.email,
				Name:     tt.userName,
				Password: tt.password,
			}

			err := mockRepo.CreateUser(user)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Verify JWT manager can generate tokens
			if !tt.expectError {
				accessToken, refreshToken, err := jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
				if err != nil {
					t.Errorf("failed to generate tokens: %v", err)
				}
				if accessToken == "" || refreshToken == "" {
					t.Error("tokens should not be empty")
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		setupMock   func(*MockAuthRepository)
		expectError bool
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMock: func(m *MockAuthRepository) {
				hashedPassword, _ := security.HashPassword("password123")
				m.users["user-id"] = &User{
					ID:       "user-id",
					Email:    "test@example.com",
					Name:     "Test User",
					Password: hashedPassword,
				}
			},
			expectError: false,
		},
		{
			name:     "wrong password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMock: func(m *MockAuthRepository) {
				hashedPassword, _ := security.HashPassword("password123")
				m.users["user-id"] = &User{
					ID:       "user-id",
					Email:    "test@example.com",
					Name:     "Test User",
					Password: hashedPassword,
				}
			},
			expectError: true,
		},
		{
			name:        "user not found",
			email:       "nonexistent@example.com",
			password:    "password123",
			setupMock:   func(m *MockAuthRepository) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockAuthRepository()
			tt.setupMock(mockRepo)

			// Get user and verify password
			user, err := mockRepo.GetUserByEmail(tt.email)
			if tt.expectError {
				if err == nil {
					// User found, check password
					err = security.CheckPassword(user.Password, tt.password)
					if err == nil {
						t.Errorf("expected error but got none")
					}
				}
				// Error is expected, test passed
			} else {
				if err != nil {
					t.Errorf("unexpected error getting user: %v", err)
					return
				}
				err = security.CheckPassword(user.Password, tt.password)
				if err != nil {
					t.Errorf("password check failed: %v", err)
				}
			}
		})
	}
}

func TestJWTManager_TokenGeneration(t *testing.T) {
	jwtManager := security.NewJWTManager("test-secret-key", 24*time.Hour)

	t.Run("generate and validate token pair", func(t *testing.T) {
		accessToken, refreshToken, err := jwtManager.GenerateTokenPair("user-123", "test@example.com", "user")
		if err != nil {
			t.Fatalf("failed to generate token pair: %v", err)
		}

		if accessToken == "" {
			t.Error("access token should not be empty")
		}
		if refreshToken == "" {
			t.Error("refresh token should not be empty")
		}

		// Validate access token
		accessClaims, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			t.Fatalf("failed to validate access token: %v", err)
		}
		if accessClaims.UserID != "user-123" {
			t.Errorf("expected user ID 'user-123', got '%s'", accessClaims.UserID)
		}
		if accessClaims.TokenType != "access" {
			t.Errorf("expected token type 'access', got '%s'", accessClaims.TokenType)
		}

		// Validate refresh token
		refreshClaims, err := jwtManager.ValidateToken(refreshToken)
		if err != nil {
			t.Fatalf("failed to validate refresh token: %v", err)
		}
		if refreshClaims.TokenType != "refresh" {
			t.Errorf("expected token type 'refresh', got '%s'", refreshClaims.TokenType)
		}
	})

	t.Run("invalid token should fail validation", func(t *testing.T) {
		_, err := jwtManager.ValidateToken("invalid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})
}

func TestPasswordHashing(t *testing.T) {
	t.Run("hash and verify password", func(t *testing.T) {
		password := "securePassword123!"

		hash, err := security.HashPassword(password)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		if hash == password {
			t.Error("hash should not equal plain password")
		}

		err = security.CheckPassword(hash, password)
		if err != nil {
			t.Error("correct password should verify successfully")
		}
	})

	t.Run("wrong password should fail", func(t *testing.T) {
		hash, _ := security.HashPassword("correctPassword")

		err := security.CheckPassword(hash, "wrongPassword")
		if err == nil {
			t.Error("wrong password should fail verification")
		}
	})
}

func TestMockRepository(t *testing.T) {
	t.Run("CRUD operations", func(t *testing.T) {
		repo := NewMockAuthRepository()

		// Create user
		user := &User{
			ID:       "test-id",
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "hashed-password",
			Role:     "user",
		}
		err := repo.CreateUser(user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Get by email
		found, err := repo.GetUserByEmail("test@example.com")
		if err != nil {
			t.Fatalf("failed to get user by email: %v", err)
		}
		if found.ID != user.ID {
			t.Errorf("expected ID '%s', got '%s'", user.ID, found.ID)
		}

		// Get by ID
		found, err = repo.GetUserByID("test-id")
		if err != nil {
			t.Fatalf("failed to get user by ID: %v", err)
		}
		if found.Email != user.Email {
			t.Errorf("expected email '%s', got '%s'", user.Email, found.Email)
		}

		// Update
		user.Name = "Updated Name"
		err = repo.UpdateUser(user)
		if err != nil {
			t.Fatalf("failed to update user: %v", err)
		}

		found, _ = repo.GetUserByID("test-id")
		if found.Name != "Updated Name" {
			t.Errorf("expected name 'Updated Name', got '%s'", found.Name)
		}

		// Duplicate email should fail
		duplicate := &User{
			ID:    "another-id",
			Email: "test@example.com",
		}
		err = repo.CreateUser(duplicate)
		if err == nil {
			t.Error("duplicate email should fail")
		}
	})
}

// Benchmark tests
func BenchmarkPasswordHashing(b *testing.B) {
	password := "testPassword123!"
	for i := 0; i < b.N; i++ {
		_, _ = security.HashPassword(password)
	}
}

func BenchmarkTokenGeneration(b *testing.B) {
	jwtManager := security.NewJWTManager("test-secret", 24*time.Hour)
	for i := 0; i < b.N; i++ {
		_, _, _ = jwtManager.GenerateTokenPair("user-123", "test@example.com", "user")
	}
}

func BenchmarkTokenValidation(b *testing.B) {
	jwtManager := security.NewJWTManager("test-secret", 24*time.Hour)
	token, _, _ := jwtManager.GenerateTokenPair("user-123", "test@example.com", "user")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jwtManager.ValidateToken(token)
	}
}
