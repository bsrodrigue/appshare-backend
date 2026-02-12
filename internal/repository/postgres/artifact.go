package postgres

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// ArtifactRepository implements repository.ArtifactRepository using PostgreSQL.
type ArtifactRepository struct {
	q *db.Queries
}

// NewArtifactRepository creates a new PostgreSQL artifact repository.
func NewArtifactRepository(q *db.Queries) *ArtifactRepository {
	return &ArtifactRepository{q: q}
}

// Create creates a new artifact.
func (r *ArtifactRepository) Create(ctx context.Context, input domain.CreateArtifactInput) (*domain.Artifact, error) {
	row, err := r.q.CreateArtifact(ctx, db.CreateArtifactParams{
		FileUrl:    input.FileURL,
		Sha256Hash: input.SHA256,
		FileSize:   input.FileSize,
		FileType:   input.FileType,
		Abi:        stringToPgtype(derefString(input.ABI)),
		ReleaseID:  uuidToPgtype(input.ReleaseID),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToArtifact(&row), nil
}

// GetByID retrieves an artifact by ID.
func (r *ArtifactRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Artifact, error) {
	row, err := r.q.GetArtifactByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, translateError(err)
	}
	return rowToArtifact(&row), nil
}

// ListByRelease retrieves all artifacts for a release.
func (r *ArtifactRepository) ListByRelease(ctx context.Context, releaseID uuid.UUID) ([]*domain.Artifact, error) {
	rows, err := r.q.ListArtifactsByRelease(ctx, uuidToPgtype(releaseID))
	if err != nil {
		return nil, translateError(err)
	}

	artifacts := make([]*domain.Artifact, len(rows))
	for i, row := range rows {
		artifacts[i] = rowToArtifact(&row)
	}
	return artifacts, nil
}

// Delete marks an artifact as deleted.
func (r *ArtifactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.q.SoftDeleteArtifact(ctx, uuidToPgtype(id))
	return translateError(err)
}

// ========== Transaction Methods ==========

// CreateTx creates a new artifact record within a transaction.
func (r *ArtifactRepository) CreateTx(ctx context.Context, q *db.Queries, input domain.CreateArtifactInput) (*domain.Artifact, error) {
	row, err := q.CreateArtifact(ctx, db.CreateArtifactParams{
		FileUrl:    input.FileURL,
		Sha256Hash: input.SHA256,
		FileSize:   input.FileSize,
		FileType:   input.FileType,
		Abi:        stringToPgtype(derefString(input.ABI)),
		ReleaseID:  uuidToPgtype(input.ReleaseID),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToArtifact(&row), nil
}

// Helper to convert DB row to domain Artifact
func rowToArtifact(row *db.Artifact) *domain.Artifact {
	return &domain.Artifact{
		ID:        pgtypeToUUID(row.ID),
		FileURL:   row.FileUrl,
		SHA256:    row.Sha256Hash,
		FileSize:  row.FileSize,
		FileType:  row.FileType,
		ABI:       pgtypeToStringPtr(row.Abi),
		ReleaseID: pgtypeToUUID(row.ReleaseID),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: pgtypeToTimePtr(row.DeletedAt),
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
