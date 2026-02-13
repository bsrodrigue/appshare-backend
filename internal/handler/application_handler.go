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

// ApplicationHandler handles application-related HTTP requests.
type ApplicationHandler struct {
	appService *service.ApplicationService
}

// NewApplicationHandler creates a new ApplicationHandler.
func NewApplicationHandler(appService *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appService: appService}
}

// Register registers application routes with the API.
func (h *ApplicationHandler) Register(api huma.API) {
	// Protected routes (auth required)
	huma.Register(api, huma.Operation{
		OperationID: "create-application",
		Method:      http.MethodPost,
		Path:        "/projects/{project_id}/applications",
		Summary:     "Create Application",
		Description: "Create a new application within a project. Only the project owner can create applications.",
		Tags:        []string{"Applications"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.createApplication)

	huma.Register(api, huma.Operation{
		OperationID: "update-application",
		Method:      http.MethodPatch,
		Path:        "/applications/{id}",
		Summary:     "Update Application",
		Description: "Update an application's title and description. Package name cannot be changed.",
		Tags:        []string{"Applications"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.updateApplication)

	huma.Register(api, huma.Operation{
		OperationID: "delete-application",
		Method:      http.MethodDelete,
		Path:        "/applications/{id}",
		Summary:     "Delete Application",
		Description: "Mark an application as deleted.",
		Tags:        []string{"Applications"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.deleteApplication)

	huma.Register(api, huma.Operation{
		OperationID: "get-application",
		Method:      http.MethodGet,
		Path:        "/applications/{id}",
		Summary:     "Get Application",
		Description: "Get an application by ID.",
		Tags:        []string{"Applications"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.getApplication)

	huma.Register(api, huma.Operation{
		OperationID: "list-applications",
		Method:      http.MethodGet,
		Path:        "/projects/{project_id}/applications",
		Summary:     "List Applications",
		Description: "List all applications for a project.",
		Tags:        []string{"Applications"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.listApplications)

	huma.Register(api, huma.Operation{
		OperationID: "create-application-from-binary",
		Method:      http.MethodPost,
		Path:        "/create-application-from-binary",
		Summary:     "Create Application from Binary",
		Description: "Create a new application, initial release and artifact from a single APK binary. Automatically extracts package name and versioning.",
		Tags:        []string{"Applications"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.createApplicationFromBinary)
}

// ========== Request/Response Types ==========

// ApplicationResponse represents an application in API responses.
type ApplicationResponse struct {
	ID          uuid.UUID `json:"id" doc:"Application unique ID"`
	Title       string    `json:"title" doc:"Application title"`
	PackageName string    `json:"package_name" doc:"Unique package name (e.g. com.example.app)"`
	Description string    `json:"description" doc:"Application description"`
	ProjectID   uuid.UUID `json:"project_id" doc:"ID of the parent project"`
	CreatedAt   time.Time `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt   time.Time `json:"updated_at" doc:"Last update timestamp"`
}

// CreateApplicationInput is the request for creating an application.
type CreateApplicationInput struct {
	ProjectID uuid.UUID `path:"project_id" doc:"Project ID"`
	Body      struct {
		Title       string `json:"title" required:"true" minLength:"3" maxLength:"100" doc:"Application title"`
		PackageName string `json:"package_name" required:"true" minLength:"3" maxLength:"255" doc:"Unique package name"`
		Description string `json:"description" maxLength:"1000" doc:"Application description"`
	}
}

// CreateApplicationOutput is the response for creating an application.
type CreateApplicationOutput struct {
	Body ApiResponse[ApplicationResponse]
}

// UpdateApplicationInput is the request for updating an application.
type UpdateApplicationInput struct {
	ID   uuid.UUID `path:"id" doc:"Application ID"`
	Body struct {
		Title       string `json:"title" minLength:"3" maxLength:"100" doc:"Application title"`
		Description string `json:"description" maxLength:"1000" doc:"Application description"`
	}
}

// UpdateApplicationOutput is the response for updating an application.
type UpdateApplicationOutput struct {
	Body ApiResponse[ApplicationResponse]
}

// DeleteApplicationInput is the request for deleting an application.
type DeleteApplicationInput struct {
	ID uuid.UUID `path:"id" doc:"Application ID"`
}

// DeleteApplicationOutput is the response for deleting an application.
type DeleteApplicationOutput struct {
	Body ApiResponse[emptyData]
}

// GetApplicationInput is the request for getting an application.
type GetApplicationInput struct {
	ID uuid.UUID `path:"id" doc:"Application ID"`
}

// GetApplicationOutput is the response for getting an application.
type GetApplicationOutput struct {
	Body ApiResponse[ApplicationResponse]
}

// ListApplicationsInput is the request for listing applications.
type ListApplicationsInput struct {
	ProjectID uuid.UUID `path:"project_id" doc:"Project ID"`
}

// ListApplicationsOutput is the response for listing applications.
type ListApplicationsOutput struct {
	Body ApiResponse[[]ApplicationResponse]
}

// CreateApplicationFromBinaryInput is the request for creating an application from a binary.
type CreateApplicationFromBinaryInput struct {
	Body struct {
		ProjectID   uuid.UUID `json:"project_id" required:"true" doc:"Project ID"`
		Title       string    `json:"title" required:"true" minLength:"3" maxLength:"100" doc:"Application title"`
		ArtifactURL string    `json:"artifact_url" required:"true" doc:"URL of the artifact in storage"`
	}
}

// CreateApplicationFromBinaryOutput is the response for creating an application from a binary.
type CreateApplicationFromBinaryOutput struct {
	Body ApiResponse[ApplicationResponse]
}

// ========== Handlers ==========

func (h *ApplicationHandler) createApplication(ctx context.Context, input *CreateApplicationInput) (*CreateApplicationOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	app, err := h.appService.Create(ctx, authUser.ID, domain.CreateApplicationInput{
		ProjectID:   input.ProjectID,
		Title:       input.Body.Title,
		PackageName: input.Body.PackageName,
		Description: input.Body.Description,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &CreateApplicationOutput{
		Body: created("Application created successfully", toApplicationResponse(app)),
	}, nil
}

func (h *ApplicationHandler) updateApplication(ctx context.Context, input *UpdateApplicationInput) (*UpdateApplicationOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	app, err := h.appService.Update(ctx, authUser.ID, input.ID, domain.UpdateApplicationInput{
		Title:       input.Body.Title,
		Description: input.Body.Description,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &UpdateApplicationOutput{
		Body: ok("Application updated successfully", toApplicationResponse(app)),
	}, nil
}

func (h *ApplicationHandler) deleteApplication(ctx context.Context, input *DeleteApplicationInput) (*DeleteApplicationOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	err := h.appService.Delete(ctx, authUser.ID, input.ID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &DeleteApplicationOutput{
		Body: ok("Application deleted successfully", emptyData{}),
	}, nil
}

func (h *ApplicationHandler) getApplication(ctx context.Context, input *GetApplicationInput) (*GetApplicationOutput, error) {
	app, err := h.appService.GetByID(ctx, input.ID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &GetApplicationOutput{
		Body: ok("Application retrieved successfully", toApplicationResponse(app)),
	}, nil
}

func (h *ApplicationHandler) listApplications(ctx context.Context, input *ListApplicationsInput) (*ListApplicationsOutput, error) {
	apps, err := h.appService.ListByProject(ctx, input.ProjectID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	responses := make([]ApplicationResponse, len(apps))
	for i, app := range apps {
		responses[i] = toApplicationResponse(app)
	}

	return &ListApplicationsOutput{
		Body: ok("Applications retrieved successfully", responses),
	}, nil
}

func (h *ApplicationHandler) createApplicationFromBinary(ctx context.Context, input *CreateApplicationFromBinaryInput) (*CreateApplicationFromBinaryOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	app, err := h.appService.CreateFromArtifact(ctx, authUser.ID, domain.CreateApplicationFromArtifactInput{
		ProjectID:   input.Body.ProjectID,
		Title:       input.Body.Title,
		ArtifactURL: input.Body.ArtifactURL,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &CreateApplicationFromBinaryOutput{
		Body: created("Application profile created from binary successfully", toApplicationResponse(app)),
	}, nil
}

// ========== Helpers ==========

func toApplicationResponse(app *domain.Application) ApplicationResponse {
	return ApplicationResponse{
		ID:          app.ID,
		Title:       app.Title,
		PackageName: app.PackageName,
		Description: app.Description,
		ProjectID:   app.ProjectID,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	}
}
