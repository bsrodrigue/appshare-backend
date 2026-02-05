-- name: CreateProjectInvite :one
INSERT INTO project_invites (
    project_id,
    invited_user_id,
    status
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetProjectInviteByID :one
SELECT * FROM project_invites 
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListInvitesByProject :many
SELECT * FROM project_invites 
WHERE project_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListInvitesByUser :many
SELECT * FROM project_invites 
WHERE invited_user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateInviteStatus :one
UPDATE project_invites SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteInvite :one
UPDATE project_invites SET
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteInvite :exec
DELETE FROM project_invites WHERE id = $1;
