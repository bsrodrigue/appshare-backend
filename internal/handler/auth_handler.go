package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/auth"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/service"
	"github.com/danielgtaylor/huma/v2"
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
	// Public routes (no auth required)
	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login",
		Description: "Authenticate with email/username and password. Returns access and refresh tokens.",
		Tags:        []string{"Auth"},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Register",
		Description: "Create a new user account. Returns access and refresh tokens.",
		Tags:        []string{"Auth"},
	}, h.register)

	huma.Register(api, huma.Operation{
		OperationID: "refresh-token",
		Method:      http.MethodPost,
		Path:        "/auth/refresh",
		Summary:     "Refresh Token",
		Description: "Exchange a valid refresh token for new access and refresh tokens.",
		Tags:        []string{"Auth"},
	}, h.refreshToken)

	// Protected routes (auth required) - registered separately with middleware
}

// RegisterProtected registers protected auth routes.
// Call this with routes that have auth middleware applied.
func (h *AuthHandler) RegisterProtected(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-current-user",
		Method:      http.MethodGet,
		Path:        "/auth/me",
		Summary:     "Get Current User",
		Description: "Get the currently authenticated user's profile.",
		Tags:        []string{"Auth"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.getCurrentUser)

	huma.Register(api, huma.Operation{
		OperationID: "change-password",
		Method:      http.MethodPost,
		Path:        "/auth/change-password",
		Summary:     "Change Password",
		Description: "Change the current user's password.",
		Tags:        []string{"Auth"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, h.changePassword)
}

// ========== Request/Response Types ==========

// TokenResponse represents JWT tokens in API responses.
type TokenResponse struct {
	AccessToken           string    `json:"access_token" doc:"JWT access token for API requests"`
	RefreshToken          string    `json:"refresh_token" doc:"JWT refresh token to get new access tokens"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at" doc:"Access token expiration time"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at" doc:"Refresh token expiration time"`
	TokenType             string    `json:"token_type" doc:"Token type (always 'Bearer')"`
}

// LoginInput is the request for login.
type LoginInput struct {
	Body struct {
		Email    string `json:"email" required:"true" doc:"Email or username"`
		Password string `json:"password" required:"true" doc:"Password"`
	}
}

// LoginResponse represents the login response data.
type LoginResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

// LoginOutput is the response for login.
type LoginOutput struct {
	Body ApiResponse[LoginResponse]
}

// RegisterInput is the request for registration.
type RegisterInput struct {
	Body struct {
		Email       string `json:"email" required:"true" doc:"Email address"`
		Username    string `json:"username" required:"true" minLength:"3" maxLength:"30" doc:"Username"`
		PhoneNumber string `json:"phone_number" required:"true" doc:"Phone number with country code"`
		Password    string `json:"password" required:"true" minLength:"8" doc:"Password (min 8 characters)"`
		FirstName   string `json:"first_name" required:"true" doc:"First name"`
		LastName    string `json:"last_name" required:"true" doc:"Last name"`
	}
}

// RegisterResponse represents the registration response.
type RegisterResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

// RegisterOutput is the response for registration.
type RegisterOutput struct {
	Body ApiResponse[RegisterResponse]
}

// RefreshTokenInput is the request for token refresh.
type RefreshTokenInput struct {
	Body struct {
		RefreshToken string `json:"refresh_token" required:"true" doc:"Valid refresh token"`
	}
}

// RefreshTokenResponse represents the token refresh response.
type RefreshTokenResponse struct {
	Tokens TokenResponse `json:"tokens"`
}

// RefreshTokenOutput is the response for token refresh.
type RefreshTokenOutput struct {
	Body ApiResponse[RefreshTokenResponse]
}

// GetCurrentUserOutput is the response for getting current user.
type GetCurrentUserOutput struct {
	Body ApiResponse[UserResponse]
}

// ChangePasswordInput is the request for changing password.
type ChangePasswordInput struct {
	Body struct {
		CurrentPassword string `json:"current_password" required:"true" doc:"Current password"`
		NewPassword     string `json:"new_password" required:"true" minLength:"8" doc:"New password (min 8 characters)"`
	}
}

// ChangePasswordOutput is the response for changing password.
type ChangePasswordOutput struct {
	Body ApiResponse[emptyData]
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
			User:   toUserResponse(result.User),
			Tokens: toTokenResponse(result.Tokens),
		}),
	}, nil
}

func (h *AuthHandler) register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
	result, err := h.authService.Register(ctx, domain.CreateUserInput{
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
			User:   toUserResponse(result.User),
			Tokens: toTokenResponse(result.Tokens),
		}),
	}, nil
}

func (h *AuthHandler) refreshToken(ctx context.Context, input *RefreshTokenInput) (*RefreshTokenOutput, error) {
	result, err := h.authService.RefreshTokens(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &RefreshTokenOutput{
		Body: ok("Token refreshed successfully", RefreshTokenResponse{
			Tokens: toTokenResponse(result.Tokens),
		}),
	}, nil
}

func (h *AuthHandler) getCurrentUser(ctx context.Context, input *struct{}) (*GetCurrentUserOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	user, err := h.authService.GetCurrentUser(ctx, authUser.ID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &GetCurrentUserOutput{
		Body: ok("User retrieved successfully", toUserResponse(user)),
	}, nil
}

func (h *AuthHandler) changePassword(ctx context.Context, input *ChangePasswordInput) (*ChangePasswordOutput, error) {
	authUser := auth.UserFromContext(ctx)
	if authUser == nil {
		return nil, mapDomainError(domain.ErrUnauthorized)
	}

	err := h.authService.ChangePassword(ctx, authUser.ID, input.Body.CurrentPassword, input.Body.NewPassword)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &ChangePasswordOutput{
		Body: ok("Password changed successfully", emptyData{}),
	}, nil
}

// ========== Helper Functions ==========

func toTokenResponse(tokens *auth.TokenPair) TokenResponse {
	return TokenResponse{
		AccessToken:           tokens.AccessToken,
		RefreshToken:          tokens.RefreshToken,
		AccessTokenExpiresAt:  tokens.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokens.RefreshTokenExpiresAt,
		TokenType:             tokens.TokenType,
	}
}
