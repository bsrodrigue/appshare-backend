package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration.
// All fields are validated at startup - fail fast if misconfigured.
type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// JWT (for future auth)
	JWTSecret     string
	JWTExpiration int // hours
}

// Load reads configuration from environment variables.
// Returns an error if required variables are missing.
func Load() (*Config, error) {
	cfg := &Config{}

	// Server config
	cfg.Port = getEnv("PORT", "8080")

	// Database URL - can be explicit or built from components
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		dbURL, err := buildDatabaseURL()
		if err != nil {
			return nil, fmt.Errorf("database configuration error: %w", err)
		}
		cfg.DatabaseURL = dbURL
	}

	// JWT config (optional for now, will be used later)
	cfg.JWTSecret = getEnv("JWT_SECRET", "change-me-in-production")
	cfg.JWTExpiration = getEnvAsInt("JWT_EXPIRATION_HOURS", 24)

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

// getEnvAsInt returns an environment variable as an integer, or a default.
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
