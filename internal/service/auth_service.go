package service

import (
	"context"
	"errors"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo repository.UserRepository
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// LoginInput represents credentials for login.
type LoginInput struct {
	Email    string // Can be email or username
	Password string
}

// LoginResult represents a successful login response.
type LoginResult struct {
	User *domain.User
	// Token string // TODO: Add JWT token when implementing auth
}

// Login authenticates a user by email and password.
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
		return nil, err
	}

	// Check if user is active
	if !creds.IsActive {
		return nil, domain.ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, creds.ID); err != nil {
		// Log error but don't fail login
		// TODO: Add logging
	}

	// Get full user data (without password hash)
	user, err := s.userRepo.GetByID(ctx, creds.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User: user,
		// Token: s.generateToken(user), // TODO
	}, nil
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	// Validation is done in UserService.Create via repository checks
	// Additional auth-specific validation could go here

	// Check email uniqueness
	exists, err := s.userRepo.EmailExists(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Check username uniqueness
	exists, err = s.userRepo.UsernameExists(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrUsernameAlreadyExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	return s.userRepo.Create(ctx, input, string(hash))
}
