package service

import (
	"context"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/bsrodrigue/appshare-backend/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles user-related business logic.
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create creates a new user with the given input.
// It validates uniqueness and hashes the password.
func (s *UserService) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	// Check email uniqueness
	exists, err := s.repo.EmailExists(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Check username uniqueness
	exists, err = s.repo.UsernameExists(ctx, input.Username)
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
	return s.repo.Create(ctx, input, string(hash))
}

// GetByID retrieves a user by their ID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByEmail retrieves a user by their email.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

// List retrieves all active users.
func (s *UserService) List(ctx context.Context) ([]*domain.User, error) {
	return s.repo.List(ctx)
}

// UpdateEmail updates a user's email after checking uniqueness.
func (s *UserService) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (*domain.User, error) {
	// Check if new email is already taken
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	return s.repo.UpdateEmail(ctx, id, email)
}

// UpdateUsername updates a user's username after checking uniqueness.
func (s *UserService) UpdateUsername(ctx context.Context, id uuid.UUID, username string) (*domain.User, error) {
	exists, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrUsernameAlreadyExists
	}

	return s.repo.UpdateUsername(ctx, id, username)
}

// UpdatePassword updates a user's password.
func (s *UserService) UpdatePassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, id, string(hash))
}

// UpdateProfile updates a user's profile information.
func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domain.User, error) {
	return s.repo.UpdateProfile(ctx, id, firstName, lastName)
}

// Delete soft-deletes a user.
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}
