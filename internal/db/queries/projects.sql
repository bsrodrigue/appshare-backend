-- name: CreateProject :one
INSERT INTO projects (
    title,
    description,
    owner_id
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects 
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProjectsByOwner :many
SELECT * FROM projects 
WHERE owner_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- ============================================================================
-- Granular Update Queries
-- ============================================================================

-- name: UpdateProjectTitle :one
UPDATE projects SET
    title = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateProjectDescription :one
UPDATE projects SET
    description = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateProject :one
-- Full update when you need to change both title and description
UPDATE projects SET
    title = $2,
    description = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: TransferProjectOwnership :one
UPDATE projects SET
    owner_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- ============================================================================
-- Delete Queries
-- ============================================================================

-- name: SoftDeleteProject :one
UPDATE projects SET
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteProject :exec
DELETE FROM projects WHERE id = $1;
