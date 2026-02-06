// Package repository defines interfaces for data access.
// Implementations are in subpackages (e.g., repository/postgres).
package repository

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access.
// Services depend on this interface, not on concrete implementations.
type UserRepository interface {
	// Create creates a new user and returns it.
	Create(ctx context.Context, input domain.CreateUserInput, passwordHash string) (*domain.User, error)

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// GetByEmail retrieves a user by their email.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByUsername retrieves a user by their username.
	GetByUsername(ctx context.Context, username string) (*domain.User, error)

	// GetCredentialsByEmail retrieves user credentials for authentication.
	GetCredentialsByEmail(ctx context.Context, email string) (*domain.UserCredentials, error)

	// GetCredentialsByUsername retrieves user credentials for authentication.
	GetCredentialsByUsername(ctx context.Context, username string) (*domain.UserCredentials, error)

	// List retrieves all active users.
	List(ctx context.Context) ([]*domain.User, error)

	// UpdateEmail updates a user's email.
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) (*domain.User, error)

	// UpdateUsername updates a user's username.
	UpdateUsername(ctx context.Context, id uuid.UUID, username string) (*domain.User, error)

	// UpdatePassword updates a user's password hash.
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error

	// UpdateProfile updates a user's profile (first name, last name).
	UpdateProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domain.User, error)

	// UpdateLastLogin updates the user's last login timestamp.
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// SoftDelete marks a user as deleted.
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// EmailExists checks if an email is already registered.
	EmailExists(ctx context.Context, email string) (bool, error)

	// UsernameExists checks if a username is already taken.
	UsernameExists(ctx context.Context, username string) (bool, error)
}
