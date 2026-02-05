-- +goose Up
CREATE TYPE project_invite_status AS ENUM (
  'PENDING',
  'ACCEPTED',
  'REJECTED',
  'CANCELLED'
);

CREATE TABLE project_invites(
    -- Identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Information
    status project_invite_status NOT NULL DEFAULT 'PENDING',

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Relations
    project_id UUID NOT NULL,
    invited_user_id UUID NOT NULL,

    -- Foreign Keys
    FOREIGN KEY(project_id)
    REFERENCES projects(id)
    ON DELETE CASCADE,

    FOREIGN KEY(invited_user_id)
    REFERENCES users(id)
    ON DELETE CASCADE,

    -- Constraints
    CONSTRAINT unique_project_invite UNIQUE (project_id, invited_user_id)
);

-- +goose Down
DROP TABLE IF EXISTS project_invites;
DROP TYPE IF EXISTS project_invite_status;