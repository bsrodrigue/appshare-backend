// Package middleware provides HTTP middleware components.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// requestIDKey is the context key for the request ID.
type requestIDKey struct{}

// RequestID retrieves the request ID from context.
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	wroteHeader  bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default to 200 if WriteHeader isn't called
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Flush implements http.Flusher if the underlying ResponseWriter supports it.
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// LoggingConfig holds configuration for the logging middleware.
type LoggingConfig struct {
	// Logger is the slog logger to use. If nil, uses slog.Default().
	Logger *slog.Logger

	// SkipPaths are paths that should not be logged (e.g., health checks).
	SkipPaths []string

	// SensitiveHeaders are header names that should be redacted in logs.
	SensitiveHeaders []string

	// LogRequestBody enables logging of request body (use with caution!).
	LogRequestBody bool

	// LogResponseBody enables logging of response body (use with caution!).
	LogResponseBody bool
}

// DefaultLoggingConfig returns sensible defaults.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Logger:    nil, // Will use slog.Default()
		SkipPaths: []string{"/health", "/healthz", "/ready", "/readyz", "/metrics"},
		SensitiveHeaders: []string{
			"Authorization",
			"Cookie",
			"Set-Cookie",
			"X-Api-Key",
			"X-Auth-Token",
		},
		LogRequestBody:  false,
		LogResponseBody: false,
	}
}

// LoggingMiddleware provides structured request/response logging.
type LoggingMiddleware struct {
	config LoggingConfig
	logger *slog.Logger
}

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware(config LoggingConfig) *LoggingMiddleware {
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &LoggingMiddleware{
		config: config,
		logger: logger,
	}
}

// Handler returns the logging middleware handler.
func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this path should be skipped
		if m.shouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Generate request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()[:8] // Short ID for readability
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		r = r.WithContext(ctx)

		// Wrap response writer to capture status and bytes
		rw := newResponseWriter(w)

		// Log request start
		m.logRequest(r, requestID)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log response
		m.logResponse(r, rw, requestID, duration)
	})
}

// logRequest logs the incoming request.
func (m *LoggingMiddleware) logRequest(r *http.Request, requestID string) {
	attrs := []slog.Attr{
		slog.String("request_id", requestID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", getClientIP(r)),
		slog.String("user_agent", r.UserAgent()),
	}

	// Add query params if present
	if r.URL.RawQuery != "" {
		attrs = append(attrs, slog.String("query", r.URL.RawQuery))
	}

	// Add content length if present
	if r.ContentLength > 0 {
		attrs = append(attrs, slog.Int64("content_length", r.ContentLength))
	}

	// Add sanitized headers
	headers := m.sanitizeHeaders(r.Header)
	if len(headers) > 0 {
		attrs = append(attrs, slog.Any("headers", headers))
	}

	m.logger.LogAttrs(r.Context(), slog.LevelInfo, "→ request",
		attrs...,
	)
}

// logResponse logs the response.
func (m *LoggingMiddleware) logResponse(r *http.Request, rw *responseWriter, requestID string, duration time.Duration) {
	level := m.getLogLevel(rw.statusCode)

	attrs := []slog.Attr{
		slog.String("request_id", requestID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", rw.statusCode),
		slog.String("status_text", http.StatusText(rw.statusCode)),
		slog.Int("bytes", rw.bytesWritten),
		slog.Duration("duration", duration),
		slog.String("duration_human", formatDuration(duration)),
	}

	// Add latency warning for slow requests
	if duration > 1*time.Second {
		attrs = append(attrs, slog.Bool("slow_request", true))
	}

	m.logger.LogAttrs(r.Context(), level, "← response",
		attrs...,
	)
}

// getLogLevel returns the appropriate log level based on status code.
func (m *LoggingMiddleware) getLogLevel(statusCode int) slog.Level {
	switch {
	case statusCode >= 500:
		return slog.LevelError // Server errors
	case statusCode >= 400:
		return slog.LevelWarn // Client errors
	default:
		return slog.LevelInfo // Success
	}
}

// shouldSkip returns true if the path should not be logged.
func (m *LoggingMiddleware) shouldSkip(path string) bool {
	for _, skip := range m.config.SkipPaths {
		if path == skip || strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// sanitizeHeaders removes sensitive headers from logs.
func (m *LoggingMiddleware) sanitizeHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)

	for key := range headers {
		if m.isSensitiveHeader(key) {
			result[key] = "[REDACTED]"
		} else {
			// Only log first value for each header
			result[key] = headers.Get(key)
		}
	}

	// Remove headers we don't need to log
	delete(result, "Accept")
	delete(result, "Accept-Encoding")
	delete(result, "Accept-Language")
	delete(result, "Connection")
	delete(result, "User-Agent") // Already logged separately

	return result
}

// isSensitiveHeader checks if a header should be redacted.
func (m *LoggingMiddleware) isSensitiveHeader(name string) bool {
	nameLower := strings.ToLower(name)
	for _, sensitive := range m.config.SensitiveHeaders {
		if strings.ToLower(sensitive) == nameLower {
			return true
		}
	}
	return false
}

// getClientIP extracts the real client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (set by some proxies)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	// Remove port if present
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return d.Round(time.Microsecond).String()
	case d < time.Second:
		return d.Round(time.Millisecond).String()
	default:
		return d.Round(10 * time.Millisecond).String()
	}
}
