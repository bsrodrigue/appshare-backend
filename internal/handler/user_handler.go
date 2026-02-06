package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/service"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register registers all user routes with the API.
func (h *UserHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "List Users",
		Description: "Retrieve a list of all active users.",
		Tags:        []string{"Users"},
	}, h.listUsers)

	huma.Register(api, huma.Operation{
		OperationID: "get-user",
		Method:      http.MethodGet,
		Path:        "/users/{id}",
		Summary:     "Get User",
		Description: "Retrieve a specific user by ID.",
		Tags:        []string{"Users"},
	}, h.getUser)

	huma.Register(api, huma.Operation{
		OperationID: "create-user",
		Method:      http.MethodPost,
		Path:        "/users",
		Summary:     "Create User",
		Description: "Create a new user account.",
		Tags:        []string{"Users"},
	}, h.createUser)

	huma.Register(api, huma.Operation{
		OperationID: "update-user-profile",
		Method:      http.MethodPatch,
		Path:        "/users/{id}/profile",
		Summary:     "Update User Profile",
		Description: "Update a user's profile (first name, last name).",
		Tags:        []string{"Users"},
	}, h.updateProfile)

	huma.Register(api, huma.Operation{
		OperationID: "delete-user",
		Method:      http.MethodDelete,
		Path:        "/users/{id}",
		Summary:     "Delete User",
		Description: "Soft delete a user account.",
		Tags:        []string{"Users"},
	}, h.deleteUser)
}

// ========== Request/Response Types ==========

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	PhoneNumber string     `json:"phone_number"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// toUserResponse converts a domain user to an API response.
func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		Username:    u.Username,
		PhoneNumber: u.PhoneNumber,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}

// ListUsersOutput is the response for listing users.
type ListUsersOutput struct {
	Body ApiResponse[[]UserResponse]
}

// GetUserInput is the request for getting a user.
type GetUserInput struct {
	ID string `path:"id" doc:"User ID (UUID)"`
}

// GetUserOutput is the response for getting a user.
type GetUserOutput struct {
	Body ApiResponse[UserResponse]
}

// CreateUserInput is the request for creating a user.
type CreateUserInput struct {
	Body struct {
		Email       string `json:"email" required:"true" doc:"User email address"`
		Username    string `json:"username" required:"true" doc:"Unique username"`
		PhoneNumber string `json:"phone_number" required:"true" doc:"Phone number with country code"`
		Password    string `json:"password" required:"true" minLength:"8" doc:"Password (min 8 characters)"`
		FirstName   string `json:"first_name" required:"true" doc:"First name"`
		LastName    string `json:"last_name" required:"true" doc:"Last name"`
	}
}

// CreateUserOutput is the response for creating a user.
type CreateUserOutput struct {
	Body ApiResponse[UserResponse]
}

// UpdateProfileInput is the request for updating user profile.
type UpdateProfileInput struct {
	ID   string `path:"id" doc:"User ID (UUID)"`
	Body struct {
		FirstName string `json:"first_name" required:"true" doc:"New first name"`
		LastName  string `json:"last_name" required:"true" doc:"New last name"`
	}
}

// UpdateProfileOutput is the response for updating profile.
type UpdateProfileOutput struct {
	Body ApiResponse[UserResponse]
}

// DeleteUserInput is the request for deleting a user.
type DeleteUserInput struct {
	ID string `path:"id" doc:"User ID (UUID)"`
}

// DeleteUserOutput is the response for deleting a user.
type DeleteUserOutput struct {
	Body ApiResponse[emptyData]
}

// ========== Handlers ==========

func (h *UserHandler) listUsers(ctx context.Context, input *struct{}) (*ListUsersOutput, error) {
	users, err := h.userService.List(ctx)
	if err != nil {
		return nil, mapDomainError(err)
	}

	response := make([]UserResponse, len(users))
	for i, u := range users {
		response[i] = toUserResponse(u)
	}

	return &ListUsersOutput{
		Body: ok("Users retrieved successfully", response),
	}, nil
}

func (h *UserHandler) getUser(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid user ID format")
	}

	user, err := h.userService.GetByID(ctx, id)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &GetUserOutput{
		Body: ok("User retrieved successfully", toUserResponse(user)),
	}, nil
}

func (h *UserHandler) createUser(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
	user, err := h.userService.Create(ctx, domain.CreateUserInput{
		Email:       input.Body.Email,
		Username:    input.Body.Username,
		PhoneNumber: input.Body.PhoneNumber,
		Password:    input.Body.Password,
		FirstName:   input.Body.FirstName,
		LastName:    input.Body.LastName,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &CreateUserOutput{
		Body: created("User created successfully", toUserResponse(user)),
	}, nil
}

func (h *UserHandler) updateProfile(ctx context.Context, input *UpdateProfileInput) (*UpdateProfileOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid user ID format")
	}

	user, err := h.userService.UpdateProfile(ctx, id, input.Body.FirstName, input.Body.LastName)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &UpdateProfileOutput{
		Body: ok("Profile updated successfully", toUserResponse(user)),
	}, nil
}

func (h *UserHandler) deleteUser(ctx context.Context, input *DeleteUserInput) (*DeleteUserOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid user ID format")
	}

	if err := h.userService.Delete(ctx, id); err != nil {
		return nil, mapDomainError(err)
	}

	return &DeleteUserOutput{
		Body: ok("User deleted successfully", emptyData{}),
	}, nil
}
