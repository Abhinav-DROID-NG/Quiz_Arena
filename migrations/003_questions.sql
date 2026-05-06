CREATE TYPE difficulty_tier AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE IF NOT EXISTS questions (
    id          UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id     UUID           NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    text        TEXT           NOT NULL,
    option_a    TEXT           NOT NULL,
    option_b    TEXT           NOT NULL,
    option_c    TEXT           NOT NULL,
    option_d    TEXT           NOT NULL,
    answer      CHAR(1)        NOT NULL CHECK (answer IN ('A','B','C','D')),
    difficulty  difficulty_tier NOT NULL DEFAULT 'medium',
    elo_weight  INTEGER        NOT NULL DEFAULT 1
);
