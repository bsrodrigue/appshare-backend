-- name: CreateArtifact :one
INSERT INTO artifacts (
    file_url,
    sha256_hash,
    file_size,
    file_type,
    abi,
    release_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetArtifactByID :one
SELECT * FROM artifacts 
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListArtifactsByRelease :many
SELECT * FROM artifacts 
WHERE release_id = $1 AND deleted_at IS NULL;

-- name: GetArtifactByReleaseAndABI :one
SELECT * FROM artifacts 
WHERE release_id = $1 AND abi = $2 AND deleted_at IS NULL;

-- name: SoftDeleteArtifact :one
UPDATE artifacts SET
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteArtifact :exec
DELETE FROM artifacts WHERE id = $1;
