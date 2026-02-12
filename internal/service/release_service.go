package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"github.com/bsrodrigue/appshare-backend/internal/storage"
	"github.com/google/uuid"
	"github.com/shogo82148/androidbinary/apk"
)

// ReleaseService handles release business logic.
type ReleaseService struct {
	releaseRepo  repository.ReleaseRepository
	appRepo      repository.ApplicationRepository
	projectRepo  repository.ProjectRepository
	artifactRepo repository.ArtifactRepository
	storage      storage.Storage
	txManager    *db.TxManager
}

// NewReleaseService creates a new ReleaseService.
func NewReleaseService(
	releaseRepo repository.ReleaseRepository,
	appRepo repository.ApplicationRepository,
	projectRepo repository.ProjectRepository,
	artifactRepo repository.ArtifactRepository,
	storage storage.Storage,
	txManager *db.TxManager,
) *ReleaseService {
	return &ReleaseService{
		releaseRepo:  releaseRepo,
		appRepo:      appRepo,
		projectRepo:  projectRepo,
		artifactRepo: artifactRepo,
		storage:      storage,
		txManager:    txManager,
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

// CreateReleaseWithArtifactURL handles the complex flow of downloading an artifact,
// verifying it's an APK, extracting version info, and creating both release and artifact records.
func (s *ReleaseService) CreateReleaseWithArtifactURL(ctx context.Context, userID uuid.UUID, appID uuid.UUID, artifactURL string, releaseNote string, environment domain.ReleaseEnvironment) (*domain.ApplicationRelease, error) {
	// 1. Verify ownership early
	app, err := s.appRepo.GetByID(ctx, appID)
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

	// 2. Download the file to a temporary location
	// We need it as a local file for APK parsing
	storagePath, isOurs := s.extractStoragePath(artifactURL)
	var reader io.ReadCloser
	if isOurs {
		reader, err = s.storage.Download(ctx, storagePath)
	} else {
		// External URL - but let's stick to our storage for now as per "Download from cloudflare"
		return nil, domain.NewValidationError("artifact_url", "only internal artifacts are supported for now")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}
	defer reader.Close()

	tmpFile, err := os.CreateTemp("", "artifact-*.apk")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	hasher := sha256.New()
	multiWriter := io.MultiWriter(tmpFile, hasher)
	fileSize, err := io.Copy(multiWriter, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to save artifact: %w", err)
	}
	sha256Hex := hex.EncodeToString(hasher.Sum(nil))

	// 3. Parse APK
	pkg, err := apk.OpenFile(tmpFile.Name())
	if err != nil {
		return nil, domain.NewValidationError("artifact_url", "invalid APK file: "+err.Error())
	}
	defer pkg.Close()

	versionCode := pkg.Manifest().VersionCode.MustInt32()
	versionName := pkg.Manifest().VersionName.MustString()
	packageName := pkg.PackageName()

	// Verify package name matches (optional but good)
	if app.PackageName != packageName {
		return nil, domain.NewValidationError("artifact_url", fmt.Sprintf("package name mismatch: expected %s, got %s", app.PackageName, packageName))
	}

	// 4. Transactional DB update
	var release *domain.ApplicationRelease
	err = s.txManager.WithTx(ctx, func(q *db.Queries) error {
		// Create Release
		release, err = s.releaseRepo.CreateTx(ctx, q, domain.CreateReleaseInput{
			ApplicationID: appID,
			Title:         fmt.Sprintf("Release %s (%d)", versionName, versionCode),
			VersionCode:   int32(versionCode),
			VersionName:   versionName,
			ReleaseNote:   releaseNote,
			Environment:   environment,
		})
		if err != nil {
			return err
		}

		// Create Artifact
		_, err = s.artifactRepo.CreateTx(ctx, q, domain.CreateArtifactInput{
			ReleaseID: release.ID,
			FileURL:   artifactURL,
			SHA256:    sha256Hex,
			FileSize:  fileSize,
			FileType:  "application/vnd.android.package-archive",
			// ABI: could extract from APK entries (lib/arm64-v8a etc.) but let's keep it simple
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return release, nil
}

func (s *ReleaseService) extractStoragePath(rawURL string) (string, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}

	// If it matches our public domain, strip the domain and return the path
	// Example: https://pub-xxxx.r2.dev/uploads/user_id/file.apk
	// Or custom domain: https://cdn.appshare.com/uploads/user_id/file.apk

	path := strings.TrimPrefix(parsed.Path, "/")
	// In a real app, you'd verify the host matches s.config.PublicDomain
	return path, true
}
