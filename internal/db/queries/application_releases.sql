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

-- name: ListReleasesByEnvironment :many
SELECT * FROM application_releases 
WHERE application_id = $1 AND environment = $2 AND deleted_at IS NULL
ORDER BY version_code DESC;

-- name: GetLatestReleaseByEnvironment :one
SELECT * FROM application_releases 
WHERE application_id = $1 AND environment = $2 AND deleted_at IS NULL
ORDER BY version_code DESC
LIMIT 1;

-- ============================================================================
-- Granular Update Queries
-- ============================================================================

-- name: UpdateReleaseTitle :one
UPDATE application_releases SET
    title = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateReleaseNote :one
UPDATE application_releases SET
    release_note = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateRelease :one
-- Full update for title + release_note
UPDATE application_releases SET
    title = $2,
    release_note = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: PromoteRelease :one
-- Change environment (e.g., development -> staging -> production)
UPDATE application_releases SET
    environment = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- ============================================================================
-- Delete Queries  
-- ============================================================================

-- name: SoftDeleteApplicationRelease :one
UPDATE application_releases SET
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteApplicationRelease :exec
DELETE FROM application_releases WHERE id = $1;
