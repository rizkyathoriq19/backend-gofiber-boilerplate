package config

import (
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
	Security  SecurityConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Swagger   SwaggerConfig
}

type AppConfig struct {
	Name    string
	Env     string
	Host    string
	Port    string
	Prefork bool
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
	DefaultTTL time.Duration
}

type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

type SecurityConfig struct {
	BCryptCost int
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

type SwaggerConfig struct {
	Hosts    []string
	BasePath string
	Schemes  []string
}

func New() *Config {
	return &Config{
		App: AppConfig{
			Name:    getEnv("APP_NAME", "Go Fiber Auth API"),
			Env:     getEnv("APP_ENV", "development"),
			Host:    getEnv("APP_HOST", "localhost"),
			Port:    getEnv("APP_PORT", "3000"),
			Prefork: getEnv("APP_PREFORK", "false") == "true",
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Name:            getEnv("DB_NAME", "go_fiber_auth"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    parseInt(getEnv("DB_MAX_OPEN_CONNS", "25"), 25),
			MaxIdleConns:    parseInt(getEnv("DB_MAX_IDLE_CONNS", "25"), 25),
			ConnMaxLifetime: parseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"), 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       parseInt(getEnv("REDIS_DB", "0"), 0),
			PoolSize: parseInt(getEnv("REDIS_POOL_SIZE", "10"), 10),
			DefaultTTL: parseDuration(getEnv("REDIS_DEFAULT_TTL", "1h"), 1*time.Hour),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-secret-key"),
			Expiry:        parseDuration(getEnv("JWT_EXPIRY", "24h"), 24*time.Hour),
			RefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"), 168*time.Hour),
		},
		Security: SecurityConfig{
			BCryptCost: 12,
		},
		CORS: CORSConfig{
			AllowedOrigins: splitAndTrim(getEnv("CORS_ALLOWED_ORIGINS", "*")),
			AllowedMethods: splitAndTrim(getEnv("CORS_ALLOWED_METHODS", "GET,POST,HEAD,PUT,DELETE,PATCH")),
			AllowedHeaders: splitAndTrim(getEnv("CORS_ALLOWED_HEADERS", "*")),
		},
		RateLimit: RateLimitConfig{
			Max:    parseInt(getEnv("RATE_LIMIT_MAX", "100"), 100),
			Window: parseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"), time.Minute),
		},
		Swagger: SwaggerConfig{
			Hosts:    splitAndTrim(getEnv("SWAGGER_HOSTS", "localhost:3002")),
			BasePath: getEnv("SWAGGER_BASE_PATH", "/api/v1"),
			Schemes:  splitAndTrim(getEnv("SWAGGER_SCHEMES", "http,https")),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(value string, defaultValue time.Duration) time.Duration {
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	return defaultValue
}

func parseInt(value string, defaultValue int) int {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultValue
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
