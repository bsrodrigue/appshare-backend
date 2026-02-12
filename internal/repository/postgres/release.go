package postgres

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// ReleaseRepository implements repository.ReleaseRepository using PostgreSQL.
type ReleaseRepository struct {
	q *db.Queries
}

// NewReleaseRepository creates a new PostgreSQL release repository.
func NewReleaseRepository(q *db.Queries) *ReleaseRepository {
	return &ReleaseRepository{q: q}
}

// Create creates a new release.
func (r *ReleaseRepository) Create(ctx context.Context, input domain.CreateReleaseInput) (*domain.ApplicationRelease, error) {
	row, err := r.q.CreateApplicationRelease(ctx, db.CreateApplicationReleaseParams{
		Title:         input.Title,
		VersionCode:   input.VersionCode,
		VersionName:   input.VersionName,
		ReleaseNote:   stringToPgtype(input.ReleaseNote),
		Environment:   db.ReleaseEnvironment(input.Environment),
		ApplicationID: uuidToPgtype(input.ApplicationID),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToRelease(&row), nil
}

// GetByID retrieves a release by ID.
func (r *ReleaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ApplicationRelease, error) {
	row, err := r.q.GetApplicationReleaseByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, translateError(err)
	}
	return rowToRelease(&row), nil
}

// ListByApplication lists all releases for an application.
func (r *ReleaseRepository) ListByApplication(ctx context.Context, appID uuid.UUID) ([]*domain.ApplicationRelease, error) {
	rows, err := r.q.ListReleasesByApplication(ctx, uuidToPgtype(appID))
	if err != nil {
		return nil, translateError(err)
	}

	releases := make([]*domain.ApplicationRelease, len(rows))
	for i, row := range rows {
		releases[i] = rowToRelease(&row)
	}
	return releases, nil
}

// ListByEnvironment lists releases by environment.
func (r *ReleaseRepository) ListByEnvironment(ctx context.Context, appID uuid.UUID, env domain.ReleaseEnvironment) ([]*domain.ApplicationRelease, error) {
	rows, err := r.q.ListReleasesByEnvironment(ctx, db.ListReleasesByEnvironmentParams{
		ApplicationID: uuidToPgtype(appID),
		Environment:   db.ReleaseEnvironment(env),
	})
	if err != nil {
		return nil, translateError(err)
	}

	releases := make([]*domain.ApplicationRelease, len(rows))
	for i, row := range rows {
		releases[i] = rowToRelease(&row)
	}
	return releases, nil
}

// GetLatestByEnvironment gets the latest release.
func (r *ReleaseRepository) GetLatestByEnvironment(ctx context.Context, appID uuid.UUID, env domain.ReleaseEnvironment) (*domain.ApplicationRelease, error) {
	row, err := r.q.GetLatestReleaseByEnvironment(ctx, db.GetLatestReleaseByEnvironmentParams{
		ApplicationID: uuidToPgtype(appID),
		Environment:   db.ReleaseEnvironment(env),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToRelease(&row), nil
}

// Update updates a release.
func (r *ReleaseRepository) Update(ctx context.Context, id uuid.UUID, title, releaseNote string) (*domain.ApplicationRelease, error) {
	row, err := r.q.UpdateRelease(ctx, db.UpdateReleaseParams{
		ID:          uuidToPgtype(id),
		Title:       title,
		ReleaseNote: stringToPgtype(releaseNote),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToRelease(&row), nil
}

// Promote updates the environment.
func (r *ReleaseRepository) Promote(ctx context.Context, id uuid.UUID, env domain.ReleaseEnvironment) (*domain.ApplicationRelease, error) {
	row, err := r.q.PromoteRelease(ctx, db.PromoteReleaseParams{
		ID:          uuidToPgtype(id),
		Environment: db.ReleaseEnvironment(env),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToRelease(&row), nil
}

// SoftDelete marks a release as deleted.
func (r *ReleaseRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.q.SoftDeleteApplicationRelease(ctx, uuidToPgtype(id))
	return translateError(err)
}

// VersionExists check - we don't have a specific SQL query for this yet,
// but we can use GetLatestReleaseByEnvironment and check if version matches,
// or better, handle the unique constraint error.
// For now, let's keep it simple and just implement it with a specific check if needed,
// but unique constraint is the source of truth.
func (r *ReleaseRepository) VersionExists(ctx context.Context, appID uuid.UUID, versionCode int32, env domain.ReleaseEnvironment) (bool, error) {
	// This is optional since DB constraint will catch it, but good for validation.
	// For now we'll return false and let the DB fail if duplicate.
	return false, nil
}

// ========== Transaction Methods ==========

// CreateTx creates a new release within a transaction.
func (r *ReleaseRepository) CreateTx(ctx context.Context, q *db.Queries, input domain.CreateReleaseInput) (*domain.ApplicationRelease, error) {
	row, err := q.CreateApplicationRelease(ctx, db.CreateApplicationReleaseParams{
		Title:         input.Title,
		VersionCode:   input.VersionCode,
		VersionName:   input.VersionName,
		ReleaseNote:   stringToPgtype(input.ReleaseNote),
		Environment:   db.ReleaseEnvironment(input.Environment),
		ApplicationID: uuidToPgtype(input.ApplicationID),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToRelease(&row), nil
}

// GetByIDTx retrieves a release by its ID within a transaction.
func (r *ReleaseRepository) GetByIDTx(ctx context.Context, q *db.Queries, id uuid.UUID) (*domain.ApplicationRelease, error) {
	row, err := q.GetApplicationReleaseByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, translateError(err)
	}
	return rowToRelease(&row), nil
}

// Helper to convert DB row to domain ApplicationRelease
func rowToRelease(row *db.ApplicationRelease) *domain.ApplicationRelease {
	return &domain.ApplicationRelease{
		ID:            pgtypeToUUID(row.ID),
		Title:         row.Title,
		VersionCode:   row.VersionCode,
		VersionName:   row.VersionName,
		ReleaseNote:   pgtypeToString(row.ReleaseNote),
		Environment:   domain.ReleaseEnvironment(row.Environment),
		ApplicationID: pgtypeToUUID(row.ApplicationID),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}
