package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
)

// UserRecord internal representation of a user with sensitive data.
type userRecord struct {
	user         domain.User
	passwordHash string
}

// UserRepository implements repository.UserRepository in memory.
type UserRepository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*userRecord
}

// NewUserRepository creates a new in-memory user repository.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[uuid.UUID]*userRecord),
	}
}

// Reset clears all data in the repository.
func (r *UserRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users = make(map[uuid.UUID]*userRecord)
}

// ============================================================================
// Standard Methods
// ============================================================================

func (r *UserRepository) Create(ctx context.Context, input domain.CreateUserInput, passwordHash string) (*domain.User, error) {
	return r.CreateTx(ctx, nil, input, passwordHash)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.GetByIDTx(ctx, nil, id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rec := range r.users {
		if rec.user.Email == email && rec.user.IsActive {
			user := rec.user
			return &user, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rec := range r.users {
		if rec.user.Username == username && rec.user.IsActive {
			user := rec.user
			return &user, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *UserRepository) GetCredentialsByEmail(ctx context.Context, email string) (*domain.UserCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rec := range r.users {
		if rec.user.Email == email {
			return &domain.UserCredentials{
				ID:           rec.user.ID,
				Email:        rec.user.Email,
				Username:     rec.user.Username,
				PasswordHash: rec.passwordHash,
				IsActive:     rec.user.IsActive,
			}, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *UserRepository) GetCredentialsByUsername(ctx context.Context, username string) (*domain.UserCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rec := range r.users {
		if rec.user.Username == username {
			return &domain.UserCredentials{
				ID:           rec.user.ID,
				Email:        rec.user.Email,
				Username:     rec.user.Username,
				PasswordHash: rec.passwordHash,
				IsActive:     rec.user.IsActive,
			}, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*domain.User, 0, len(r.users))
	for _, rec := range r.users {
		if rec.user.IsActive {
			user := rec.user
			users = append(users, &user)
		}
	}

	// Sort by ID to ensure deterministic output
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID.String() < users[j].ID.String()
	})

	return users, nil
}

func (r *UserRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.users[id]
	if !ok || !rec.user.IsActive {
		return nil, domain.ErrNotFound
	}

	// Check if email already exists
	for _, other := range r.users {
		if other.user.ID != id && other.user.Email == email {
			return nil, domain.ErrEmailAlreadyExists
		}
	}

	rec.user.Email = email
	rec.user.UpdatedAt = time.Now()
	user := rec.user
	return &user, nil
}

func (r *UserRepository) UpdateUsername(ctx context.Context, id uuid.UUID, username string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.users[id]
	if !ok || !rec.user.IsActive {
		return nil, domain.ErrNotFound
	}

	// Check if username already exists
	for _, other := range r.users {
		if other.user.ID != id && other.user.Username == username {
			return nil, domain.ErrUsernameAlreadyExists
		}
	}

	rec.user.Username = username
	rec.user.UpdatedAt = time.Now()
	user := rec.user
	return &user, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.users[id]
	if !ok || !rec.user.IsActive {
		return domain.ErrNotFound
	}

	rec.passwordHash = passwordHash
	rec.user.UpdatedAt = time.Now()
	return nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.users[id]
	if !ok || !rec.user.IsActive {
		return nil, domain.ErrNotFound
	}

	rec.user.FirstName = firstName
	rec.user.LastName = lastName
	rec.user.UpdatedAt = time.Now()
	user := rec.user
	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.users[id]
	if !ok || !rec.user.IsActive {
		return domain.ErrNotFound
	}

	now := time.Now()
	rec.user.LastLoginAt = &now
	return nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.SoftDeleteTx(ctx, nil, id)
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.EmailExistsTx(ctx, nil, email)
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	return r.UsernameExistsTx(ctx, nil, username)
}

// ============================================================================
// Transaction Methods
// ============================================================================

func (r *UserRepository) CreateTx(ctx context.Context, q *db.Queries, input domain.CreateUserInput, passwordHash string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check unique constraints
	for _, rec := range r.users {
		if rec.user.Email == input.Email {
			return nil, domain.ErrEmailAlreadyExists
		}
		if rec.user.Username == input.Username {
			return nil, domain.ErrUsernameAlreadyExists
		}
	}

	id := uuid.New()
	now := time.Now()
	user := domain.User{
		ID:          id,
		Email:       input.Email,
		Username:    input.Username,
		PhoneNumber: input.PhoneNumber,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.users[id] = &userRecord{
		user:         user,
		passwordHash: passwordHash,
	}

	return &user, nil
}

func (r *UserRepository) GetByIDTx(ctx context.Context, q *db.Queries, id uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rec, ok := r.users[id]
	if !ok || !rec.user.IsActive {
		return nil, domain.ErrNotFound
	}

	user := rec.user
	return &user, nil
}

func (r *UserRepository) EmailExistsTx(ctx context.Context, q *db.Queries, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rec := range r.users {
		if rec.user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (r *UserRepository) UsernameExistsTx(ctx context.Context, q *db.Queries, username string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rec := range r.users {
		if rec.user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (r *UserRepository) SoftDeleteTx(ctx context.Context, q *db.Queries, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.users[id]
	if !ok || !rec.user.IsActive {
		return domain.ErrNotFound
	}

	rec.user.IsActive = false
	rec.user.UpdatedAt = time.Now()
	return nil
}
