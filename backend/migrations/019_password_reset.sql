BEGIN;
ALTER TABLE users
    ADD COLUMN session_version BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN must_change_password BOOLEAN NOT NULL DEFAULT FALSE;
CREATE TABLE password_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    changed_by UUID NOT NULL REFERENCES users(id),
    event VARCHAR(20) NOT NULL CHECK (event IN ('reset', 'change')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMIT;
