package handler

import (
	"context"
	"net/http"

	"github.com/bsrodrigue/appshare-backend/internal/auth"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/service"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// ArtifactHandler handles artifact-related HTTP requests.
type ArtifactHandler struct {
	artifactService *service.ArtifactService
}

// NewArtifactHandler creates a new ArtifactHandler.
func NewArtifactHandler(artifactService *service.ArtifactService) *ArtifactHandler {
	return &ArtifactHandler{artifactService: artifactService}
}

// Register registers artifact routes with the API.
func (h *ArtifactHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-upload-url",
		Method:      http.MethodPost,
		Path:        "/artifacts/upload-url",
		Summary:     "Get Upload URL",
		Description: "Generate a signed URL for uploading an artifact. The client must perform a PUT request to the returned URL with the file content and 'Content-Type: application/octet-stream' header.",
		Tags:        []string{"Artifacts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.getUploadURL)

	huma.Register(api, huma.Operation{
		OperationID: "create-artifact",
		Method:      http.MethodPost,
		Path:        "/artifacts",
		Summary:     "Create Artifact",
		Description: "Record a new artifact in the database after it has been uploaded to storage.",
		Tags:        []string{"Artifacts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.createArtifact)

	huma.Register(api, huma.Operation{
		OperationID: "list-artifacts-by-release",
		Method:      http.MethodGet,
		Path:        "/releases/{release_id}/artifacts",
		Summary:     "List Artifacts",
		Description: "List all artifacts for a specific release.",
		Tags:        []string{"Artifacts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.listByRelease)
}

// ========== Request/Response Types ==========

type GetUploadURLInput struct {
	Body struct {
		ReleaseID uuid.UUID `json:"release_id" required:"true" doc:"Release ID"`
		Filename  string    `json:"filename" required:"true" doc:"Original filename"`
	}
}

type GetUploadURLOutput struct {
	Body ApiResponse[domain.UploadURLResponse]
}

type CreateArtifactInput struct {
	Body struct {
		ReleaseID uuid.UUID `json:"release_id" required:"true" doc:"Release ID"`
		FileURL   string    `json:"file_url" required:"true" doc:"Public URL of the uploaded file"`
		SHA256    string    `json:"sha256" required:"true" doc:"SHA256 hash of the file"`
		FileSize  int64     `json:"file_size" required:"true" doc:"File size in bytes"`
		FileType  string    `json:"file_type" required:"true" doc:"File MIME type"`
		ABI       *string   `json:"abi" doc:"System ABI (e.g. arm64-v8a)"`
	}
}

type CreateArtifactOutput struct {
	Body ApiResponse[domain.Artifact]
}

type ListArtifactsInput struct {
	ReleaseID uuid.UUID `path:"release_id" doc:"Release ID"`
}

type ListArtifactsOutput struct {
	Body ApiResponse[[]domain.Artifact]
}

// ========== Handlers ==========

func (h *ArtifactHandler) getUploadURL(ctx context.Context, input *GetUploadURLInput) (*GetUploadURLOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	res, err := h.artifactService.GetUploadURL(ctx, authUser.ID, input.Body.ReleaseID, input.Body.Filename)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &GetUploadURLOutput{
		Body: ok("Upload URL generated successfully", *res),
	}, nil
}

func (h *ArtifactHandler) createArtifact(ctx context.Context, input *CreateArtifactInput) (*CreateArtifactOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	artifact, err := h.artifactService.CreateArtifact(ctx, authUser.ID, domain.CreateArtifactInput{
		FileURL:   input.Body.FileURL,
		SHA256:    input.Body.SHA256,
		FileSize:  input.Body.FileSize,
		FileType:  input.Body.FileType,
		ABI:       input.Body.ABI,
		ReleaseID: input.Body.ReleaseID,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &CreateArtifactOutput{
		Body: created("Artifact recorded successfully", *artifact),
	}, nil
}

func (h *ArtifactHandler) listByRelease(ctx context.Context, input *ListArtifactsInput) (*ListArtifactsOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	artifacts, err := h.artifactService.ListByRelease(ctx, authUser.ID, input.ReleaseID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	// Convert pointer slice to value slice for response
	result := make([]domain.Artifact, len(artifacts))
	for i, a := range artifacts {
		result[i] = *a
	}

	return &ListArtifactsOutput{
		Body: ok("Artifacts retrieved successfully", result),
	}, nil
}
