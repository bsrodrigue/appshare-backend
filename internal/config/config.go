package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port        string
	Environment string // development, staging, production

	// Logging
	LogLevel  string // debug, info, warn, error
	LogFormat string // text, json

	// Database
	DatabaseURL string

	// JWT
	JWTSecretKey            string
	JWTAccessTokenDuration  time.Duration
	JWTRefreshTokenDuration time.Duration
	JWTIssuer               string
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Server config
	cfg.Port = getEnv("PORT", "8080")
	cfg.Environment = getEnv("ENVIRONMENT", "development")

	// Logging config - defaults based on environment
	if cfg.Environment == "production" {
		cfg.LogLevel = getEnv("LOG_LEVEL", "info")
		cfg.LogFormat = getEnv("LOG_FORMAT", "json")
	} else {
		cfg.LogLevel = getEnv("LOG_LEVEL", "debug")
		cfg.LogFormat = getEnv("LOG_FORMAT", "text")
	}

	// Database URL - can be explicit or built from components
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		dbURL, err := buildDatabaseURL()
		if err != nil {
			return nil, fmt.Errorf("database configuration error: %w", err)
		}
		cfg.DatabaseURL = dbURL
	}

	// JWT config
	cfg.JWTSecretKey = os.Getenv("JWT_SECRET_KEY")
	if cfg.JWTSecretKey == "" {
		// In development, use a default (NEVER in production!)
		if cfg.Environment == "production" {
			return nil, fmt.Errorf("JWT_SECRET_KEY is required in production")
		}
		cfg.JWTSecretKey = "CHANGE-THIS-IN-PRODUCTION-use-openssl-rand-base64-32"
	}
	if len(cfg.JWTSecretKey) < 32 {
		return nil, fmt.Errorf("JWT_SECRET_KEY must be at least 32 characters")
	}

	// Token durations (with sensible defaults)
	cfg.JWTAccessTokenDuration = getEnvAsDuration("JWT_ACCESS_TOKEN_MINUTES", 15*time.Minute)
	cfg.JWTRefreshTokenDuration = getEnvAsDuration("JWT_REFRESH_TOKEN_DAYS", 7*24*time.Hour)
	cfg.JWTIssuer = getEnv("JWT_ISSUER", "appshare")

	return cfg, nil
}

// buildDatabaseURL constructs a PostgreSQL connection URL from individual env vars.
func buildDatabaseURL() (string, error) {
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return "", fmt.Errorf("POSTGRES_USER is required")
	}

	pass := os.Getenv("POSTGRES_PASSWORD")
	if pass == "" {
		return "", fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		return "", fmt.Errorf("POSTGRES_DB is required")
	}

	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	sslMode := getEnv("POSTGRES_SSLMODE", "disable")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, pass, host, port, dbName, sslMode,
	), nil
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			// Determine the unit based on the key name
			if containsIgnoreCase(key, "MINUTES") {
				return time.Duration(intValue) * time.Minute
			}
			if containsIgnoreCase(key, "DAYS") {
				return time.Duration(intValue) * 24 * time.Hour
			}
			if containsIgnoreCase(key, "HOURS") {
				return time.Duration(intValue) * time.Hour
			}
			// Default to minutes
			return time.Duration(intValue) * time.Minute
		}
	}
	return defaultValue
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
