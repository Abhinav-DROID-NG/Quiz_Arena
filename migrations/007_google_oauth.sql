-- 007: Add Google OAuth fields to users table

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS google_id   TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS provider    TEXT         NOT NULL DEFAULT 'local',
    ADD COLUMN IF NOT EXISTS avatar_url  TEXT         NOT NULL DEFAULT '';

-- Create unique index on google_id (partial — only when set)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id
    ON users (google_id)
    WHERE google_id != '';
