package postgres

import (
	"context"
	"errors"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ApplicationRepository implements repository.ApplicationRepository using PostgreSQL.
type ApplicationRepository struct {
	q *db.Queries
}

// NewApplicationRepository creates a new PostgreSQL application repository.
func NewApplicationRepository(q *db.Queries) *ApplicationRepository {
	return &ApplicationRepository{q: q}
}

// Create creates a new application.
func (r *ApplicationRepository) Create(ctx context.Context, input domain.CreateApplicationInput) (*domain.Application, error) {
	row, err := r.q.CreateApplication(ctx, db.CreateApplicationParams{
		Title:       input.Title,
		PackageName: input.PackageName,
		Description: stringToPgtype(input.Description),
		ProjectID:   uuidToPgtype(input.ProjectID),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToApplication(&row), nil
}

// GetByID retrieves an application by ID.
func (r *ApplicationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	row, err := r.q.GetApplicationByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, translateError(err)
	}
	return rowToApplication(&row), nil
}

// GetByPackageName retrieves an application by package name.
func (r *ApplicationRepository) GetByPackageName(ctx context.Context, packageName string) (*domain.Application, error) {
	row, err := r.q.GetApplicationByPackageName(ctx, packageName)
	if err != nil {
		return nil, translateError(err)
	}
	return rowToApplication(&row), nil
}

// ListByProject retrieves all applications for a project.
func (r *ApplicationRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.Application, error) {
	rows, err := r.q.ListApplicationsByProject(ctx, uuidToPgtype(projectID))
	if err != nil {
		return nil, translateError(err)
	}

	apps := make([]*domain.Application, len(rows))
	for i, row := range rows {
		apps[i] = rowToApplication(&row)
	}
	return apps, nil
}

// Update updates an application.
func (r *ApplicationRepository) Update(ctx context.Context, id uuid.UUID, title, description string) (*domain.Application, error) {
	row, err := r.q.UpdateApplication(ctx, db.UpdateApplicationParams{
		ID:          uuidToPgtype(id),
		Title:       title,
		Description: stringToPgtype(description),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToApplication(&row), nil
}

// SoftDelete marks an application as deleted.
func (r *ApplicationRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.q.SoftDeleteApplication(ctx, uuidToPgtype(id))
	return translateError(err)
}

// PackageNameExists checks if a package name exists.
func (r *ApplicationRepository) PackageNameExists(ctx context.Context, packageName string) (bool, error) {
	_, err := r.q.GetApplicationByPackageName(ctx, packageName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, translateError(err)
	}
	return true, nil
}

// Helper to convert DB row to domain Application
func rowToApplication(row *db.Application) *domain.Application {
	return &domain.Application{
		ID:          pgtypeToUUID(row.ID),
		Title:       row.Title,
		PackageName: row.PackageName,
		Description: pgtypeToString(row.Description),
		ProjectID:   pgtypeToUUID(row.ProjectID),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
