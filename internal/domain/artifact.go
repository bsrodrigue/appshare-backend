package domain

import (
	"time"

	"github.com/google/uuid"
)

// Artifact represents a binary file associated with a release.
type Artifact struct {
	ID        uuid.UUID  `json:"id"`
	FileURL   string     `json:"file_url"`
	SHA256    string     `json:"sha256"`
	FileSize  int64      `json:"file_size"`
	FileType  string     `json:"file_type"`
	ABI       *string    `json:"abi,omitempty"`
	ReleaseID uuid.UUID  `json:"release_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CreateArtifactInput represents data needed to record a new artifact.
type CreateArtifactInput struct {
	FileURL   string
	SHA256    string
	FileSize  int64
	FileType  string
	ABI       *string
	ReleaseID uuid.UUID
}

// UploadURLResponse contains the signed URL and the storage path for the file.
type UploadURLResponse struct {
	UploadURL string `json:"upload_url" doc:"Signed URL for PUT upload"`
	FileURL   string `json:"file_url" doc:"Final public URL of the file"`
	Path      string `json:"path" doc:"Storage path/key"`
}
