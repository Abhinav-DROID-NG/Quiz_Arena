CREATE TABLE IF NOT EXISTS leaderboard (
    user_id               UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_id               UUID        NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    score                 INTEGER     NOT NULL DEFAULT 0,
    time_elapsed_seconds  INTEGER     NOT NULL DEFAULT 0,
    accuracy              FLOAT       NOT NULL DEFAULT 0,
    composite_score       FLOAT       NOT NULL DEFAULT 0,
    rank                  INTEGER     NOT NULL DEFAULT 0,
    attempted_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, quiz_id)
);
