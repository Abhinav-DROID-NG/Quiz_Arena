-- 008: Create subjects table and user_subject_elo table

CREATE TABLE IF NOT EXISTS subjects (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_subject_elo (
    user_id    UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    subject_id UUID        NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    elo_rating INTEGER     NOT NULL DEFAULT 1000,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, subject_id)
);

CREATE INDEX IF NOT EXISTS idx_user_subject_elo_subject
    ON user_subject_elo (subject_id, elo_rating DESC);
