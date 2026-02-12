package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/auth"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/service"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// ReleaseHandler handles application release HTTP requests.
type ReleaseHandler struct {
	releaseService *service.ReleaseService
}

// NewReleaseHandler creates a new ReleaseHandler.
func NewReleaseHandler(releaseService *service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{releaseService: releaseService}
}

// Register registers release routes with the API.
func (h *ReleaseHandler) Register(api huma.API) {
	// Protected routes (auth required)
	huma.Register(api, huma.Operation{
		OperationID: "create-release",
		Method:      http.MethodPost,
		Path:        "/applications/{app_id}/releases",
		Summary:     "Create Release",
		Description: "Create a new release for an application. Only the project owner can create releases.",
		Tags:        []string{"Releases"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.createRelease)

	huma.Register(api, huma.Operation{
		OperationID: "update-release",
		Method:      http.MethodPatch,
		Path:        "/releases/{id}",
		Summary:     "Update Release",
		Description: "Update a release's title and release notes.",
		Tags:        []string{"Releases"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.updateRelease)

	huma.Register(api, huma.Operation{
		OperationID: "promote-release",
		Method:      http.MethodPost,
		Path:        "/releases/{id}/promote",
		Summary:     "Promote Release",
		Description: "Promote a release to a different environment (e.g. development -> staging).",
		Tags:        []string{"Releases"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.promoteRelease)

	huma.Register(api, huma.Operation{
		OperationID: "delete-release",
		Method:      http.MethodDelete,
		Path:        "/releases/{id}",
		Summary:     "Delete Release",
		Description: "Mark a release as deleted.",
		Tags:        []string{"Releases"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.deleteRelease)

	huma.Register(api, huma.Operation{
		OperationID: "get-release",
		Method:      http.MethodGet,
		Path:        "/releases/{id}",
		Summary:     "Get Release",
		Description: "Get a specific release by ID.",
		Tags:        []string{"Releases"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.getRelease)

	huma.Register(api, huma.Operation{
		OperationID: "list-releases",
		Method:      http.MethodGet,
		Path:        "/applications/{app_id}/releases",
		Summary:     "List Releases",
		Description: "List all releases for an application.",
		Tags:        []string{"Releases"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.listReleases)

	huma.Register(api, huma.Operation{
		OperationID: "create-release-with-artifact",
		Method:      http.MethodPost,
		Path:        "/applications/{app_id}/releases/with-artifact",
		Summary:     "Create Release with Artifact",
		Description: "Create a new release and artifact by processing an existing uploaded file. The file must be a valid APK. Version info is extracted automatically.",
		Tags:        []string{"Releases"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.createReleaseWithArtifact)
}

// ========== Request/Response Types ==========

// ReleaseResponse represents a release in API responses.
type ReleaseResponse struct {
	ID            uuid.UUID                 `json:"id" doc:"Release unique ID"`
	Title         string                    `json:"title" doc:"Release title"`
	VersionCode   int32                     `json:"version_code" doc:"Numeric version code (e.g. 101)"`
	VersionName   string                    `json:"version_name" doc:"Semantic version string (e.g. 1.0.1)"`
	ReleaseNote   string                    `json:"release_note" doc:"Description of changes in this release"`
	Environment   domain.ReleaseEnvironment `json:"environment" doc:"Target environment"`
	ApplicationID uuid.UUID                 `json:"application_id" doc:"Parent application ID"`
	CreatedAt     time.Time                 `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt     time.Time                 `json:"updated_at" doc:"Last update timestamp"`
}

// CreateReleaseInput is the request for creating a release.
type CreateReleaseInput struct {
	AppID uuid.UUID `path:"app_id" doc:"Application ID"`
	Body  struct {
		Title       string                    `json:"title" required:"true" minLength:"3" maxLength:"100" doc:"Release title"`
		VersionCode int32                     `json:"version_code" required:"true" minimum:"1" doc:"Version code"`
		VersionName string                    `json:"version_name" required:"true" doc:"Version name"`
		ReleaseNote string                    `json:"release_note" maxLength:"2000" doc:"Release notes"`
		Environment domain.ReleaseEnvironment `json:"environment" required:"true" enum:"development,staging,production" doc:"Environment"`
	}
}

// CreateReleaseOutput is the response for creating a release.
type CreateReleaseOutput struct {
	Body ApiResponse[ReleaseResponse]
}

// UpdateReleaseInput is the request for updating a release.
type UpdateReleaseInput struct {
	ID   uuid.UUID `path:"id" doc:"Release ID"`
	Body struct {
		Title       string `json:"title" minLength:"3" maxLength:"100" doc:"Release title"`
		ReleaseNote string `json:"release_note" maxLength:"2000" doc:"Release notes"`
	}
}

// UpdateReleaseOutput is the response for updating a release.
type UpdateReleaseOutput struct {
	Body ApiResponse[ReleaseResponse]
}

// PromoteReleaseInput is the request for promoting a release.
type PromoteReleaseInput struct {
	ID   uuid.UUID `path:"id" doc:"Release ID"`
	Body struct {
		Environment domain.ReleaseEnvironment `json:"environment" required:"true" enum:"development,staging,production" doc:"New environment"`
	}
}

// PromoteReleaseOutput is the response for promoting a release.
type PromoteReleaseOutput struct {
	Body ApiResponse[ReleaseResponse]
}

// DeleteReleaseInput is the request for deleting a release.
type DeleteReleaseInput struct {
	ID uuid.UUID `path:"id" doc:"Release ID"`
}

// DeleteReleaseOutput is the response for deleting a release.
type DeleteReleaseOutput struct {
	Body ApiResponse[emptyData]
}

// GetReleaseInput is the request for getting a release.
type GetReleaseInput struct {
	ID uuid.UUID `path:"id" doc:"Release ID"`
}

// GetReleaseOutput is the response for getting a release.
type GetReleaseOutput struct {
	Body ApiResponse[ReleaseResponse]
}

// ListReleasesInput is the request for listing releases.
type ListReleasesInput struct {
	AppID uuid.UUID `path:"app_id" doc:"Application ID"`
}

// ListReleasesOutput is the response for listing releases.
type ListReleasesOutput struct {
	Body ApiResponse[[]ReleaseResponse]
}

// CreateReleaseWithArtifactInput is the request for creating a release with an artifact URL.
type CreateReleaseWithArtifactInput struct {
	AppID uuid.UUID `path:"app_id" doc:"Application ID"`
	Body  struct {
		ArtifactURL string                    `json:"artifact_url" required:"true" doc:"URL of the uploaded artifact (must be in our storage)"`
		ReleaseNote string                    `json:"release_note" maxLength:"2000" doc:"Release notes"`
		Environment domain.ReleaseEnvironment `json:"environment" required:"true" enum:"development,staging,production" doc:"Environment"`
	}
}

// CreateReleaseWithArtifactOutput is the response for creating a release with an artifact.
type CreateReleaseWithArtifactOutput struct {
	Body ApiResponse[ReleaseResponse]
}

// ========== Handlers ==========

func (h *ReleaseHandler) createRelease(ctx context.Context, input *CreateReleaseInput) (*CreateReleaseOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	release, err := h.releaseService.Create(ctx, authUser.ID, domain.CreateReleaseInput{
		ApplicationID: input.AppID,
		Title:         input.Body.Title,
		VersionCode:   input.Body.VersionCode,
		VersionName:   input.Body.VersionName,
		ReleaseNote:   input.Body.ReleaseNote,
		Environment:   input.Body.Environment,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &CreateReleaseOutput{
		Body: created("Release created successfully", toReleaseResponse(release)),
	}, nil
}

func (h *ReleaseHandler) updateRelease(ctx context.Context, input *UpdateReleaseInput) (*UpdateReleaseOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	release, err := h.releaseService.Update(ctx, authUser.ID, input.ID, domain.UpdateReleaseInput{
		Title:       input.Body.Title,
		ReleaseNote: input.Body.ReleaseNote,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &UpdateReleaseOutput{
		Body: ok("Release updated successfully", toReleaseResponse(release)),
	}, nil
}

func (h *ReleaseHandler) promoteRelease(ctx context.Context, input *PromoteReleaseInput) (*PromoteReleaseOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	release, err := h.releaseService.Promote(ctx, authUser.ID, input.ID, input.Body.Environment)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &PromoteReleaseOutput{
		Body: ok("Release promoted successfully", toReleaseResponse(release)),
	}, nil
}

func (h *ReleaseHandler) deleteRelease(ctx context.Context, input *DeleteReleaseInput) (*DeleteReleaseOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	err := h.releaseService.Delete(ctx, authUser.ID, input.ID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &DeleteReleaseOutput{
		Body: ok("Release deleted successfully", emptyData{}),
	}, nil
}

func (h *ReleaseHandler) getRelease(ctx context.Context, input *GetReleaseInput) (*GetReleaseOutput, error) {
	release, err := h.releaseService.GetByID(ctx, input.ID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &GetReleaseOutput{
		Body: ok("Release retrieved successfully", toReleaseResponse(release)),
	}, nil
}

func (h *ReleaseHandler) listReleases(ctx context.Context, input *ListReleasesInput) (*ListReleasesOutput, error) {
	releases, err := h.releaseService.ListByApplication(ctx, input.AppID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	responses := make([]ReleaseResponse, len(releases))
	for i, r := range releases {
		responses[i] = toReleaseResponse(r)
	}

	return &ListReleasesOutput{
		Body: ok("Releases retrieved successfully", responses),
	}, nil
}

func (h *ReleaseHandler) createReleaseWithArtifact(ctx context.Context, input *CreateReleaseWithArtifactInput) (*CreateReleaseWithArtifactOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	release, err := h.releaseService.CreateReleaseWithArtifactURL(ctx, authUser.ID, input.AppID, input.Body.ArtifactURL, input.Body.ReleaseNote, input.Body.Environment)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &CreateReleaseWithArtifactOutput{
		Body: created("Release and artifact created successfully", toReleaseResponse(release)),
	}, nil
}

// ========== Helpers ==========

func toReleaseResponse(r *domain.ApplicationRelease) ReleaseResponse {
	return ReleaseResponse{
		ID:            r.ID,
		Title:         r.Title,
		VersionCode:   r.VersionCode,
		VersionName:   r.VersionName,
		ReleaseNote:   r.ReleaseNote,
		Environment:   r.Environment,
		ApplicationID: r.ApplicationID,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
