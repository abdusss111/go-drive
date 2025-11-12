CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS buckets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id, name)
);

CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bucket_id UUID NOT NULL REFERENCES buckets(id) ON DELETE CASCADE,
    object_name TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    content_type TEXT,
    checksum TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (bucket_id, object_name)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    UNIQUE (user_id, token_hash)
);

CREATE TABLE IF NOT EXISTS usage_snapshots (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_bytes BIGINT NOT NULL,
    file_count BIGINT NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bucket_usage (
    bucket_id UUID PRIMARY KEY REFERENCES buckets(id) ON DELETE CASCADE,
    total_bytes BIGINT NOT NULL DEFAULT 0,
    file_count BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_files_bucket ON files (bucket_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_usage_snapshots_user ON usage_snapshots (user_id, collected_at DESC);

INSERT INTO users (email, password_hash, display_name, is_admin)
VALUES ('admin@godrive.local', '$2a$12$IC8vu5BZ.ZILB.U72hVVf.fZjs51Rm/1vHRjNp/H7NnjoR9f6NVqK', 'Administrator', TRUE)
ON CONFLICT (email) DO NOTHING;

