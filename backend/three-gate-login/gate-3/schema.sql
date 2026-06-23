-- Users table
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS webauthn_users (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username     TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Credentials table (one user may have multiple registered devices)
CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id             BYTEA PRIMARY KEY,         -- credential.ID (bytes)
    user_id        UUID REFERENCES webauthn_users(id) ON DELETE CASCADE,
    public_key     JSONB NOT NULL,            -- full credential JSON
    sign_count     BIGINT NOT NULL DEFAULT 0, -- replay counter
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    last_used_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_webauthn_credentials_user_id ON webauthn_credentials(user_id);
