package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port      string
	Env       string
	DB        DBConfig
	JWT       JWTConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Timeout   time.Duration
	Google    GoogleConfig
	Storage   StorageConfig
}

// DBConfig holds PostgreSQL connection settings.
type DBConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
}

// JWTConfig holds JWT signing settings.
type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

// CORSConfig holds CORS allowed origins.
type CORSConfig struct {
	Origins []string
}

// RateLimitConfig holds rate limiting parameters.
type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
}

// GoogleConfig holds Google OAuth credentials.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// StorageConfig holds file storage settings.
type StorageConfig struct {
	Driver  string // "local" or "s3"
	BaseDir string // local filesystem base directory
	BaseURL string // public URL prefix for local storage
}

// Load reads configuration from environment variables, optionally loading a .env file first.
func Load() (*Config, error) {
	// Load .env file if present; ignore error if file doesn't exist.
	_ = godotenv.Load()

	dbMaxConns, err := parseInt32Env("DB_MAX_CONNS", 25)
	if err != nil {
		return nil, fmt.Errorf("DB_MAX_CONNS: %w", err)
	}
	dbMinConns, err := parseInt32Env("DB_MIN_CONNS", 5)
	if err != nil {
		return nil, fmt.Errorf("DB_MIN_CONNS: %w", err)
	}
	dbMaxConnLifetime, err := parseDurationEnv("DB_MAX_CONN_LIFETIME", 30*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("DB_MAX_CONN_LIFETIME: %w", err)
	}

	jwtExpiry, err := parseDurationEnv("JWT_EXPIRY", 168*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRY: %w", err)
	}

	timeout, err := parseDurationEnv("REQUEST_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("REQUEST_TIMEOUT: %w", err)
	}

	rateRPS, err := parseFloat64Env("RATE_LIMIT_RPS", 100)
	if err != nil {
		return nil, fmt.Errorf("RATE_LIMIT_RPS: %w", err)
	}
	rateBurst, err := parseIntEnv("RATE_LIMIT_BURST", 20)
	if err != nil {
		return nil, fmt.Errorf("RATE_LIMIT_BURST: %w", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	corsOrigins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ",")

	return &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("ENV", "development"),
		DB: DBConfig{
			URL:             dbURL,
			MaxConns:        dbMaxConns,
			MinConns:        dbMinConns,
			MaxConnLifetime: dbMaxConnLifetime,
		},
		JWT: JWTConfig{
			Secret: jwtSecret,
			Expiry: jwtExpiry,
		},
		CORS: CORSConfig{
			Origins: corsOrigins,
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: rateRPS,
			Burst:             rateBurst,
		},
		Timeout: timeout,
		Google: GoogleConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		},
		Storage: StorageConfig{
			Driver:  getEnv("STORAGE_DRIVER", "local"),
			BaseDir: getEnv("STORAGE_BASE_DIR", "assets"),
			BaseURL: getEnv("STORAGE_BASE_URL", "http://localhost:8080/assets"),
		},
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func parseInt32Env(key string, defaultVal int32) (int32, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal, nil
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q: %w", v, err)
	}
	return int32(n), nil
}

func parseIntEnv(key string, defaultVal int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q: %w", v, err)
	}
	return n, nil
}

func parseFloat64Env(key string, defaultVal float64) (float64, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q: %w", v, err)
	}
	return f, nil
}

func parseDurationEnv(key string, defaultVal time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", v, err)
	}
	return d, nil
}
