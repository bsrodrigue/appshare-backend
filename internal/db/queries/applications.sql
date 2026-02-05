-- name: CreateApplication :one
INSERT INTO applications (
    title,
    package_name,
    description,
    project_id
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetApplicationByID :one
SELECT * FROM applications 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetApplicationByPackageName :one
SELECT * FROM applications 
WHERE package_name = $1 AND deleted_at IS NULL;

-- name: ListApplicationsByProject :many
SELECT * FROM applications 
WHERE project_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateApplication :one
UPDATE applications SET
    title = $2,
    description = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteApplication :one
UPDATE applications SET
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteApplication :exec
DELETE FROM applications WHERE id = $1;
