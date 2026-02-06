package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a project that can contain applications.
type Project struct {
	ID          uuid.UUID
	Title       string
	Description string
	OwnerID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateProjectInput represents the data needed to create a new project.
type CreateProjectInput struct {
	Title       string
	Description string
	OwnerID     uuid.UUID
}

// UpdateProjectInput represents updateable project fields.
type UpdateProjectInput struct {
	Title       *string // nil means don't update
	Description *string
}
