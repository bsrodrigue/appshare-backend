package service

import (
	"context"
	"errors"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"github.com/google/uuid"
)

// ProjectService handles project-related business logic.
type ProjectService struct {
	projectRepo repository.ProjectRepository
	userRepo    repository.UserRepository
}

// NewProjectService creates a new ProjectService.
func NewProjectService(projectRepo repository.ProjectRepository, userRepo repository.UserRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

// Create creates a new project.
func (s *ProjectService) Create(ctx context.Context, input domain.CreateProjectInput) (*domain.Project, error) {
	// Verify owner exists
	_, err := s.userRepo.GetByID(ctx, input.OwnerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewValidationError("owner_id", "owner does not exist")
		}
		return nil, err
	}

	return s.projectRepo.Create(ctx, input)
}

// GetByID retrieves a project by ID.
func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

// ListByOwner retrieves all projects owned by a user.
func (s *ProjectService) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Project, error) {
	return s.projectRepo.ListByOwner(ctx, ownerID)
}

// Update updates a project. Only the owner can update their project.
func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateProjectInput, requesterID uuid.UUID) (*domain.Project, error) {
	// Get project to verify ownership
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if project.OwnerID != requesterID {
		return nil, domain.ErrForbidden
	}

	// Apply updates
	title := project.Title
	description := project.Description

	if input.Title != nil {
		title = *input.Title
	}
	if input.Description != nil {
		description = *input.Description
	}

	return s.projectRepo.Update(ctx, id, title, description)
}

// Delete soft-deletes a project. Only the owner can delete.
func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) error {
	// Get project to verify ownership
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check ownership
	if project.OwnerID != requesterID {
		return domain.ErrForbidden
	}

	return s.projectRepo.SoftDelete(ctx, id)
}

// TransferOwnership transfers project ownership to another user.
func (s *ProjectService) TransferOwnership(ctx context.Context, projectID, newOwnerID, requesterID uuid.UUID) (*domain.Project, error) {
	// Get project to verify current ownership
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Only current owner can transfer
	if project.OwnerID != requesterID {
		return nil, domain.ErrForbidden
	}

	// Verify new owner exists
	_, err = s.userRepo.GetByID(ctx, newOwnerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewValidationError("new_owner_id", "new owner does not exist")
		}
		return nil, err
	}

	return s.projectRepo.TransferOwnership(ctx, projectID, newOwnerID)
}
