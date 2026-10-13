package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
}

type AppConfig struct {
	Name    string
	Env     string
	Port    int
	Prefork bool
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	PoolSize   int
	DefaultTTL time.Duration
}

type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type RateLimitConfig struct {
	Max    int
	Window time.Duration
}

func New() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "Go Fiber Auth API"),
			Env:  getEnv("APP_ENV", "development"),
			Port: 3000,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            5432,
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Name:            getEnv("DB_NAME", "go_fiber_auth"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			Host:       getEnv("REDIS_HOST", "localhost"),
			Port:       6379,
			Password:   getEnv("REDIS_PASSWORD", ""),
			PoolSize:   10,
			DefaultTTL: time.Hour,
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-secret-key"),
			Expiry:        24 * time.Hour,
			RefreshExpiry: 168 * time.Hour,
		},
		CORS: CORSConfig{
			AllowedOrigins: splitAndTrim(getEnv("CORS_ALLOWED_ORIGINS", "*")),
			AllowedMethods: splitAndTrim(getEnv("CORS_ALLOWED_METHODS", "GET,POST,HEAD,PUT,DELETE,PATCH")),
			AllowedHeaders: splitAndTrim(getEnv("CORS_ALLOWED_HEADERS", "*")),
		},
		RateLimit: RateLimitConfig{
			Max:    100,
			Window: time.Minute,
		},
	}

	var err error
	if cfg.App.Port, err = intEnv("APP_PORT", cfg.App.Port); err != nil {
		return nil, err
	}
	if cfg.App.Prefork, err = boolEnv("APP_PREFORK", false); err != nil {
		return nil, err
	}
	if cfg.Database.Port, err = intEnv("DB_PORT", cfg.Database.Port); err != nil {
		return nil, err
	}
	if cfg.Database.MaxOpenConns, err = intEnv("DB_MAX_OPEN_CONNS", cfg.Database.MaxOpenConns); err != nil {
		return nil, err
	}
	if cfg.Database.MaxIdleConns, err = intEnv("DB_MAX_IDLE_CONNS", cfg.Database.MaxIdleConns); err != nil {
		return nil, err
	}
	if cfg.Database.ConnMaxLifetime, err = durationEnv("DB_CONN_MAX_LIFETIME", cfg.Database.ConnMaxLifetime); err != nil {
		return nil, err
	}
	if cfg.Redis.Port, err = intEnv("REDIS_PORT", cfg.Redis.Port); err != nil {
		return nil, err
	}
	if cfg.Redis.DB, err = intEnv("REDIS_DB", cfg.Redis.DB); err != nil {
		return nil, err
	}
	if cfg.Redis.PoolSize, err = intEnv("REDIS_POOL_SIZE", cfg.Redis.PoolSize); err != nil {
		return nil, err
	}
	if cfg.Redis.DefaultTTL, err = durationEnv("REDIS_DEFAULT_TTL", cfg.Redis.DefaultTTL); err != nil {
		return nil, err
	}
	if cfg.JWT.Expiry, err = durationEnv("JWT_EXPIRY", cfg.JWT.Expiry); err != nil {
		return nil, err
	}
	if cfg.JWT.RefreshExpiry, err = durationEnv("JWT_REFRESH_EXPIRY", cfg.JWT.RefreshExpiry); err != nil {
		return nil, err
	}
	if cfg.RateLimit.Max, err = intEnv("RATE_LIMIT_MAX", cfg.RateLimit.Max); err != nil {
		return nil, err
	}
	if cfg.RateLimit.Window, err = durationEnv("RATE_LIMIT_WINDOW", cfg.RateLimit.Window); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func intEnv(key string, defaultValue int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue, nil
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return i, nil
}

func durationEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	return duration, nil
}

func boolEnv(key string, defaultValue bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return parsed, nil
}

func (c *Config) validate() error {
	c.App.Env = strings.ToLower(strings.TrimSpace(c.App.Env))
	if c.App.Env == "" {
		return fmt.Errorf("APP_ENV must not be empty")
	}
	if c.App.Env == "prod" {
		c.App.Env = "production"
	}

	for _, port := range []struct {
		name  string
		value int
	}{{"APP_PORT", c.App.Port}, {"DB_PORT", c.Database.Port}, {"REDIS_PORT", c.Redis.Port}} {
		if port.value < 1 || port.value > 65535 {
			return fmt.Errorf("%s must be between 1 and 65535", port.name)
		}
	}
	if c.Database.MaxOpenConns < 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be non-negative")
	}
	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("DB_MAX_IDLE_CONNS must be non-negative")
	}
	if c.Database.MaxOpenConns > 0 && c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("DB_MAX_IDLE_CONNS cannot exceed DB_MAX_OPEN_CONNS")
	}
	if c.Redis.DB < 0 {
		return fmt.Errorf("REDIS_DB must be non-negative")
	}
	if c.Redis.PoolSize < 0 {
		return fmt.Errorf("REDIS_POOL_SIZE must be non-negative")
	}
	if c.Redis.DefaultTTL < 0 {
		return fmt.Errorf("REDIS_DEFAULT_TTL must be non-negative")
	}
	if c.JWT.Expiry <= 0 {
		return fmt.Errorf("JWT_EXPIRY must be positive")
	}
	if c.JWT.RefreshExpiry <= 0 {
		return fmt.Errorf("JWT_REFRESH_EXPIRY must be positive")
	}
	if strings.TrimSpace(c.JWT.Secret) == "" {
		return fmt.Errorf("JWT_SECRET must not be empty")
	}
	if c.App.Env == "production" && isPlaceholderSecret(c.JWT.Secret) {
		return fmt.Errorf("JWT_SECRET must be set to a non-placeholder value when APP_ENV=production")
	}
	if c.RateLimit.Max <= 0 {
		return fmt.Errorf("RATE_LIMIT_MAX must be positive")
	}
	if c.RateLimit.Window <= 0 {
		return fmt.Errorf("RATE_LIMIT_WINDOW must be positive")
	}
	return nil
}

func isPlaceholderSecret(secret string) bool {
	switch strings.ToLower(strings.TrimSpace(secret)) {
	case "your-secret-key", "your-secret", "your-jwt-secret", "your-super-secret-jwt-key-change-this-in-production",
		"change-this-in-production", "change-me", "changeme", "replace-me", "secret", "password":
		return true
	default:
		return false
	}
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
