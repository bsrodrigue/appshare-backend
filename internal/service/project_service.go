package service

import (
	"context"
	"errors"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"github.com/google/uuid"
)

// ProjectService handles project-related business logic.
type ProjectService struct {
	projectRepo repository.ProjectRepository
	userRepo    repository.UserRepository
	txManager   *db.TxManager
}

// NewProjectService creates a new ProjectService.
func NewProjectService(
	projectRepo repository.ProjectRepository,
	userRepo repository.UserRepository,
	txManager *db.TxManager,
) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
		txManager:   txManager,
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
		return nil, domain.WrapError(domain.CodeInternal, "failed to verify owner", err)
	}

	project, err := s.projectRepo.Create(ctx, input)
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to create project", err)
	}

	return project, nil
}

// GetByID retrieves a project by ID.
func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}
	return project, nil
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
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}

	// Check ownership
	if project.OwnerID != requesterID {
		return nil, domain.ErrNotProjectOwner
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
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrProjectNotFound
		}
		return err
	}

	// Check ownership
	if project.OwnerID != requesterID {
		return domain.ErrNotProjectOwner
	}

	return s.projectRepo.SoftDelete(ctx, id)
}

// TransferOwnership transfers project ownership to another user.
// This is a transactional operation as it may involve multiple updates.
func (s *ProjectService) TransferOwnership(ctx context.Context, projectID, newOwnerID, requesterID uuid.UUID) (*domain.Project, error) {
	var result *domain.Project

	err := s.txManager.WithTx(ctx, func(q *db.Queries) error {
		// Get project to verify current ownership
		project, err := s.projectRepo.GetByIDTx(ctx, q, projectID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrProjectNotFound
			}
			return err
		}

		// Only current owner can transfer
		if project.OwnerID != requesterID {
			return domain.ErrNotProjectOwner
		}

		// Verify new owner exists
		_, err = s.userRepo.GetByIDTx(ctx, q, newOwnerID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.NewValidationError("new_owner_id", "new owner does not exist")
			}
			return err
		}

		// Transfer ownership
		result, err = s.projectRepo.TransferOwnershipTx(ctx, q, projectID, newOwnerID)
		if err != nil {
			return domain.WrapError(domain.CodeInternal, "failed to transfer ownership", err)
		}

		// Future: Create membership for previous owner to keep access
		// _, err = q.CreateMembership(ctx, ...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
