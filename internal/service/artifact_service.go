package service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"github.com/bsrodrigue/appshare-backend/internal/storage"
	"github.com/google/uuid"
)

// ArtifactService handles artifact-related business logic.
type ArtifactService struct {
	artifactRepo repository.ArtifactRepository
	releaseRepo  repository.ReleaseRepository
	appRepo      repository.ApplicationRepository
	projectRepo  repository.ProjectRepository
	storage      storage.Storage
}

// NewArtifactService creates a new ArtifactService.
func NewArtifactService(
	artifactRepo repository.ArtifactRepository,
	releaseRepo repository.ReleaseRepository,
	appRepo repository.ApplicationRepository,
	projectRepo repository.ProjectRepository,
	storage storage.Storage,
) *ArtifactService {
	return &ArtifactService{
		artifactRepo: artifactRepo,
		releaseRepo:  releaseRepo,
		appRepo:      appRepo,
		projectRepo:  projectRepo,
		storage:      storage,
	}
}

// GetUploadURL generates a signed URL for uploading an artifact.
func (s *ArtifactService) GetUploadURL(ctx context.Context, userID uuid.UUID, releaseID uuid.UUID, filename string) (*domain.UploadURLResponse, error) {
	// 1. Verify ownership
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
		return nil, domain.WrapError(domain.CodeNotProjectOwner, fmt.Sprintf("access denied: user %s is not the owner of project %s", userID, project.ID), domain.ErrNotProjectOwner)
	}

	// 2. Generate storage path
	// Structure: apps/{app_id}/releases/{release_id}/{timestamp}_{filename}
	timestamp := time.Now().Unix()
	safeFilename := filepath.Base(filename)
	storagePath := fmt.Sprintf("apps/%s/releases/%s/%d_%s", app.ID, release.ID, timestamp, safeFilename)

	// 3. Generate signed URL (expires in 15 minutes)
	uploadURL, err := s.storage.GenerateUploadURL(ctx, storagePath, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &domain.UploadURLResponse{
		UploadURL: uploadURL,
		FileURL:   s.storage.GetPublicURL(storagePath),
		Path:      storagePath,
	}, nil
}

// CreateArtifact records a new artifact in the database.
func (s *ArtifactService) CreateArtifact(ctx context.Context, userID uuid.UUID, input domain.CreateArtifactInput) (*domain.Artifact, error) {
	// Ownership check
	release, err := s.releaseRepo.GetByID(ctx, input.ReleaseID)
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

	return s.artifactRepo.Create(ctx, input)
}
