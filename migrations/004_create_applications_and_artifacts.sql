-- +goose Up
CREATE TYPE release_environment AS ENUM ('development', 'staging', 'production');

CREATE TABLE applications (
    -- Identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(256) NOT NULL,
    package_name VARCHAR(256) NOT NULL UNIQUE,

    -- Information
    description TEXT,

    -- Relations
    project_id UUID NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign Keys
    FOREIGN KEY(project_id)
    REFERENCES projects(id)
    ON DELETE CASCADE
);

CREATE TABLE application_releases (
    -- Identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(256) NOT NULL,
    version_code INTEGER NOT NULL,
    version_name VARCHAR(256) NOT NULL,

    -- Information
    release_note TEXT,
    environment release_environment NOT NULL DEFAULT 'development',

    -- Relations
    application_id UUID NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign Keys
    FOREIGN KEY(application_id)
    REFERENCES applications(id)
    ON DELETE CASCADE,

    -- Constraints
    CONSTRAINT unique_application_release UNIQUE (application_id, version_code, environment)
);

CREATE TABLE artifacts (
    -- Identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_url VARCHAR(512) NOT NULL,
    sha256_hash VARCHAR(512) NOT NULL,

    -- Information
    file_size BIGINT NOT NULL,
    file_type VARCHAR(256) NOT NULL,
    abi VARCHAR(256), -- e.g. arm64-v8a, x86_64

    -- Relations
    release_id UUID NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign Keys
    FOREIGN KEY(release_id)
    REFERENCES application_releases(id)
    ON DELETE CASCADE,

    -- Constraints
    CONSTRAINT unique_artifact UNIQUE (release_id, abi)
);

-- +goose Down
DROP TABLE IF EXISTS artifacts;
DROP TABLE IF EXISTS application_releases;
DROP TABLE IF EXISTS applications;
DROP TYPE IF EXISTS release_environment;
