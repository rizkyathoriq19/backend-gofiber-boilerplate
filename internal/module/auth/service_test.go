package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"boilerplate-be/internal/shared/enum"
	apperrors "boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/security"
)

type authSessionStore struct {
	mu   sync.Mutex
	data map[string]string
	err  error
}

func newAuthSessionStore() *authSessionStore {
	return &authSessionStore{data: make(map[string]string)}
}

func (s *authSessionStore) Get(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return "", s.err
	}
	return s.data[key], nil
}

func (s *authSessionStore) Set(_ context.Context, key, value string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.data[key] = value
	return nil
}

func (s *authSessionStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	delete(s.data, key)
	return nil
}

func (s *authSessionStore) CompareAndSwap(_ context.Context, key, expected, replacement string, _ time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return false, s.err
	}
	if s.data[key] != expected {
		return false, nil
	}
	s.data[key] = replacement
	return true, nil
}

func newSessionAuthUseCase(t *testing.T) (*authUseCase, *authSessionStore, *security.LoginSessionManager) {
	t.Helper()
	repo := &testAuthRepository{users: map[string]*User{}}
	password, err := security.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	repo.users["user-1"] = &User{ID: "user-1", Email: "user@example.com", Password: password, Role: enum.UserRoleUser}
	store := newAuthSessionStore()
	jwtManager := security.NewJWTManager("test-secret", time.Hour, time.Hour)
	manager := security.NewLoginSessionManager(jwtManager, store)
	return NewAuthUseCase(repo, manager), store, manager
}

func TestAuthUseCase_UsesProductionSessionLifecycle(t *testing.T) {
	useCase, _, manager := newSessionAuthUseCase(t)

	accessOne, refreshOne, err := useCase.Login("user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	accessTwo, refreshTwo, err := useCase.Login("user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}

	claimsOne, err := manager.ValidateAccess(accessOne)
	if err != nil {
		t.Fatal(err)
	}
	if err := useCase.Logout(claimsOne.SessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ValidateAccess(accessOne); !errors.Is(err, security.ErrInvalidSession) {
		t.Fatalf("logged out access error = %v, want ErrInvalidSession", err)
	}
	if _, err := manager.ValidateAccess(accessTwo); err != nil {
		t.Fatalf("logout affected another login session: %v", err)
	}

	renewedAccess, _, err := useCase.RefreshToken(refreshTwo)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := useCase.RefreshToken(refreshTwo); err == nil {
		t.Fatal("reusing a refresh token should fail")
	}
	if _, err := manager.ValidateAccess(renewedAccess); err != nil {
		t.Fatalf("renewed access error = %v", err)
	}
	if _, _, err := useCase.RefreshToken(refreshOne); err == nil {
		t.Fatal("refresh token from logged out session should fail")
	}
}

func TestAuthUseCase_MapsSessionStoreFailureToUnavailable(t *testing.T) {
	useCase, store, _ := newSessionAuthUseCase(t)
	_, _, err := useCase.Login("user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	store.err = errors.New("redis unavailable")

	_, _, err = useCase.Login("user@example.com", "password123")
	appErr, ok := apperrors.IsAppError(err)
	if !ok || appErr.Code != apperrors.AuthServiceUnavailable {
		t.Fatalf("error = %v, want AuthServiceUnavailable", err)
	}
}

type testAuthRepository struct {
	users map[string]*User
}

func (r *testAuthRepository) CreateUser(user *User) error {
	for _, existing := range r.users {
		if existing.Email == user.Email {
			return apperrors.New(apperrors.EmailExists)
		}
	}
	if user.ID == "" {
		user.ID = "new-user"
	}
	if user.Role == "" {
		user.Role = enum.UserRoleUser
	}
	r.users[user.ID] = user
	return nil
}

func (r *testAuthRepository) GetUserByEmail(email string) (*User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, apperrors.New(apperrors.AccountNotFound)
}

func (r *testAuthRepository) GetUserByID(id string) (*User, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, apperrors.New(apperrors.AccountNotFound)
}

func (r *testAuthRepository) UpdateUser(user *User) error {
	if _, ok := r.users[user.ID]; !ok {
		return apperrors.New(apperrors.AccountNotFound)
	}
	r.users[user.ID] = user
	return nil
}
