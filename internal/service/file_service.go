package service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/storage"
	"github.com/google/uuid"
)

// FileService handles generic file-related business logic.
type FileService struct {
	storage storage.Storage
}

// NewFileService creates a new FileService.
func NewFileService(storage storage.Storage) *FileService {
	return &FileService{storage: storage}
}

// GetUploadURL generates a signed URL for a generic file upload.
func (s *FileService) GetUploadURL(ctx context.Context, userID uuid.UUID, filename string) (*domain.UploadURLResponse, error) {
	timestamp := time.Now().Unix()
	safeFilename := filepath.Base(filename)

	// Generic path: uploads/{user_id}/{timestamp}_{filename}
	storagePath := fmt.Sprintf("uploads/%s/%d_%s", userID, timestamp, safeFilename)

	// Generate signed URL (expires in 15 minutes)
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
