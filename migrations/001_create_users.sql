-- +goose Up
CREATE TABLE users(
    -- Identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Credentials
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL UNIQUE,
    phone_number VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,

    -- Information
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    deleted_at TIMESTAMP

    -- Constraints
    -- No Constraints yet...
);

-- +goose Down
DROP TABLE IF EXISTS users;
