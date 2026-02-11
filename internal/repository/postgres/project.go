package postgres

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// ProjectRepository implements repository.ProjectRepository using PostgreSQL.
type ProjectRepository struct {
	q *db.Queries
}

// NewProjectRepository creates a new PostgreSQL project repository.
func NewProjectRepository(q *db.Queries) *ProjectRepository {
	return &ProjectRepository{q: q}
}

// ============================================================================
// Standard Methods (use internal queries)
// ============================================================================

// Create creates a new project.
func (r *ProjectRepository) Create(ctx context.Context, input domain.CreateProjectInput) (*domain.Project, error) {
	return r.CreateTx(ctx, r.q, input)
}

// GetByID retrieves a project by ID.
func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return r.GetByIDTx(ctx, r.q, id)
}

// ListByOwner retrieves all projects owned by a user.
func (r *ProjectRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Project, error) {
	rows, err := r.q.ListProjectsByOwner(ctx, uuidToPgtype(ownerID))
	if err != nil {
		return nil, translateError(err)
	}

	projects := make([]*domain.Project, len(rows))
	for i, row := range rows {
		projects[i] = projectToDoMain(&row)
	}
	return projects, nil
}

// UpdateTitle updates a project's title.
func (r *ProjectRepository) UpdateTitle(ctx context.Context, id uuid.UUID, title string) (*domain.Project, error) {
	row, err := r.q.UpdateProjectTitle(ctx, db.UpdateProjectTitleParams{
		ID:    uuidToPgtype(id),
		Title: title,
	})
	if err != nil {
		return nil, translateError(err)
	}
	return projectToDoMain(&row), nil
}

// UpdateDescription updates a project's description.
func (r *ProjectRepository) UpdateDescription(ctx context.Context, id uuid.UUID, description string) (*domain.Project, error) {
	row, err := r.q.UpdateProjectDescription(ctx, db.UpdateProjectDescriptionParams{
		ID:          uuidToPgtype(id),
		Description: description,
	})
	if err != nil {
		return nil, translateError(err)
	}
	return projectToDoMain(&row), nil
}

// Update updates both title and description.
func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, title, description string) (*domain.Project, error) {
	row, err := r.q.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          uuidToPgtype(id),
		Title:       title,
		Description: description,
	})
	if err != nil {
		return nil, translateError(err)
	}
	return projectToDoMain(&row), nil
}

// TransferOwnership transfers the project to a new owner.
func (r *ProjectRepository) TransferOwnership(ctx context.Context, id, newOwnerID uuid.UUID) (*domain.Project, error) {
	return r.TransferOwnershipTx(ctx, r.q, id, newOwnerID)
}

// SoftDelete marks a project as deleted.
func (r *ProjectRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.SoftDeleteTx(ctx, r.q, id)
}

// ============================================================================
// Transaction Methods (use provided queries)
// ============================================================================

// CreateTx creates a project within a transaction.
func (r *ProjectRepository) CreateTx(ctx context.Context, q *db.Queries, input domain.CreateProjectInput) (*domain.Project, error) {
	row, err := q.CreateProject(ctx, db.CreateProjectParams{
		Title:       input.Title,
		Description: input.Description,
		OwnerID:     uuidToPgtype(input.OwnerID),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return projectToDoMain(&row), nil
}

// GetByIDTx retrieves a project by ID within a transaction.
func (r *ProjectRepository) GetByIDTx(ctx context.Context, q *db.Queries, id uuid.UUID) (*domain.Project, error) {
	row, err := q.GetProjectByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, translateError(err)
	}
	return projectToDoMain(&row), nil
}

// TransferOwnershipTx transfers ownership within a transaction.
func (r *ProjectRepository) TransferOwnershipTx(ctx context.Context, q *db.Queries, id, newOwnerID uuid.UUID) (*domain.Project, error) {
	row, err := q.TransferProjectOwnership(ctx, db.TransferProjectOwnershipParams{
		ID:      uuidToPgtype(id),
		OwnerID: uuidToPgtype(newOwnerID),
	})
	if err != nil {
		return nil, translateError(err)
	}
	return projectToDoMain(&row), nil
}

// SoftDeleteTx marks a project as deleted within a transaction.
func (r *ProjectRepository) SoftDeleteTx(ctx context.Context, q *db.Queries, id uuid.UUID) error {
	_, err := q.SoftDeleteProject(ctx, uuidToPgtype(id))
	return translateError(err)
}

// ============================================================================
// Helper Functions
// ============================================================================

// projectToDoMain converts a db.Project to a domain.Project.
func projectToDoMain(row *db.Project) *domain.Project {
	return &domain.Project{
		ID:          pgtypeToUUID(row.ID),
		Title:       row.Title,
		Description: row.Description,
		OwnerID:     pgtypeToUUID(row.OwnerID),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
