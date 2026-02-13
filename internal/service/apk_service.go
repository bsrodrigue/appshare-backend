package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/storage"
	"github.com/shogo82148/androidbinary/apk"
)

// APKService handles APK file operations.
type APKService struct {
	storage storage.Storage
}

// NewAPKService creates a new APKService.
func NewAPKService(storage storage.Storage) *APKService {
	return &APKService{
		storage: storage,
	}
}

// ExtractMetadataFromURL downloads an APK from a URL and extracts its metadata.
func (s *APKService) ExtractMetadataFromURL(ctx context.Context, artifactURL string) (*domain.ApplicationMetadata, error) {
	storagePath, isOurs := s.storage.ExtractStoragePath(artifactURL)
	if !isOurs {
		slog.Warn("Attempted to extract metadata from non-internal URL", "url", artifactURL)
		return nil, domain.NewValidationError("artifact_url", "only internal artifacts are supported for now")
	}

	slog.Debug("Downloading artifact for metadata extraction", "path", storagePath)
	reader, err := s.storage.Download(ctx, storagePath)
	if err != nil {
		slog.Error("Failed to download artifact", "path", storagePath, "error", err)
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}
	defer reader.Close()

	// Create a temporary file to store the APK
	// Warning, can be dangerous with tmpfs (memory exhaustion)
	tmpFile, err := os.CreateTemp("", "artifact-*.apk")
	if err != nil {
		slog.Error("Failed to create temp file", "error", err)
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	slog.Debug("Hashing and saving artifact to disk", "tempFile", tmpFile.Name())
	hasher := sha256.New()
	multiWriter := io.MultiWriter(tmpFile, hasher)
	fileSize, err := io.Copy(multiWriter, reader)
	if err != nil {
		slog.Error("Failed to save artifact to disk", "error", err)
		return nil, fmt.Errorf("failed to save artifact to temp file: %w", err)
	}
	sha256Hex := hex.EncodeToString(hasher.Sum(nil))
	slog.Debug("Artifact saved", "size", fileSize, "sha256", sha256Hex)

	// Open the APK file for reading
	apkFile, err := apk.OpenFile(tmpFile.Name())
	if err != nil {
		slog.Error("Failed to open APK file", "tempFile", tmpFile.Name(), "error", err)
		return nil, fmt.Errorf("failed to open APK file: %w", err)
	}
	defer apkFile.Close()

	// Extract metadata from the APK
	manifest := apkFile.Manifest()
	packageName := apkFile.PackageName()
	versionCode := manifest.VersionCode.MustInt32()
	versionName := manifest.VersionName.MustString()

	// Determine platform - it's an APK, so it's android
	platform := "android"

	// Get minimum and target SDK versions
	var minSdkVersion int32
	var targetSdkVersion int32

	minSdkVersion = manifest.SDK.Min.MustInt32()
	targetSdkVersion = manifest.SDK.Target.MustInt32()

	// Create metadata object
	metadata := &domain.ApplicationMetadata{
		PackageName:      packageName,
		VersionCode:      int64(versionCode),
		VersionName:      versionName,
		MinSdkVersion:    int(minSdkVersion),
		TargetSdkVersion: int(targetSdkVersion),
		Architecture:     "universal",
		Platform:         platform,
		SHA256:           sha256Hex,
		FileSize:         fileSize,
	}

	return metadata, nil
}
