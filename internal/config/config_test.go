package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

var configEnvKeys = []string{
	"APP_NAME", "APP_ENV", "APP_PORT", "APP_PREFORK",
	"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSL_MODE",
	"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME",
	"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB", "REDIS_POOL_SIZE", "REDIS_DEFAULT_TTL",
	"JWT_SECRET", "JWT_EXPIRY", "JWT_REFRESH_EXPIRY",
	"RATE_LIMIT_MAX", "RATE_LIMIT_WINDOW",
	"CORS_ALLOWED_ORIGINS", "CORS_ALLOWED_METHODS", "CORS_ALLOWED_HEADERS",
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range configEnvKeys {
		value, wasSet := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
		key, value, wasSet := key, value, wasSet
		t.Cleanup(func() {
			if wasSet {
				_ = os.Setenv(key, value)
			} else {
				_ = os.Unsetenv(key)
			}
		})
	}
}

func TestNewUsesDefaultsAndValidOverrides(t *testing.T) {
	tests := []struct {
		name  string
		env   map[string]string
		check func(*testing.T, *Config)
	}{
		{
			name: "defaults when settings are absent",
			check: func(t *testing.T, cfg *Config) {
				if cfg.App.Name != "Go Fiber Auth API" || cfg.App.Env != "development" || cfg.App.Port != 3000 || cfg.App.Prefork {
					t.Fatalf("unexpected app defaults: %+v", cfg.App)
				}
				if cfg.Database.Port != 5432 || cfg.Database.MaxOpenConns != 25 || cfg.Database.MaxIdleConns != 25 || cfg.Database.ConnMaxLifetime != 5*time.Minute {
					t.Fatalf("unexpected database defaults: %+v", cfg.Database)
				}
				if cfg.Redis.Port != 6379 || cfg.Redis.DB != 0 || cfg.Redis.PoolSize != 10 || cfg.Redis.DefaultTTL != time.Hour {
					t.Fatalf("unexpected Redis defaults: %+v", cfg.Redis)
				}
				if cfg.JWT.Secret != "your-secret-key" || cfg.JWT.Expiry != 24*time.Hour || cfg.JWT.RefreshExpiry != 168*time.Hour {
					t.Fatalf("unexpected JWT defaults: %+v", cfg.JWT)
				}
				if cfg.RateLimit.Max != 100 || cfg.RateLimit.Window != time.Minute {
					t.Fatalf("unexpected rate-limit defaults: %+v", cfg.RateLimit)
				}
			},
		},
		{
			name: "valid overrides",
			env: map[string]string{
				"APP_ENV":              "Production",
				"APP_PORT":             "8443",
				"APP_PREFORK":          "TRUE",
				"DB_PORT":              "5433",
				"DB_MAX_OPEN_CONNS":    "20",
				"DB_MAX_IDLE_CONNS":    "10",
				"DB_CONN_MAX_LIFETIME": "0s",
				"REDIS_PORT":           "6380",
				"REDIS_DB":             "15",
				"REDIS_POOL_SIZE":      "0",
				"REDIS_DEFAULT_TTL":    "0s",
				"JWT_SECRET":           "custom-production-secret",
				"JWT_EXPIRY":           "45m",
				"JWT_REFRESH_EXPIRY":   "36h",
				"RATE_LIMIT_MAX":       "50",
				"RATE_LIMIT_WINDOW":    "2m",
			},
			check: func(t *testing.T, cfg *Config) {
				if cfg.App.Env != "production" || cfg.App.Port != 8443 || !cfg.App.Prefork {
					t.Fatalf("unexpected app overrides: %+v", cfg.App)
				}
				if cfg.Database.Port != 5433 || cfg.Database.MaxOpenConns != 20 || cfg.Database.MaxIdleConns != 10 || cfg.Database.ConnMaxLifetime != 0 {
					t.Fatalf("unexpected database overrides: %+v", cfg.Database)
				}
				if cfg.Redis.Port != 6380 || cfg.Redis.DB != 15 || cfg.Redis.PoolSize != 0 || cfg.Redis.DefaultTTL != 0 {
					t.Fatalf("unexpected Redis overrides: %+v", cfg.Redis)
				}
				if cfg.JWT.Secret != "custom-production-secret" || cfg.JWT.Expiry != 45*time.Minute || cfg.JWT.RefreshExpiry != 36*time.Hour {
					t.Fatalf("unexpected JWT overrides: %+v", cfg.JWT)
				}
				if cfg.RateLimit.Max != 50 || cfg.RateLimit.Window != 2*time.Minute {
					t.Fatalf("unexpected rate-limit overrides: %+v", cfg.RateLimit)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv(t)
			for key, value := range tt.env {
				t.Setenv(key, value)
			}
			cfg, err := New()
			if err != nil {
				t.Fatal(err)
			}
			tt.check(t, cfg)
		})
	}
}

func TestNewRejectsMalformedOrIncoherentSettings(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantInError string
	}{
		{"empty app port", map[string]string{"APP_PORT": ""}, "APP_PORT"},
		{"malformed app port", map[string]string{"APP_PORT": "abc"}, "APP_PORT"},
		{"app port below range", map[string]string{"APP_PORT": "0"}, "APP_PORT"},
		{"app port above range", map[string]string{"APP_PORT": "65536"}, "APP_PORT"},
		{"malformed database port", map[string]string{"DB_PORT": "abc"}, "DB_PORT"},
		{"database port below range", map[string]string{"DB_PORT": "0"}, "DB_PORT"},
		{"database port above range", map[string]string{"DB_PORT": "65536"}, "DB_PORT"},
		{"malformed Redis port", map[string]string{"REDIS_PORT": "abc"}, "REDIS_PORT"},
		{"Redis port below range", map[string]string{"REDIS_PORT": "65536"}, "REDIS_PORT"},
		{"negative Redis database", map[string]string{"REDIS_DB": "-1"}, "REDIS_DB"},
		{"negative Redis pool size", map[string]string{"REDIS_POOL_SIZE": "-1"}, "REDIS_POOL_SIZE"},
		{"malformed prefork flag", map[string]string{"APP_PREFORK": "yes"}, "APP_PREFORK"},
		{"empty prefork flag", map[string]string{"APP_PREFORK": ""}, "APP_PREFORK"},
		{"malformed max open connections", map[string]string{"DB_MAX_OPEN_CONNS": "many"}, "DB_MAX_OPEN_CONNS"},
		{"negative max open connections", map[string]string{"DB_MAX_OPEN_CONNS": "-1"}, "DB_MAX_OPEN_CONNS"},
		{"malformed max idle connections", map[string]string{"DB_MAX_IDLE_CONNS": "many"}, "DB_MAX_IDLE_CONNS"},
		{"negative max idle connections", map[string]string{"DB_MAX_IDLE_CONNS": "-1"}, "DB_MAX_IDLE_CONNS"},
		{"idle connections exceed open limit", map[string]string{"DB_MAX_OPEN_CONNS": "4", "DB_MAX_IDLE_CONNS": "5"}, "DB_MAX_IDLE_CONNS"},
		{"malformed connection lifetime", map[string]string{"DB_CONN_MAX_LIFETIME": "soon"}, "DB_CONN_MAX_LIFETIME"},
		{"malformed Redis TTL", map[string]string{"REDIS_DEFAULT_TTL": "soon"}, "REDIS_DEFAULT_TTL"},
		{"negative Redis TTL", map[string]string{"REDIS_DEFAULT_TTL": "-1s"}, "REDIS_DEFAULT_TTL"},
		{"malformed access lifetime", map[string]string{"JWT_EXPIRY": "soon"}, "JWT_EXPIRY"},
		{"zero access lifetime", map[string]string{"JWT_EXPIRY": "0s"}, "JWT_EXPIRY"},
		{"negative access lifetime", map[string]string{"JWT_EXPIRY": "-1s"}, "JWT_EXPIRY"},
		{"malformed refresh lifetime", map[string]string{"JWT_REFRESH_EXPIRY": "soon"}, "JWT_REFRESH_EXPIRY"},
		{"zero refresh lifetime", map[string]string{"JWT_REFRESH_EXPIRY": "0s"}, "JWT_REFRESH_EXPIRY"},
		{"negative refresh lifetime", map[string]string{"JWT_REFRESH_EXPIRY": "-1s"}, "JWT_REFRESH_EXPIRY"},
		{"malformed rate limit", map[string]string{"RATE_LIMIT_MAX": "many"}, "RATE_LIMIT_MAX"},
		{"zero rate limit", map[string]string{"RATE_LIMIT_MAX": "0"}, "RATE_LIMIT_MAX"},
		{"negative rate limit", map[string]string{"RATE_LIMIT_MAX": "-1"}, "RATE_LIMIT_MAX"},
		{"malformed rate window", map[string]string{"RATE_LIMIT_WINDOW": "soon"}, "RATE_LIMIT_WINDOW"},
		{"zero rate window", map[string]string{"RATE_LIMIT_WINDOW": "0s"}, "RATE_LIMIT_WINDOW"},
		{"negative rate window", map[string]string{"RATE_LIMIT_WINDOW": "-1s"}, "RATE_LIMIT_WINDOW"},
		{"empty environment", map[string]string{"APP_ENV": ""}, "APP_ENV"},
		{"empty JWT secret", map[string]string{"JWT_SECRET": ""}, "JWT_SECRET"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv(t)
			for key, value := range tt.env {
				t.Setenv(key, value)
			}
			if _, err := New(); err == nil || !strings.Contains(err.Error(), tt.wantInError) {
				t.Fatalf("New() error = %v, want error mentioning %s", err, tt.wantInError)
			}
		})
	}
}

func TestNewAllowsUnlimitedSQLConnections(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DB_MAX_OPEN_CONNS", "0")
	t.Setenv("DB_MAX_IDLE_CONNS", "25")

	cfg, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.MaxOpenConns != 0 || cfg.Database.MaxIdleConns != 25 {
		t.Fatalf("unexpected SQL pool settings: %+v", cfg.Database)
	}
}

func TestNewRequiresSafeProductionJWTSecret(t *testing.T) {
	for _, secret := range []string{"", "your-secret-key", "YOUR-SUPER-SECRET-JWT-KEY-CHANGE-THIS-IN-PRODUCTION", "change-this-in-production"} {
		t.Run("rejects "+secret, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv("APP_ENV", "production")
			if secret != "" {
				t.Setenv("JWT_SECRET", secret)
			}
			_, err := New()
			if err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
				t.Fatalf("New() error = %v, want production JWT_SECRET error", err)
			}
			if secret != "" && strings.Contains(err.Error(), secret) {
				t.Fatal("error exposed JWT_SECRET value")
			}
		})
	}

	clearConfigEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "custom-production-secret")
	if _, err := New(); err != nil {
		t.Fatalf("New() rejected configured production secret: %v", err)
	}
}
