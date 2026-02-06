// Package middleware provides HTTP middleware components.
package middleware

import (
	"net/http"

	"github.com/bsrodrigue/appshare-backend/internal/auth"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
)

// AuthMiddleware handles JWT authentication for protected routes.
type AuthMiddleware struct {
	jwtService *auth.JWTService
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

// RequireAuth returns a middleware that requires a valid JWT token.
// If the token is invalid or missing, it returns a 401 Unauthorized response.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token, err := auth.ExtractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeUnauthorized(w, err)
			return
		}

		// Validate the access token
		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			writeUnauthorized(w, err)
			return
		}

		// Add user to context
		authUser := &auth.AuthenticatedUser{
			ID:    claims.UserID,
			Email: claims.Email,
		}
		ctx := auth.ContextWithUser(r.Context(), authUser)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth returns a middleware that extracts user info if a token is present,
// but doesn't require authentication. Useful for endpoints that behave differently
// for authenticated vs anonymous users.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header - continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Try to extract and validate token
		token, err := auth.ExtractBearerToken(authHeader)
		if err != nil {
			// Invalid header format - continue without user
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			// Invalid token - continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Valid token - add user to context
		authUser := &auth.AuthenticatedUser{
			ID:    claims.UserID,
			Email: claims.Email,
		}
		ctx := auth.ContextWithUser(r.Context(), authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// writeUnauthorized writes a 401 response with proper JSON format.
func writeUnauthorized(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	code := domain.GetErrorCode(err)
	message := domain.GetErrorMessage(err)

	// Write JSON error response
	response := `{"status":401,"code":"` + string(code) + `","message":"` + message + `"}`
	w.Write([]byte(response))
}
