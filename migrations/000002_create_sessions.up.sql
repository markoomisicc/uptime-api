CREATE TABLE sessions (
    id         UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    user_agent TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);