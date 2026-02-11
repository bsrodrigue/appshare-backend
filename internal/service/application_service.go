package service

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"github.com/google/uuid"
)

// ApplicationService handles application business logic.
type ApplicationService struct {
	appRepo     repository.ApplicationRepository
	projectRepo repository.ProjectRepository
}

// NewApplicationService creates a new ApplicationService.
func NewApplicationService(appRepo repository.ApplicationRepository, projectRepo repository.ProjectRepository) *ApplicationService {
	return &ApplicationService{
		appRepo:     appRepo,
		projectRepo: projectRepo,
	}
}

// Create creates a new application within a project.
func (s *ApplicationService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateApplicationInput) (*domain.Application, error) {
	// Verify project exists and user is the owner
	project, err := s.projectRepo.GetByID(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, domain.ErrNotProjectOwner
	}

	// Check if package name is already taken
	exists, err := s.appRepo.PackageNameExists(ctx, input.PackageName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrPackageNameExists
	}

	return s.appRepo.Create(ctx, input)
}

// Update updates an application.
func (s *ApplicationService) Update(ctx context.Context, userID uuid.UUID, appID uuid.UUID, input domain.UpdateApplicationInput) (*domain.Application, error) {
	// Get app
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// Verify ownership through project
	project, err := s.projectRepo.GetByID(ctx, app.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, domain.ErrNotProjectOwner
	}

	return s.appRepo.Update(ctx, appID, input.Title, input.Description)
}

// Delete deletes an application.
func (s *ApplicationService) Delete(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error {
	// Get app
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return err
	}

	// Verify ownership
	project, err := s.projectRepo.GetByID(ctx, app.ProjectID)
	if err != nil {
		return err
	}

	if project.OwnerID != userID {
		return domain.ErrNotProjectOwner
	}

	return s.appRepo.SoftDelete(ctx, appID)
}

// GetByID retrieves an application by ID.
func (s *ApplicationService) GetByID(ctx context.Context, appID uuid.UUID) (*domain.Application, error) {
	return s.appRepo.GetByID(ctx, appID)
}

// ListByProject lists all applications for a project.
func (s *ApplicationService) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.Application, error) {
	return s.appRepo.ListByProject(ctx, projectID)
}
