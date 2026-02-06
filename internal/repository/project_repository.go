package repository

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// ProjectRepository defines the interface for project data access.
type ProjectRepository interface {
	// ========== Standard Methods (auto-commit) ==========

	// Create creates a new project.
	Create(ctx context.Context, input domain.CreateProjectInput) (*domain.Project, error)

	// GetByID retrieves a project by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)

	// ListByOwner retrieves all projects owned by a user.
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Project, error)

	// UpdateTitle updates a project's title.
	UpdateTitle(ctx context.Context, id uuid.UUID, title string) (*domain.Project, error)

	// UpdateDescription updates a project's description.
	UpdateDescription(ctx context.Context, id uuid.UUID, description string) (*domain.Project, error)

	// Update updates both title and description.
	Update(ctx context.Context, id uuid.UUID, title, description string) (*domain.Project, error)

	// TransferOwnership transfers the project to a new owner.
	TransferOwnership(ctx context.Context, id, newOwnerID uuid.UUID) (*domain.Project, error)

	// SoftDelete marks a project as deleted.
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// ========== Transaction Methods ==========

	// CreateTx creates a project within a transaction.
	CreateTx(ctx context.Context, q *db.Queries, input domain.CreateProjectInput) (*domain.Project, error)

	// GetByIDTx retrieves a project by ID within a transaction.
	GetByIDTx(ctx context.Context, q *db.Queries, id uuid.UUID) (*domain.Project, error)

	// TransferOwnershipTx transfers ownership within a transaction.
	TransferOwnershipTx(ctx context.Context, q *db.Queries, id, newOwnerID uuid.UUID) (*domain.Project, error)

	// SoftDeleteTx marks a project as deleted within a transaction.
	SoftDeleteTx(ctx context.Context, q *db.Queries, id uuid.UUID) error
}
