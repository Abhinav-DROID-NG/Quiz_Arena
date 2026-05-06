-- Index on (user_id, quiz_id) for attempt lookups.
CREATE INDEX IF NOT EXISTS idx_attempts_user_quiz ON attempts(user_id, quiz_id);

-- Index on (quiz_id, difficulty) for adaptive question selection.
CREATE INDEX IF NOT EXISTS idx_questions_quiz_difficulty ON questions(quiz_id, difficulty);

-- Leaderboard query indexes.
CREATE INDEX IF NOT EXISTS idx_leaderboard_quiz_rank ON leaderboard(quiz_id, rank ASC);
CREATE INDEX IF NOT EXISTS idx_leaderboard_composite ON leaderboard(quiz_id, composite_score DESC);

-- Index for user email lookups during login.
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Index for Elo-based global ranking.
CREATE INDEX IF NOT EXISTS idx_users_elo ON users(elo_rating DESC);

-- Index for answer lookups by attempt.
CREATE INDEX IF NOT EXISTS idx_answers_attempt ON answers(attempt_id);
