CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT uuidv7(),
    email             TEXT NOT NULL,
    username          TEXT NOT NULL,
    password_hash     TEXT NOT NULL,
    first_name        TEXT NOT NULL,
    last_name         TEXT NOT NULL,
    phone             TEXT CHECK (phone ~ '^\+[1-9][0-9]{6,14}$'),
    timezone          TEXT NOT NULL DEFAULT 'UTC',
    email_verified_at TIMESTAMPTZ,
    last_login_at     TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_unique ON users (lower(email));
CREATE UNIQUE INDEX users_username_unique ON users (lower(username));

CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT uuidv7(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE roles (
    code        TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL
);

INSERT INTO roles (code, name, description) VALUES
    ('owner',  'Owner',  'Full access to all resources and settings'),
    ('admin',  'Admin',  'Manage monitors, alerts and members'),
    ('member', 'Member', 'View monitors and incidents');

CREATE TABLE organization_members (
    organization_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role            TEXT NOT NULL REFERENCES roles (code),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, user_id)
);

CREATE INDEX organization_members_user_id_idx ON organization_members (user_id);