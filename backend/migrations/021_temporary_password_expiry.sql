BEGIN;
ALTER TABLE users ADD COLUMN password_reset_expires_at TIMESTAMPTZ;
-- Previously issued temporary passwords receive one final 24-hour window.
UPDATE users SET password_reset_expires_at=NOW()+INTERVAL '24 hours' WHERE must_change_password;
COMMIT;
