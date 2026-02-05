-- name: CreateApplicationRelease :one
INSERT INTO application_releases (
    title,
    version_code,
    version_name,
    release_note,
    environment,
    application_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetApplicationReleaseByID :one
SELECT * FROM application_releases 
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListReleasesByApplication :many
SELECT * FROM application_releases 
WHERE application_id = $1 AND deleted_at IS NULL
ORDER BY version_code DESC;

-- name: GetLatestReleaseByEnvironment :one
SELECT * FROM application_releases 
WHERE application_id = $1 AND environment = $2 AND deleted_at IS NULL
ORDER BY version_code DESC
LIMIT 1;

-- name: UpdateApplicationRelease :one
UPDATE application_releases SET
    title = $2,
    release_note = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteApplicationRelease :one
UPDATE application_releases SET
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteApplicationRelease :exec
DELETE FROM application_releases WHERE id = $1;
