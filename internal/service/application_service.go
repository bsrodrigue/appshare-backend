package service

import (
	"context"
	"fmt"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"github.com/google/uuid"
)

// ApplicationService handles application business logic.
type ApplicationService struct {
	// Services
	apkService *APKService

	// Repositories
	appRepo      repository.ApplicationRepository
	projectRepo  repository.ProjectRepository
	releaseRepo  repository.ReleaseRepository
	artifactRepo repository.ArtifactRepository
	txManager    *db.TxManager
}

// NewApplicationService creates a new ApplicationService.
func NewApplicationService(
	appRepo repository.ApplicationRepository,
	projectRepo repository.ProjectRepository,
	releaseRepo repository.ReleaseRepository,
	artifactRepo repository.ArtifactRepository,
	apkService *APKService,
	txManager *db.TxManager,
) *ApplicationService {
	return &ApplicationService{
		appRepo:      appRepo,
		projectRepo:  projectRepo,
		releaseRepo:  releaseRepo,
		artifactRepo: artifactRepo,
		apkService:   apkService,
		txManager:    txManager,
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

// Create application, release and artifact from a single first app binary
func (s *ApplicationService) CreateFromArtifact(ctx context.Context, userId uuid.UUID, input domain.CreateApplicationFromArtifactInput) (*domain.Application, error) {
	// Verify project exists and user is the owner
	project, err := s.projectRepo.GetByID(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userId {
		return nil, domain.ErrNotProjectOwner
	}

	// Parse the APK
	metadata, err := s.apkService.ExtractMetadataFromURL(ctx, input.ArtifactURL)
	if err != nil {
		return nil, err
	}

	// Check if package name is already taken
	exists, err := s.appRepo.PackageNameExists(ctx, metadata.PackageName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrPackageNameExists
	}

	// Transaction: Create Application, Release and Artifact
	var app *domain.Application
	err = s.txManager.WithTx(ctx, func(q *db.Queries) error {
		// 1. Create Application
		app, err = s.appRepo.CreateTx(ctx, q, domain.CreateApplicationInput{
			Title:       input.Title,
			PackageName: metadata.PackageName,
			ProjectID:   input.ProjectID,
		})
		if err != nil {
			return err
		}

		// 2. Create Initial Release
		release, err := s.releaseRepo.CreateTx(ctx, q, domain.CreateReleaseInput{
			ApplicationID: app.ID,
			Title:         fmt.Sprintf("Initial Release %s (%d)", metadata.VersionName, metadata.VersionCode),
			VersionCode:   int32(metadata.VersionCode),
			VersionName:   metadata.VersionName,
			ReleaseNote:   "Initial release from creation",
			Environment:   domain.EnvironmentProduction, // Default to production for first upload? Or based on input?
		})
		if err != nil {
			return err
		}

		// 3. Create Artifact
		_, err = s.artifactRepo.CreateTx(ctx, q, domain.CreateArtifactInput{
			ReleaseID: release.ID,
			FileURL:   input.ArtifactURL,
			SHA256:    metadata.SHA256,
			FileSize:  metadata.FileSize,
			FileType:  "application/vnd.android.package-archive",
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return app, nil
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
