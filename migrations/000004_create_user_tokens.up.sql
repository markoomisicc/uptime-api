CREATE TABLE user_tokens (
    id         UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    purpose    TEXT NOT NULL CHECK (purpose IN ('email_verification')),
    token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX user_tokens_user_id_idx ON user_tokens (user_id);