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
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT * FROM users 
WHERE username = $1 AND deleted_at IS NULL;

-- name: GetUserByPhoneNumber :one
SELECT * FROM users 
WHERE phone_number = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT * FROM users 
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT * FROM users 
WHERE deleted_at IS NULL 
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users SET
    email = $2,
    username = $3,
    phone_number = $4,
    password_hash = $5,
    is_active = $6,
    first_name = $7,
    last_name = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateLastLogin :one
UPDATE users SET
    last_login_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteUser :one
UPDATE users SET
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteUser :one
DELETE FROM users WHERE id = $1 RETURNING *;