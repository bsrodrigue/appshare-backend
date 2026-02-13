package repository

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// ApplicationRepository defines the interface for application data access.
type ApplicationRepository interface {
	// Create creates a new application.
	Create(ctx context.Context, input domain.CreateApplicationInput) (*domain.Application, error)

	// GetByID retrieves an application by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Application, error)

	// GetByPackageName retrieves an application by its package name.
	GetByPackageName(ctx context.Context, packageName string) (*domain.Application, error)

	// ListByProject retrieves all applications belonging to a project.
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.Application, error)

	// Update updates an application's title and description.
	Update(ctx context.Context, id uuid.UUID, title, description string) (*domain.Application, error)

	// SoftDelete marks an application as deleted.
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// PackageNameExists checks if a package name is already in use.
	PackageNameExists(ctx context.Context, packageName string) (bool, error)

	// ========== Transaction Methods ==========

	// CreateTx creates a new application within a transaction.
	CreateTx(ctx context.Context, q *db.Queries, input domain.CreateApplicationInput) (*domain.Application, error)
}
