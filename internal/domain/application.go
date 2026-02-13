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

type ApplicationMetadata struct {
	PackageName      string
	VersionCode      int64
	VersionName      string
	MinSdkVersion    int
	TargetSdkVersion int
	Architecture     string
	Platform         string
	SHA256           string
	FileSize         int64
}

// CreateApplicationInput represents data needed to create a new application.
type CreateApplicationInput struct {
	Title       string
	PackageName string
	Description string
	ProjectID   uuid.UUID
}

// CreateApplicationFromArtifactInput represents data needed to create a new application from an artifact.
type CreateApplicationFromArtifactInput struct {
	Title       string
	ProjectID   uuid.UUID
	ArtifactURL string
}

// UpdateApplicationInput represents data needed to update an existing application.
type UpdateApplicationInput struct {
	Title       string
	Description string
}
