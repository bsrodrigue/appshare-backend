package auth

import (
	"context"
	"strings"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	userContextKey contextKey = "user"
)

// AuthenticatedUser represents the user data stored in context after authentication.
type AuthenticatedUser struct {
	ID    uuid.UUID
	Email string
}

// UserFromContext extracts the authenticated user from context.
// Returns nil if no user is authenticated.
func UserFromContext(ctx context.Context) *AuthenticatedUser {
	user, ok := ctx.Value(userContextKey).(*AuthenticatedUser)
	if !ok {
		return nil
	}
	return user
}

// MustUserFromContext extracts the authenticated user from context.
// Panics if no user is found - use only in handlers that are definitely authenticated.
func MustUserFromContext(ctx context.Context) *AuthenticatedUser {
	user := UserFromContext(ctx)
	if user == nil {
		panic("no authenticated user in context - middleware missing?")
	}
	return user
}

// ContextWithUser adds an authenticated user to the context.
func ContextWithUser(ctx context.Context, user *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// ExtractBearerToken extracts the token from an Authorization header.
// Expected format: "Bearer <token>"
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", domain.NewAppError(domain.CodeUnauthorized, "authorization header required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", domain.NewAppError(domain.CodeUnauthorized, "invalid authorization header format")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", domain.NewAppError(domain.CodeUnauthorized, "token required")
	}

	return token, nil
}
