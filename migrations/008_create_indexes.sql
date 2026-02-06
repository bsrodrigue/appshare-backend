-- +goose Up

-- ============================================================================
-- Performance Indexes
-- Using PARTIAL INDEXES with WHERE deleted_at IS NULL to:
--   1. Keep index size smaller (excludes soft-deleted rows)
--   2. Match our query patterns exactly (all queries filter deleted_at IS NULL)
--   3. Speed up lookups without indexing dead data
-- ============================================================================

-- Users: lookups by email, username, phone (for login/registration)
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_phone_number ON users(phone_number) WHERE deleted_at IS NULL;

-- Projects: list by owner
CREATE INDEX idx_projects_owner_id ON projects(owner_id) WHERE deleted_at IS NULL;

-- Applications: lookup by project, package name
CREATE INDEX idx_applications_project_id ON applications(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_package_name ON applications(package_name) WHERE deleted_at IS NULL;

-- Application Releases: list by application, filter by environment
CREATE INDEX idx_releases_application_id ON application_releases(application_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_releases_app_env ON application_releases(application_id, environment) WHERE deleted_at IS NULL;

-- Artifacts: list by release
CREATE INDEX idx_artifacts_release_id ON artifacts(release_id) WHERE deleted_at IS NULL;

-- Project Invites: list by project, list by user
CREATE INDEX idx_project_invites_project_id ON project_invites(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_project_invites_user_id ON project_invites(invited_user_id) WHERE deleted_at IS NULL;

-- Project Memberships: list by project, list by user
CREATE INDEX idx_memberships_project_id ON project_memberships(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_memberships_user_id ON project_memberships(user_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_memberships_user_id;
DROP INDEX IF EXISTS idx_memberships_project_id;
DROP INDEX IF EXISTS idx_project_invites_user_id;
DROP INDEX IF EXISTS idx_project_invites_project_id;
DROP INDEX IF EXISTS idx_artifacts_release_id;
DROP INDEX IF EXISTS idx_releases_app_env;
DROP INDEX IF EXISTS idx_releases_application_id;
DROP INDEX IF EXISTS idx_applications_package_name;
DROP INDEX IF EXISTS idx_applications_project_id;
DROP INDEX IF EXISTS idx_projects_owner_id;
DROP INDEX IF EXISTS idx_users_phone_number;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
