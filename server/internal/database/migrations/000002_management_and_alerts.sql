ALTER TABLE nodes
    ALTER COLUMN last_seen_at DROP NOT NULL,
    ADD COLUMN name TEXT,
    ADD COLUMN environment TEXT NOT NULL DEFAULT 'production',
    ADD COLUMN display_address TEXT;

UPDATE nodes SET name = hostname WHERE name IS NULL;
ALTER TABLE nodes ALTER COLUMN name SET NOT NULL;

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE agent_tokens (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    label TEXT NOT NULL DEFAULT 'agent',
    token_hash BYTEA NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX agent_tokens_node_id_idx ON agent_tokens (node_id, created_at DESC);

CREATE TABLE alert_events (
    id BIGSERIAL PRIMARY KEY,
    fingerprint TEXT NOT NULL,
    node_id TEXT REFERENCES nodes(id) ON DELETE SET NULL,
    rule_name TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'warning',
    state TEXT NOT NULL CHECK (state IN ('firing', 'resolved')),
    summary TEXT NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    raw_payload JSONB NOT NULL,
    UNIQUE (fingerprint, state, starts_at)
);

CREATE INDEX alert_events_received_at_idx ON alert_events (received_at DESC);
CREATE INDEX alert_events_node_received_at_idx ON alert_events (node_id, received_at DESC);
CREATE INDEX alert_events_active_idx ON alert_events (state, received_at DESC);

CREATE TABLE refresh_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX refresh_sessions_user_id_idx ON refresh_sessions (user_id, created_at DESC);
CREATE INDEX refresh_sessions_expiry_idx ON refresh_sessions (expires_at);
