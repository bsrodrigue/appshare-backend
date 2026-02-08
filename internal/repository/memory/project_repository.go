package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// ProjectRepository implements repository.ProjectRepository in memory.
type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[uuid.UUID]*domain.Project
}

// NewProjectRepository creates a new in-memory project repository.
func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		projects: make(map[uuid.UUID]*domain.Project),
	}
}

// Reset clears all data in the repository.
func (r *ProjectRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.projects = make(map[uuid.UUID]*domain.Project)
}

// ============================================================================
// Standard Methods
// ============================================================================

func (r *ProjectRepository) Create(ctx context.Context, input domain.CreateProjectInput) (*domain.Project, error) {
	return r.CreateTx(ctx, nil, input)
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return r.GetByIDTx(ctx, nil, id)
}

func (r *ProjectRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	projects := make([]*domain.Project, 0)
	for _, p := range r.projects {
		if p.OwnerID == ownerID {
			project := *p
			projects = append(projects, &project)
		}
	}

	// Sort by CreatedAt descending
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].CreatedAt.After(projects[j].CreatedAt)
	})

	return projects, nil
}

func (r *ProjectRepository) UpdateTitle(ctx context.Context, id uuid.UUID, title string) (*domain.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.projects[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}

	p.Title = title
	p.UpdatedAt = time.Now()
	project := *p
	return &project, nil
}

func (r *ProjectRepository) UpdateDescription(ctx context.Context, id uuid.UUID, description string) (*domain.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.projects[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}

	p.Description = description
	p.UpdatedAt = time.Now()
	project := *p
	return &project, nil
}

func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, title, description string) (*domain.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.projects[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}

	p.Title = title
	p.Description = description
	p.UpdatedAt = time.Now()
	project := *p
	return &project, nil
}

func (r *ProjectRepository) TransferOwnership(ctx context.Context, id, newOwnerID uuid.UUID) (*domain.Project, error) {
	return r.TransferOwnershipTx(ctx, nil, id, newOwnerID)
}

func (r *ProjectRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.SoftDeleteTx(ctx, nil, id)
}

// ============================================================================
// Transaction Methods
// ============================================================================

func (r *ProjectRepository) CreateTx(ctx context.Context, q *db.Queries, input domain.CreateProjectInput) (*domain.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New()
	now := time.Now()
	project := &domain.Project{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		OwnerID:     input.OwnerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.projects[id] = project
	p := *project
	return &p, nil
}

func (r *ProjectRepository) GetByIDTx(ctx context.Context, q *db.Queries, id uuid.UUID) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.projects[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}

	project := *p
	return &project, nil
}

func (r *ProjectRepository) TransferOwnershipTx(ctx context.Context, q *db.Queries, id, newOwnerID uuid.UUID) (*domain.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.projects[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}

	p.OwnerID = newOwnerID
	p.UpdatedAt = time.Now()
	project := *p
	return &project, nil
}

func (r *ProjectRepository) SoftDeleteTx(ctx context.Context, q *db.Queries, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.projects[id]; !ok {
		return domain.ErrProjectNotFound
	}

	delete(r.projects, id)
	return nil
}
