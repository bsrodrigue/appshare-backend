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

// ProjectHandler handles project-related HTTP requests.
type ProjectHandler struct {
	projectService *service.ProjectService
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// Register registers all project routes with the API.
// All project routes require authentication.
func (h *ProjectHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-my-projects",
		Method:      http.MethodGet,
		Path:        "/projects",
		Summary:     "List My Projects",
		Description: "Retrieve all projects owned by the authenticated user.",
		Tags:        []string{"Projects"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, h.listMyProjects)

	huma.Register(api, huma.Operation{
		OperationID: "get-project",
		Method:      http.MethodGet,
		Path:        "/projects/{id}",
		Summary:     "Get Project",
		Description: "Retrieve a specific project by ID. Only the owner can access their project.",
		Tags:        []string{"Projects"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, h.getProject)

	huma.Register(api, huma.Operation{
		OperationID: "create-project",
		Method:      http.MethodPost,
		Path:        "/projects",
		Summary:     "Create Project",
		Description: "Create a new project. The authenticated user becomes the owner.",
		Tags:        []string{"Projects"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, h.createProject)

	huma.Register(api, huma.Operation{
		OperationID: "update-project",
		Method:      http.MethodPatch,
		Path:        "/projects/{id}",
		Summary:     "Update Project",
		Description: "Update a project's title and/or description. Only the owner can update.",
		Tags:        []string{"Projects"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, h.updateProject)

	huma.Register(api, huma.Operation{
		OperationID: "delete-project",
		Method:      http.MethodDelete,
		Path:        "/projects/{id}",
		Summary:     "Delete Project",
		Description: "Soft delete a project. Only the owner can delete.",
		Tags:        []string{"Projects"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, h.deleteProject)

	huma.Register(api, huma.Operation{
		OperationID: "transfer-project-ownership",
		Method:      http.MethodPost,
		Path:        "/projects/{id}/transfer",
		Summary:     "Transfer Project Ownership",
		Description: "Transfer ownership of a project to another user. Only the current owner can transfer.",
		Tags:        []string{"Projects"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, h.transferOwnership)
}

// ========== Request/Response Types ==========

// ProjectResponse represents a project in API responses.
type ProjectResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// toProjectResponse converts a domain project to an API response.
func toProjectResponse(p *domain.Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID.String(),
		Title:       p.Title,
		Description: p.Description,
		OwnerID:     p.OwnerID.String(),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ListMyProjectsOutput is the response for listing user's projects.
type ListMyProjectsOutput struct {
	Body ApiResponse[[]ProjectResponse]
}

// GetProjectInput is the request for getting a project.
type GetProjectInput struct {
	ID string `path:"id" doc:"Project ID (UUID)"`
}

// GetProjectOutput is the response for getting a project.
type GetProjectOutput struct {
	Body ApiResponse[ProjectResponse]
}

// CreateProjectInput is the request for creating a project.
type CreateProjectInput struct {
	Body struct {
		Title       string `json:"title" required:"true" minLength:"1" maxLength:"100" doc:"Project title"`
		Description string `json:"description" maxLength:"1000" doc:"Project description (optional)"`
	}
}

// CreateProjectOutput is the response for creating a project.
type CreateProjectOutput struct {
	Body ApiResponse[ProjectResponse]
}

// UpdateProjectInput is the request for updating a project.
type UpdateProjectInput struct {
	ID   string `path:"id" doc:"Project ID (UUID)"`
	Body struct {
		Title       *string `json:"title,omitempty" minLength:"1" maxLength:"100" doc:"New project title (optional)"`
		Description *string `json:"description,omitempty" maxLength:"1000" doc:"New project description (optional)"`
	}
}

// UpdateProjectOutput is the response for updating a project.
type UpdateProjectOutput struct {
	Body ApiResponse[ProjectResponse]
}

// DeleteProjectInput is the request for deleting a project.
type DeleteProjectInput struct {
	ID string `path:"id" doc:"Project ID (UUID)"`
}

// DeleteProjectOutput is the response for deleting a project.
type DeleteProjectOutput struct {
	Body ApiResponse[emptyData]
}

// TransferOwnershipInput is the request for transferring project ownership.
type TransferOwnershipInput struct {
	ID   string `path:"id" doc:"Project ID (UUID)"`
	Body struct {
		NewOwnerID string `json:"new_owner_id" required:"true" doc:"UUID of the new owner"`
	}
}

// TransferOwnershipOutput is the response for transferring project ownership.
type TransferOwnershipOutput struct {
	Body ApiResponse[ProjectResponse]
}

// ========== Handlers ==========

func (h *ProjectHandler) listMyProjects(ctx context.Context, input *struct{}) (*ListMyProjectsOutput, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	projects, err := h.projectService.ListByOwner(ctx, user.ID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	response := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		response[i] = toProjectResponse(p)
	}

	return &ListMyProjectsOutput{
		Body: ok("Projects retrieved successfully", response),
	}, nil
}

func (h *ProjectHandler) getProject(ctx context.Context, input *GetProjectInput) (*GetProjectOutput, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid project ID format")
	}

	project, err := h.projectService.GetByID(ctx, id)
	if err != nil {
		return nil, mapDomainError(err)
	}

	// Only owner can access their project (for now)
	// TODO: Add support for project members/collaborators
	if project.OwnerID != user.ID {
		return nil, huma.Error403Forbidden("not authorized to access this project")
	}

	return &GetProjectOutput{
		Body: ok("Project retrieved successfully", toProjectResponse(project)),
	}, nil
}

func (h *ProjectHandler) createProject(ctx context.Context, input *CreateProjectInput) (*CreateProjectOutput, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	project, err := h.projectService.Create(ctx, domain.CreateProjectInput{
		Title:       input.Body.Title,
		Description: input.Body.Description,
		OwnerID:     user.ID,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &CreateProjectOutput{
		Body: created("Project created successfully", toProjectResponse(project)),
	}, nil
}

func (h *ProjectHandler) updateProject(ctx context.Context, input *UpdateProjectInput) (*UpdateProjectOutput, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid project ID format")
	}

	project, err := h.projectService.Update(ctx, id, domain.UpdateProjectInput{
		Title:       input.Body.Title,
		Description: input.Body.Description,
	}, user.ID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &UpdateProjectOutput{
		Body: ok("Project updated successfully", toProjectResponse(project)),
	}, nil
}

func (h *ProjectHandler) deleteProject(ctx context.Context, input *DeleteProjectInput) (*DeleteProjectOutput, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid project ID format")
	}

	if err := h.projectService.Delete(ctx, id, user.ID); err != nil {
		return nil, mapDomainError(err)
	}

	return &DeleteProjectOutput{
		Body: ok("Project deleted successfully", emptyData{}),
	}, nil
}

func (h *ProjectHandler) transferOwnership(ctx context.Context, input *TransferOwnershipInput) (*TransferOwnershipOutput, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	projectID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid project ID format")
	}

	newOwnerID, err := uuid.Parse(input.Body.NewOwnerID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid new owner ID format")
	}

	project, err := h.projectService.TransferOwnership(ctx, projectID, newOwnerID, user.ID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &TransferOwnershipOutput{
		Body: ok("Project ownership transferred successfully", toProjectResponse(project)),
	}, nil
}
