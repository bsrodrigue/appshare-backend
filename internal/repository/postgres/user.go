// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserRepository implements repository.UserRepository using PostgreSQL.
type UserRepository struct {
	q *db.Queries
}

// NewUserRepository creates a new PostgreSQL user repository.
func NewUserRepository(q *db.Queries) *UserRepository {
	return &UserRepository{q: q}
}

// ============================================================================
// Standard Methods (use internal queries)
// ============================================================================

// Create creates a new user in the database.
func (r *UserRepository) Create(ctx context.Context, input domain.CreateUserInput, passwordHash string) (*domain.User, error) {
	return r.CreateTx(ctx, r.q, input, passwordHash)
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.GetByIDTx(ctx, r.q, id)
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, translateError(err)
	}
	return getUserByEmailRowToUser(&row), nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, translateError(err)
	}
	return getUserByUsernameRowToUser(&row), nil
}

// GetCredentialsByEmail retrieves user credentials for authentication.
func (r *UserRepository) GetCredentialsByEmail(ctx context.Context, email string) (*domain.UserCredentials, error) {
	row, err := r.q.GetUserCredentialsByEmail(ctx, email)
	if err != nil {
		return nil, translateError(err)
	}
	return &domain.UserCredentials{
		ID:           pgtypeToUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		IsActive:     row.IsActive,
	}, nil
}

// GetCredentialsByUsername retrieves user credentials for authentication.
func (r *UserRepository) GetCredentialsByUsername(ctx context.Context, username string) (*domain.UserCredentials, error) {
	row, err := r.q.GetUserCredentialsByUsername(ctx, username)
	if err != nil {
		return nil, translateError(err)
	}
	return &domain.UserCredentials{
		ID:           pgtypeToUUID(row.ID),
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		IsActive:     row.IsActive,
	}, nil
}

// List retrieves all active users.
func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.q.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(rows))
	for i, row := range rows {
		users[i] = listUserRowToUser(&row)
	}
	return users, nil
}

// UpdateEmail updates a user's email.
func (r *UserRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (*domain.User, error) {
	row, err := r.q.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
		ID:    uuidToPgtype(id),
		Email: email,
	})
	if err != nil {
		return nil, translateError(err)
	}
	return updateUserEmailRowToUser(&row), nil
}

// UpdateUsername updates a user's username.
func (r *UserRepository) UpdateUsername(ctx context.Context, id uuid.UUID, username string) (*domain.User, error) {
	row, err := r.q.UpdateUserUsername(ctx, db.UpdateUserUsernameParams{
		ID:       uuidToPgtype(id),
		Username: username,
	})
	if err != nil {
		return nil, translateError(err)
	}
	return updateUserUsernameRowToUser(&row), nil
}

// UpdatePassword updates a user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := r.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           uuidToPgtype(id),
		PasswordHash: passwordHash,
	})
	return translateError(err)
}

// UpdateProfile updates a user's first and last name.
func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) (*domain.User, error) {
	row, err := r.q.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:        uuidToPgtype(id),
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return nil, translateError(err)
	}
	return updateUserProfileRowToUser(&row), nil
}

// UpdateLastLogin updates the last login timestamp.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.q.UpdateLastLogin(ctx, uuidToPgtype(id))
	return translateError(err)
}

// SoftDelete marks a user as deleted.
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.SoftDeleteTx(ctx, r.q, id)
}

// EmailExists checks if an email is already registered.
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.EmailExistsTx(ctx, r.q, email)
}

// UsernameExists checks if a username is already taken.
func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	return r.UsernameExistsTx(ctx, r.q, username)
}

// ============================================================================
// Transaction Methods (use provided queries)
// ============================================================================

// CreateTx creates a user within a transaction.
func (r *UserRepository) CreateTx(ctx context.Context, q *db.Queries, input domain.CreateUserInput, passwordHash string) (*domain.User, error) {
	row, err := q.CreateUser(ctx, db.CreateUserParams{
		Email:        input.Email,
		Username:     input.Username,
		PhoneNumber:  input.PhoneNumber,
		PasswordHash: passwordHash,
		IsActive:     true,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
	})
	if err != nil {
		return nil, translateError(err)
	}
	return rowToUser(&row), nil
}

// GetByIDTx retrieves a user by ID within a transaction.
func (r *UserRepository) GetByIDTx(ctx context.Context, q *db.Queries, id uuid.UUID) (*domain.User, error) {
	row, err := q.GetUserByID(ctx, uuidToPgtype(id))
	if err != nil {
		return nil, translateError(err)
	}
	return getUserByIDRowToUser(&row), nil
}

// EmailExistsTx checks email existence within a transaction.
func (r *UserRepository) EmailExistsTx(ctx context.Context, q *db.Queries, email string) (bool, error) {
	_, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UsernameExistsTx checks username existence within a transaction.
func (r *UserRepository) UsernameExistsTx(ctx context.Context, q *db.Queries, username string) (bool, error) {
	_, err := q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SoftDeleteTx marks a user as deleted within a transaction.
func (r *UserRepository) SoftDeleteTx(ctx context.Context, q *db.Queries, id uuid.UUID) error {
	_, err := q.SoftDeleteUser(ctx, uuidToPgtype(id))
	return translateError(err)
}

// ============================================================================
// Helper Functions
// ============================================================================

// translateError converts database errors to domain errors.
func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

// uuidToPgtype converts a google/uuid to pgtype.UUID.
func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// pgtypeToUUID converts a pgtype.UUID to google/uuid.
func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return id.Bytes
}

// pgtypeToTime converts pgtype.Timestamp to *time.Time.
func pgtypeToTime(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

// ============================================================================
// Row Conversion Functions
// ============================================================================

func rowToUser(row *db.CreateUserRow) *domain.User {
	return &domain.User{
		ID:          pgtypeToUUID(row.ID),
		Email:       row.Email,
		Username:    row.Username,
		PhoneNumber: row.PhoneNumber,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		LastLoginAt: pgtypeToTime(row.LastLoginAt),
	}
}

func getUserByIDRowToUser(row *db.GetUserByIDRow) *domain.User {
	return &domain.User{
		ID:          pgtypeToUUID(row.ID),
		Email:       row.Email,
		Username:    row.Username,
		PhoneNumber: row.PhoneNumber,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		LastLoginAt: pgtypeToTime(row.LastLoginAt),
	}
}

func getUserByEmailRowToUser(row *db.GetUserByEmailRow) *domain.User {
	return &domain.User{
		ID:          pgtypeToUUID(row.ID),
		Email:       row.Email,
		Username:    row.Username,
		PhoneNumber: row.PhoneNumber,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		LastLoginAt: pgtypeToTime(row.LastLoginAt),
	}
}

func getUserByUsernameRowToUser(row *db.GetUserByUsernameRow) *domain.User {
	return &domain.User{
		ID:          pgtypeToUUID(row.ID),
		Email:       row.Email,
		Username:    row.Username,
		PhoneNumber: row.PhoneNumber,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		LastLoginAt: pgtypeToTime(row.LastLoginAt),
	}
}

func listUserRowToUser(row *db.ListUsersRow) *domain.User {
	return &domain.User{
		ID:          pgtypeToUUID(row.ID),
		Email:       row.Email,
		Username:    row.Username,
		PhoneNumber: row.PhoneNumber,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		LastLoginAt: pgtypeToTime(row.LastLoginAt),
	}
}

func updateUserEmailRowToUser(row *db.UpdateUserEmailRow) *domain.User {
	return &domain.User{
		ID:          pgtypeToUUID(row.ID),
		Email:       row.Email,
		Username:    row.Username,
		PhoneNumber: row.PhoneNumber,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		LastLoginAt: pgtypeToTime(row.LastLoginAt),
	}
}

func updateUserUsernameRowToUser(row *db.UpdateUserUsernameRow) *domain.User {
	return &domain.User{
		ID:          pgtypeToUUID(row.ID),
		Email:       row.Email,
		Username:    row.Username,
		PhoneNumber: row.PhoneNumber,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		LastLoginAt: pgtypeToTime(row.LastLoginAt),
	}
}

func updateUserProfileRowToUser(row *db.UpdateUserProfileRow) *domain.User {
	return &domain.User{
		ID:          pgtypeToUUID(row.ID),
		Email:       row.Email,
		Username:    row.Username,
		PhoneNumber: row.PhoneNumber,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		LastLoginAt: pgtypeToTime(row.LastLoginAt),
	}
}
