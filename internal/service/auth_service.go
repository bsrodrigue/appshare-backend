package service

import (
	"context"
	"errors"

	"github.com/bsrodrigue/appshare-backend/internal/auth"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo   repository.UserRepository
	jwtService *auth.JWTService
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, jwtService *auth.JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// LoginInput represents credentials for login.
type LoginInput struct {
	Email    string // Can be email or username
	Password string
}

// LoginResult represents a successful login response.
type LoginResult struct {
	User   *domain.User
	Tokens *auth.TokenPair
}

// Login authenticates a user by email/username and password.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	// Try to get credentials by email first, then by username
	creds, err := s.userRepo.GetCredentialsByEmail(ctx, input.Email)
	if errors.Is(err, domain.ErrNotFound) {
		// Try username
		creds, err = s.userRepo.GetCredentialsByUsername(ctx, input.Email)
	}
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, domain.WrapError(domain.CodeInternal, "failed to retrieve credentials", err)
	}

	// Check if user is active
	if !creds.IsActive {
		return nil, domain.ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Get full user data (without password hash)
	user, err := s.userRepo.GetByID(ctx, creds.ID)
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to retrieve user", err)
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to generate tokens", err)
	}

	// Update last login (fire and forget)
	_ = s.userRepo.UpdateLastLogin(ctx, creds.ID)

	return &LoginResult{
		User:   user,
		Tokens: tokens,
	}, nil
}

// RegisterInput extends domain.CreateUserInput with any registration-specific fields.
type RegisterInput struct {
	domain.CreateUserInput
}

// RegisterResult represents a successful registration response.
type RegisterResult struct {
	User   *domain.User
	Tokens *auth.TokenPair
}

// Register creates a new user account and returns tokens.
func (s *AuthService) Register(ctx context.Context, input domain.CreateUserInput) (*RegisterResult, error) {
	// Validate input
	if input.Email == "" {
		return nil, domain.NewValidationError("email", "email is required")
	}
	if input.Username == "" {
		return nil, domain.NewValidationError("username", "username is required")
	}
	if len(input.Password) < 8 {
		return nil, domain.NewValidationError("password", "password must be at least 8 characters")
	}

	// Check email uniqueness
	exists, err := s.userRepo.EmailExists(ctx, input.Email)
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to check email", err)
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Check username uniqueness
	exists, err = s.userRepo.UsernameExists(ctx, input.Username)
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to check username", err)
	}
	if exists {
		return nil, domain.ErrUsernameAlreadyExists
	}

	// Check phone number uniqueness
	exists, err = s.userRepo.PhoneNumberExists(ctx, input.PhoneNumber)
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to check phone number", err)
	}
	if exists {
		return nil, domain.ErrPhoneAlreadyExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to hash password", err)
	}

	// Create user
	user, err := s.userRepo.Create(ctx, input, string(hash))
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to create user", err)
	}

	// Generate tokens (auto-login after registration)
	tokens, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to generate tokens", err)
	}

	return &RegisterResult{
		User:   user,
		Tokens: tokens,
	}, nil
}

// RefreshTokenInput represents the refresh token request.
type RefreshTokenInput struct {
	RefreshToken string
}

// RefreshResult represents the result of a token refresh.
type RefreshResult struct {
	Tokens *auth.TokenPair
}

// RefreshTokens generates new access and refresh tokens from a valid refresh token.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	// Validate the refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get the user (to ensure they still exist and are active)
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrTokenInvalid
		}
		return nil, domain.WrapError(domain.CodeInternal, "failed to retrieve user", err)
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	// Generate new token pair
	tokens, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, domain.WrapError(domain.CodeInternal, "failed to generate tokens", err)
	}

	return &RefreshResult{
		Tokens: tokens,
	}, nil
}

// GetCurrentUser retrieves the current authenticated user.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

// ChangePassword changes the user's password.
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get current credentials
	creds, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// We need to get credentials with password hash
	credsWithHash, err := s.userRepo.GetCredentialsByEmail(ctx, creds.Email)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(credsWithHash.PasswordHash), []byte(currentPassword)); err != nil {
		return domain.NewAppError(domain.CodeInvalidCredentials, "current password is incorrect")
	}

	// Validate new password
	if len(newPassword) < 8 {
		return domain.NewValidationError("new_password", "password must be at least 8 characters")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return domain.WrapError(domain.CodeInternal, "failed to hash password", err)
	}

	// Update password
	return s.userRepo.UpdatePassword(ctx, userID, string(hash))
}
