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

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register registers all auth routes with the API.
func (h *AuthHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login",
		Description: "Authenticate with email/username and password.",
		Tags:        []string{"Auth"},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Register",
		Description: "Create a new user account.",
		Tags:        []string{"Auth"},
	}, h.register)
}

// ========== Request/Response Types ==========

// LoginInput is the request for login.
type LoginInput struct {
	Body struct {
		Email    string `json:"email" required:"true" doc:"Email or username"`
		Password string `json:"password" required:"true" doc:"Password"`
	}
}

// LoginResponse represents the login response data.
type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token,omitempty"` // TODO: Implement JWT
}

// LoginOutput is the response for login.
type LoginOutput struct {
	Body ApiResponse[LoginResponse]
}

// RegisterInput is the request for registration.
type RegisterInput struct {
	Body struct {
		Email       string `json:"email" required:"true" doc:"Email address"`
		Username    string `json:"username" required:"true" doc:"Username (3-30 characters)"`
		PhoneNumber string `json:"phone_number" required:"true" doc:"Phone number with country code"`
		Password    string `json:"password" required:"true" minLength:"8" doc:"Password (min 8 characters)"`
		FirstName   string `json:"first_name" required:"true" doc:"First name"`
		LastName    string `json:"last_name" required:"true" doc:"Last name"`
	}
}

// RegisterResponse represents the registration response.
type RegisterResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	PhoneNumber string    `json:"phone_number"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// RegisterOutput is the response for registration.
type RegisterOutput struct {
	Body ApiResponse[RegisterResponse]
}

// ========== Handlers ==========

func (h *AuthHandler) login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	result, err := h.authService.Login(ctx, service.LoginInput{
		Email:    input.Body.Email,
		Password: input.Body.Password,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &LoginOutput{
		Body: ok("Login successful", LoginResponse{
			User: toUserResponse(result.User),
			// Token: result.Token, // TODO
		}),
	}, nil
}

func (h *AuthHandler) register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
	user, err := h.authService.Register(ctx, domain.CreateUserInput{
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

	return &RegisterOutput{
		Body: created("Registration successful", RegisterResponse{
			ID:          user.ID.String(),
			Email:       user.Email,
			Username:    user.Username,
			PhoneNumber: user.PhoneNumber,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			CreatedAt:   user.CreatedAt,
		}),
	}, nil
}

// parseUUID is a helper to parse and validate UUIDs.
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
