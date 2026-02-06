// Package domain contains the core business entities and errors.
// This package has NO external dependencies - it's the heart of the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the core user entity in the domain.
// This is NOT the database model - it's the business representation.
type User struct {
	ID          uuid.UUID
	Email       string
	Username    string
	PhoneNumber string
	FirstName   string
	LastName    string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time // nil if never logged in
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// UserCredentials contains sensitive auth data - used only for login.
type UserCredentials struct {
	ID           uuid.UUID
	Email        string
	Username     string
	PasswordHash string
	IsActive     bool
}

// CreateUserInput represents the data needed to create a new user.
type CreateUserInput struct {
	Email       string
	Username    string
	PhoneNumber string
	Password    string // Plain text - will be hashed by service
	FirstName   string
	LastName    string
}
