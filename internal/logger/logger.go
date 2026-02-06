// Package logger provides structured logging utilities.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Config holds logger configuration.
type Config struct {
	// Level is the minimum log level (debug, info, warn, error).
	Level string

	// Format is the log format (json, text).
	Format string

	// Output is where logs are written (stdout, stderr, or a file path).
	Output string

	// AddSource adds source file and line number to logs.
	AddSource bool
}

// DefaultConfig returns sensible defaults for development.
func DefaultConfig() Config {
	return Config{
		Level:     "info",
		Format:    "text", // Use "json" for production
		Output:    "stdout",
		AddSource: false,
	}
}

// ProductionConfig returns sensible defaults for production.
func ProductionConfig() Config {
	return Config{
		Level:     "info",
		Format:    "json",
		Output:    "stdout",
		AddSource: true,
	}
}

// New creates a new slog.Logger based on the configuration.
func New(cfg Config) (*slog.Logger, error) {
	// Parse log level
	level := parseLevel(cfg.Level)

	// Get output writer
	output, err := getOutput(cfg.Output)
	if err != nil {
		return nil, err
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	// Create handler based on format
	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler), nil
}

// SetDefault creates a logger and sets it as the default slog logger.
func SetDefault(cfg Config) error {
	logger, err := New(cfg)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)
	return nil
}

// parseLevel parses a string log level.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// getOutput returns the appropriate io.Writer for the given output string.
func getOutput(output string) (io.Writer, error) {
	switch strings.ToLower(output) {
	case "stdout", "":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		// Assume it's a file path
		return os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	}
}

// With returns a logger with the given attributes.
func With(logger *slog.Logger, attrs ...slog.Attr) *slog.Logger {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return logger.With(args...)
}

// WithRequestID returns a logger with a request ID attribute.
func WithRequestID(logger *slog.Logger, requestID string) *slog.Logger {
	return logger.With(slog.String("request_id", requestID))
}

// WithUserID returns a logger with a user ID attribute.
func WithUserID(logger *slog.Logger, userID string) *slog.Logger {
	return logger.With(slog.String("user_id", userID))
}

// WithError returns a logger with an error attribute.
func WithError(logger *slog.Logger, err error) *slog.Logger {
	return logger.With(slog.String("error", err.Error()))
}
