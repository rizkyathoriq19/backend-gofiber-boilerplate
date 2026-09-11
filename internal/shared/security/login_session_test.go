package security

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"boilerplate-be/internal/config"
	"boilerplate-be/internal/database"
	"boilerplate-be/internal/shared/enum"
)

type fakeSessionStore struct {
	mu   sync.Mutex
	data map[string]string
	err  error
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{data: make(map[string]string)}
}

func (s *fakeSessionStore) Get(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return "", s.err
	}
	return s.data[key], nil
}

func (s *fakeSessionStore) Set(_ context.Context, key, value string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.data[key] = value
	return nil
}

func (s *fakeSessionStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	delete(s.data, key)
	return nil
}

func (s *fakeSessionStore) CompareAndSwap(_ context.Context, key, expected, replacement string, _ time.Duration) (bool, error) {
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

func TestLoginSessionManager_IsolatesSessionsAndLogout(t *testing.T) {
	store := newFakeSessionStore()
	manager := NewLoginSessionManager(NewJWTManager("secret", time.Hour, time.Hour), store)

	first, err := manager.Issue("user-1", "one@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatal(err)
	}
	second, err := manager.Issue("user-1", "one@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := manager.ValidateAccess(first.AccessToken); err != nil {
		t.Fatalf("first session should be active: %v", err)
	}
	if _, err := manager.ValidateAccess(second.AccessToken); err != nil {
		t.Fatalf("second session should be active: %v", err)
	}

	firstClaims, err := manager.ValidateAccess(first.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Logout(firstClaims.SessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ValidateAccess(first.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("logged out session error = %v, want ErrInvalidSession", err)
	}
	if _, err := manager.ValidateAccess(second.AccessToken); err != nil {
		t.Fatalf("logout must not affect another session: %v", err)
	}
}

func TestLoginSessionManager_RenewalIsSingleUse(t *testing.T) {
	store := newFakeSessionStore()
	manager := NewLoginSessionManager(NewJWTManager("secret", time.Hour, time.Hour), store)

	pair, err := manager.Issue("user-1", "one@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatal(err)
	}

	renewed, err := manager.Renew(pair.RefreshToken, "user-1", "one@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ValidateAccess(renewed.AccessToken); err != nil {
		t.Fatalf("renewed session should remain active: %v", err)
	}
	if _, err := manager.Renew(pair.RefreshToken, "user-1", "one@example.com", enum.UserRoleUser); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("reused refresh token error = %v, want ErrInvalidSession", err)
	}
	if _, err := manager.ValidateAccess(renewed.AccessToken); err != nil {
		t.Fatalf("rejected reuse must not revoke winning session: %v", err)
	}
}

func TestLoginSessionManager_RenewalAllowsOnlyOneConcurrentExchange(t *testing.T) {
	store := newFakeSessionStore()
	manager := NewLoginSessionManager(NewJWTManager("secret", time.Hour, time.Hour), store)

	pair, err := manager.Issue("user-1", "one@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := manager.Renew(pair.RefreshToken, "user-1", "one@example.com", enum.UserRoleUser)
			results <- err
		}()
	}

	var successes int
	for range 2 {
		if err := <-results; err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful renewals = %d, want 1", successes)
	}
}

func TestLoginSessionManager_FailsClosedWhenStoreUnavailable(t *testing.T) {
	store := newFakeSessionStore()
	manager := NewLoginSessionManager(NewJWTManager("secret", time.Hour, time.Hour), store)
	pair, err := manager.Issue("user-1", "one@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatal(err)
	}

	store.err = errors.New("redis unavailable")
	if _, err := manager.ValidateAccess(pair.AccessToken); !errors.Is(err, ErrSessionStoreUnavailable) {
		t.Fatalf("validation error = %v, want ErrSessionStoreUnavailable", err)
	}
}

func TestLoginSessionManager_RejectsLegacySessionlessToken(t *testing.T) {
	store := newFakeSessionStore()
	jwtManager := NewJWTManager("secret", time.Hour, time.Hour)
	manager := NewLoginSessionManager(jwtManager, store)
	legacyToken, err := jwtManager.generateToken("user-1", "one@example.com", enum.UserRoleUser, "access", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := manager.ValidateAccess(legacyToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("legacy token error = %v, want ErrInvalidSession", err)
	}
}

func TestRedisSessionManager_RenewalIsAtomic(t *testing.T) {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	client, err := database.NewRedis(config.RedisConfig{Host: host, Port: port, DB: 15})
	if err != nil {
		t.Skipf("Redis unavailable: %v", err)
	}
	defer client.Close()

	manager := NewLoginSessionManager(NewJWTManager("secret", time.Hour, time.Hour), NewRedisSessionStore(client))
	pair, err := manager.Issue("user-1", "one@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := manager.Renew(pair.RefreshToken, "user-1", "one@example.com", enum.UserRoleUser)
			results <- err
		}()
	}

	var successes int
	for range 2 {
		if <-results == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful renewals = %d, want 1", successes)
	}
}
