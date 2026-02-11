package service

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"github.com/google/uuid"
)

// ReleaseService handles release business logic.
type ReleaseService struct {
	releaseRepo repository.ReleaseRepository
	appRepo     repository.ApplicationRepository
	projectRepo repository.ProjectRepository
}

// NewReleaseService creates a new ReleaseService.
func NewReleaseService(
	releaseRepo repository.ReleaseRepository,
	appRepo repository.ApplicationRepository,
	projectRepo repository.ProjectRepository,
) *ReleaseService {
	return &ReleaseService{
		releaseRepo: releaseRepo,
		appRepo:     appRepo,
		projectRepo: projectRepo,
	}
}

// Create creates a new release for an application.
func (s *ReleaseService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateReleaseInput) (*domain.ApplicationRelease, error) {
	// Verify application exists and user owns the project
	app, err := s.appRepo.GetByID(ctx, input.ApplicationID)
	if err != nil {
		return nil, err
	}

	project, err := s.projectRepo.GetByID(ctx, app.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, domain.ErrNotProjectOwner
	}

	// Create release (DB unique constraint will handle duplicate version_code/environment)
	return s.releaseRepo.Create(ctx, input)
}

// Update updates a release.
func (s *ReleaseService) Update(ctx context.Context, userID uuid.UUID, releaseID uuid.UUID, input domain.UpdateReleaseInput) (*domain.ApplicationRelease, error) {
	// Get release and verify ownership
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		return nil, err
	}

	app, err := s.appRepo.GetByID(ctx, release.ApplicationID)
	if err != nil {
		return nil, err
	}

	project, err := s.projectRepo.GetByID(ctx, app.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, domain.ErrNotProjectOwner
	}

	return s.releaseRepo.Update(ctx, releaseID, input.Title, input.ReleaseNote)
}

// Promote promotes a release to another environment.
func (s *ReleaseService) Promote(ctx context.Context, userID uuid.UUID, releaseID uuid.UUID, env domain.ReleaseEnvironment) (*domain.ApplicationRelease, error) {
	// Ownership check
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		return nil, err
	}

	app, err := s.appRepo.GetByID(ctx, release.ApplicationID)
	if err != nil {
		return nil, err
	}

	project, err := s.projectRepo.GetByID(ctx, app.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, domain.ErrNotProjectOwner
	}

	return s.releaseRepo.Promote(ctx, releaseID, env)
}

// Delete deletes a release.
func (s *ReleaseService) Delete(ctx context.Context, userID uuid.UUID, releaseID uuid.UUID) error {
	// Ownership check
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		return err
	}

	app, err := s.appRepo.GetByID(ctx, release.ApplicationID)
	if err != nil {
		return err
	}

	project, err := s.projectRepo.GetByID(ctx, app.ProjectID)
	if err != nil {
		return err
	}

	if project.OwnerID != userID {
		return domain.ErrNotProjectOwner
	}

	return s.releaseRepo.SoftDelete(ctx, releaseID)
}

// GetByID retrieves a release by ID.
func (s *ReleaseService) GetByID(ctx context.Context, releaseID uuid.UUID) (*domain.ApplicationRelease, error) {
	return s.releaseRepo.GetByID(ctx, releaseID)
}

// ListByApplication lists all releases for an application.
func (s *ReleaseService) ListByApplication(ctx context.Context, appID uuid.UUID) ([]*domain.ApplicationRelease, error) {
	return s.releaseRepo.ListByApplication(ctx, appID)
}

// GetLatestByEnvironment gets the latest release.
func (s *ReleaseService) GetLatestByEnvironment(ctx context.Context, appID uuid.UUID, env domain.ReleaseEnvironment) (*domain.ApplicationRelease, error) {
	return s.releaseRepo.GetLatestByEnvironment(ctx, appID, env)
}
