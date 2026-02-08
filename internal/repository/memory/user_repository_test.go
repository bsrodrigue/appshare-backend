package memory

import (
	"context"
	"testing"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	input := domain.CreateUserInput{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}
	passwordHash := "hashed_password"

	user, err := repo.Create(ctx, input, passwordHash)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, input.Email, user.Email)
	assert.Equal(t, input.Username, user.Username)

	// Test unique email
	_, err = repo.Create(ctx, input, passwordHash)
	assert.Error(t, err)
	assert.Equal(t, domain.CodeEmailExists, domain.GetErrorCode(err))
}

func TestUserRepository_GetByID(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	input := domain.CreateUserInput{
		Email:    "test@example.com",
		Username: "testuser",
	}
	created, _ := repo.Create(ctx, input, "hash")

	user, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, user.ID)

	_, err = repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, domain.CodeNotFound, domain.GetErrorCode(err))
}

func TestUserRepository_UpdateEmail(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	u1, _ := repo.Create(ctx, domain.CreateUserInput{Email: "u1@ex.com", Username: "u1"}, "hash")
	u2, _ := repo.Create(ctx, domain.CreateUserInput{Email: "u2@ex.com", Username: "u2"}, "hash")
	require.NotNil(t, u2)

	// Successful update
	updated, err := repo.UpdateEmail(ctx, u1.ID, "new@ex.com")
	require.NoError(t, err)
	assert.Equal(t, "new@ex.com", updated.Email)

	// Duplicate email update
	_, err = repo.UpdateEmail(ctx, u1.ID, "u2@ex.com")
	assert.Error(t, err)
	assert.Equal(t, domain.CodeEmailExists, domain.GetErrorCode(err))
}

func TestUserRepository_SoftDelete(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	u, _ := repo.Create(ctx, domain.CreateUserInput{Email: "u@ex.com", Username: "u"}, "hash")

	err := repo.SoftDelete(ctx, u.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, u.ID)
	assert.Error(t, err)
	assert.Equal(t, domain.CodeNotFound, domain.GetErrorCode(err))
}
