package repository

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// ArtifactRepository defines the interface for artifact data access.
type ArtifactRepository interface {
	// Create creates a new artifact record.
	Create(ctx context.Context, input domain.CreateArtifactInput) (*domain.Artifact, error)

	// GetByID retrieves an artifact by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Artifact, error)

	// ListByRelease retrieves all artifacts for a release.
	ListByRelease(ctx context.Context, releaseID uuid.UUID) ([]*domain.Artifact, error)

	// Delete removes an artifact record.
	Delete(ctx context.Context, id uuid.UUID) error

	// ========== Transaction Methods ==========

	// CreateTx creates a new artifact record within a transaction.
	CreateTx(ctx context.Context, q *db.Queries, input domain.CreateArtifactInput) (*domain.Artifact, error)
}
