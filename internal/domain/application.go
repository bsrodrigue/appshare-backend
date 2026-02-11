package domain

import (
	"time"

	"github.com/google/uuid"
)

// Application represents an app within a project (e.g., "Main App Android").
type Application struct {
	ID          uuid.UUID
	Title       string
	PackageName string
	Description string
	ProjectID   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateApplicationInput represents data needed to create a new application.
type CreateApplicationInput struct {
	Title       string
	PackageName string
	Description string
	ProjectID   uuid.UUID
}

// UpdateApplicationInput represents data needed to update an existing application.
type UpdateApplicationInput struct {
	Title       string
	Description string
}
