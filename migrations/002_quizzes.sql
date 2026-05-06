CREATE TABLE IF NOT EXISTS quizzes (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    title               VARCHAR(200) NOT NULL,
    description         TEXT         NOT NULL DEFAULT '',
    created_by          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    time_limit_seconds  INTEGER      NOT NULL DEFAULT 1800,
    is_published        BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
