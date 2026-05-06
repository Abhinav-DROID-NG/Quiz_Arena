CREATE TYPE attempt_status AS ENUM ('active', 'completed', 'abandoned');

CREATE TABLE IF NOT EXISTS attempts (
    id                    UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id               UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_id               UUID           NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    status                attempt_status NOT NULL DEFAULT 'active',
    score                 INTEGER        NOT NULL DEFAULT 0,
    raw_score             INTEGER        NOT NULL DEFAULT 0,
    correct_answers       INTEGER        NOT NULL DEFAULT 0,
    wrong_answers         INTEGER        NOT NULL DEFAULT 0,
    unanswered            INTEGER        NOT NULL DEFAULT 0,
    time_elapsed_seconds  INTEGER        NOT NULL DEFAULT 0,
    elo_delta             INTEGER        NOT NULL DEFAULT 0,
    started_at            TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    completed_at          TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS answers (
    id               UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    attempt_id       UUID        NOT NULL REFERENCES attempts(id) ON DELETE CASCADE,
    question_id      UUID        NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    selected_option  CHAR(1)     NOT NULL DEFAULT '',
    is_correct       BOOLEAN     NOT NULL DEFAULT FALSE,
    answered_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (attempt_id, question_id)
);
