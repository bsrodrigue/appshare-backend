-- name: CreateUser :one
INSERT INTO users (
    email,
    username,
    phone_number,
    password_hash,
    is_active,
    first_name,
    last_name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at;

-- name: GetUserByEmail :one
SELECT id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at
FROM users 
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at
FROM users 
WHERE username = $1 AND deleted_at IS NULL;

-- name: GetUserByPhoneNumber :one
SELECT id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at
FROM users 
WHERE phone_number = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at
FROM users 
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at
FROM users 
WHERE deleted_at IS NULL 
ORDER BY created_at DESC;

-- ============================================================================
-- Authentication-specific queries (these DO return password_hash)
-- ============================================================================

-- name: GetUserCredentialsByEmail :one
-- Used for login - returns password hash for verification
SELECT id, email, password_hash, is_active
FROM users 
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserCredentialsByUsername :one
-- Used for login - returns password hash for verification
SELECT id, username, password_hash, is_active
FROM users 
WHERE username = $1 AND deleted_at IS NULL;

-- ============================================================================
-- Granular Update Queries
-- ============================================================================

-- name: UpdateUserEmail :one
UPDATE users SET
    email = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at;

-- name: UpdateUserUsername :one
UPDATE users SET
    username = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at;

-- name: UpdateUserPhoneNumber :one
UPDATE users SET
    phone_number = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at;

-- name: UpdateUserPassword :one
UPDATE users SET
    password_hash = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at;

-- name: UpdateUserProfile :one
UPDATE users SET
    first_name = $2,
    last_name = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at;

-- name: UpdateUserActiveStatus :one
UPDATE users SET
    is_active = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at;

-- name: UpdateLastLogin :one
UPDATE users SET
    last_login_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at;

-- ============================================================================
-- Delete Queries
-- ============================================================================

-- name: SoftDeleteUser :one
UPDATE users SET
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, phone_number, is_active, first_name, last_name, created_at, updated_at, last_login_at;

-- name: HardDeleteUser :exec
DELETE FROM users WHERE id = $1;