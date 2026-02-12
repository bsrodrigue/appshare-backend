package handler

import (
	"context"
	"net/http"

	"github.com/bsrodrigue/appshare-backend/internal/auth"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/service"
	"github.com/danielgtaylor/huma/v2"
)

// FileHandler handles generic file-related HTTP requests.
type FileHandler struct {
	fileService *service.FileService
}

// NewFileHandler creates a new FileHandler.
func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// Register registers generic file routes with the API.
func (h *FileHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "upload-file",
		Method:      http.MethodPost,
		Path:        "/uploadFile",
		Summary:     "Get Generic Upload URL",
		Description: "Generate a signed URL for uploading any file. This is not tied to a specific release or artifact.",
		Tags:        []string{"Files"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.getUploadURL)
}

// ========== Request/Response Types ==========

type UploadFileInput struct {
	Body struct {
		Filename string `json:"filename" required:"true" doc:"Original filename"`
	}
}

type UploadFileOutput struct {
	Body ApiResponse[domain.UploadURLResponse]
}

// ========== Handlers ==========

func (h *FileHandler) getUploadURL(ctx context.Context, input *UploadFileInput) (*UploadFileOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	res, err := h.fileService.GetUploadURL(ctx, authUser.ID, input.Body.Filename)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &UploadFileOutput{
		Body: ok("Upload URL generated successfully", *res),
	}, nil
}
