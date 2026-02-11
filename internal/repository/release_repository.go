package repository

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// ReleaseRepository defines the interface for release data access.
type ReleaseRepository interface {
	// Create creates a new release.
	Create(ctx context.Context, input domain.CreateReleaseInput) (*domain.ApplicationRelease, error)

	// GetByID retrieves a release by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ApplicationRelease, error)

	// ListByApplication retrieves all releases for an application.
	ListByApplication(ctx context.Context, appID uuid.UUID) ([]*domain.ApplicationRelease, error)

	// ListByEnvironment retrieves releases for an application filtered by environment.
	ListByEnvironment(ctx context.Context, appID uuid.UUID, env domain.ReleaseEnvironment) ([]*domain.ApplicationRelease, error)

	// GetLatestByEnvironment retrieves the latest release for an application in an environment.
	GetLatestByEnvironment(ctx context.Context, appID uuid.UUID, env domain.ReleaseEnvironment) (*domain.ApplicationRelease, error)

	// Update updates a release's title and release note.
	Update(ctx context.Context, id uuid.UUID, title, releaseNote string) (*domain.ApplicationRelease, error)

	// Promote updates the environment of a release.
	Promote(ctx context.Context, id uuid.UUID, env domain.ReleaseEnvironment) (*domain.ApplicationRelease, error)

	// SoftDelete marks a release as deleted.
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// VersionExists checks if a version code already exists for an application in an environment.
	VersionExists(ctx context.Context, appID uuid.UUID, versionCode int32, env domain.ReleaseEnvironment) (bool, error)
}
