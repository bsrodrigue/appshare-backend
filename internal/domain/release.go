package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReleaseEnvironment represents the environment of a release.
type ReleaseEnvironment string

const (
	EnvironmentDevelopment ReleaseEnvironment = "development"
	EnvironmentStaging     ReleaseEnvironment = "staging"
	EnvironmentProduction  ReleaseEnvironment = "production"
)

// ApplicationRelease represents a specific version of an application.
type ApplicationRelease struct {
	ID            uuid.UUID
	Title         string
	VersionCode   int32
	VersionName   string
	ReleaseNote   string
	Environment   ReleaseEnvironment
	ApplicationID uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CreateReleaseInput represents data needed to create a new release.
type CreateReleaseInput struct {
	Title         string
	VersionCode   int32
	VersionName   string
	ReleaseNote   string
	Environment   ReleaseEnvironment
	ApplicationID uuid.UUID
}

// UpdateReleaseInput represents data needed to update an existing release.
type UpdateReleaseInput struct {
	Title       *string
	ReleaseNote *string
}
