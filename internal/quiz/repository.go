package quiz

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/question"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for quiz sessions.
type Repository struct {
	pool    *pgxpool.Pool
	// In-memory session store (replace with Redis/DB for multi-instance deployments).
	mu       sync.RWMutex
	sessions map[uuid.UUID]*SessionState
}

// NewRepository creates a new quiz Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:     pool,
		sessions: make(map[uuid.UUID]*SessionState),
	}
}

// StoreSession persists a session state in memory.
func (r *Repository) StoreSession(s *SessionState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.AttemptID] = s
}

// GetSession retrieves an active session by attempt ID.
func (r *Repository) GetSession(id uuid.UUID) (*SessionState, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	return s, ok
}

// DeleteSession removes a session from the in-memory store.
func (r *Repository) DeleteSession(id uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id)
}

// CreateAttempt inserts a new attempt record into the database.
func (r *Repository) CreateAttempt(ctx context.Context, a *models.Attempt) (*models.Attempt, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO attempts (id, user_id, quiz_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, quiz_id, status, score, raw_score,
		          correct_answers, wrong_answers, unanswered,
		          time_elapsed_seconds, elo_delta, started_at, completed_at
	`, a.ID, a.UserID, a.QuizID, a.Status,
	).Scan(
		&a.ID, &a.UserID, &a.QuizID, &a.Status,
		&a.Score, &a.RawScore, &a.CorrectAnswers, &a.WrongAnswers, &a.Unanswered,
		&a.TimeElapsed, &a.EloDelta, &a.StartedAt, &a.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create attempt: %w", err)
	}
	return a, nil
}

// GetAttemptByID returns an attempt record from the database.
func (r *Repository) GetAttemptByID(ctx context.Context, id uuid.UUID) (*models.Attempt, error) {
	a := &models.Attempt{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, quiz_id, status, score, raw_score,
		       correct_answers, wrong_answers, unanswered,
		       time_elapsed_seconds, elo_delta, started_at, completed_at
		FROM attempts WHERE id = $1
	`, id).Scan(
		&a.ID, &a.UserID, &a.QuizID, &a.Status,
		&a.Score, &a.RawScore, &a.CorrectAnswers, &a.WrongAnswers, &a.Unanswered,
		&a.TimeElapsed, &a.EloDelta, &a.StartedAt, &a.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get attempt: %w", err)
	}
	return a, nil
}

// SaveAnswer records a user answer.
func (r *Repository) SaveAnswer(ctx context.Context, ans *models.UserAnswer) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_answers
			(id, attempt_id, question_id, selected_option, is_correct, response_time_ms, answered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (attempt_id, question_id) DO UPDATE
		SET selected_option  = EXCLUDED.selected_option,
		    is_correct        = EXCLUDED.is_correct,
		    response_time_ms  = EXCLUDED.response_time_ms,
		    answered_at       = EXCLUDED.answered_at
	`, ans.ID, ans.AttemptID, ans.QuestionID, ans.SelectedOpt, ans.IsCorrect, ans.ResponseTime, ans.AnsweredAt)
	if err != nil {
		return fmt.Errorf("save answer: %w", err)
	}
	return nil
}

// CompleteAttempt finalises the attempt in the database.
func (r *Repository) CompleteAttempt(
	ctx context.Context,
	id uuid.UUID,
	score, rawScore, correct, wrong, unanswered, timeElapsed, eloDelta int,
) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE attempts
		SET status              = 'completed',
		    score               = $2,
		    raw_score           = $3,
		    correct_answers     = $4,
		    wrong_answers       = $5,
		    unanswered          = $6,
		    time_elapsed_seconds = $7,
		    elo_delta           = $8,
		    completed_at        = $9
		WHERE id = $1
	`, id, score, rawScore, correct, wrong, unanswered, timeElapsed, eloDelta, now)
	if err != nil {
		return fmt.Errorf("complete attempt: %w", err)
	}
	return nil
}

// GetQuizByID retrieves a quiz record.
func (r *Repository) GetQuizByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
	q := &models.Quiz{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, title, description, created_by, time_limit_seconds, is_published, created_at, updated_at
		FROM quizzes WHERE id = $1
	`, id).Scan(
		&q.ID, &q.Title, &q.Description, &q.CreatedBy,
		&q.TimeLimit, &q.IsPublished, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}
	return q, nil
}

// GetAnswersByAttempt returns all recorded answers for an attempt from the DB.
func (r *Repository) GetAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]*models.UserAnswer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, attempt_id, question_id, selected_option, is_correct, response_time_ms, answered_at
		FROM user_answers WHERE attempt_id = $1
	`, attemptID)
	if err != nil {
		return nil, fmt.Errorf("get answers: %w", err)
	}
	defer rows.Close()

	var answers []*models.UserAnswer
	for rows.Next() {
		a := &models.UserAnswer{}
		if err := rows.Scan(
			&a.ID, &a.AttemptID, &a.QuestionID, &a.SelectedOpt, &a.IsCorrect, &a.ResponseTime, &a.AnsweredAt,
		); err != nil {
			return nil, fmt.Errorf("scan answer: %w", err)
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

// UpdateUserElo updates a user's global ELO rating.
func (r *Repository) UpdateUserElo(ctx context.Context, userID uuid.UUID, newRating int) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET elo_rating = $1, updated_at = NOW() WHERE id = $2",
		newRating, userID,
	)
	if err != nil {
		return fmt.Errorf("update elo: %w", err)
	}
	return nil
}

// GetUserElo returns a user's current global ELO rating.
func (r *Repository) GetUserElo(ctx context.Context, userID uuid.UUID) (int, error) {
	var elo int
	err := r.pool.QueryRow(ctx,
		"SELECT elo_rating FROM users WHERE id = $1", userID,
	).Scan(&elo)
	if err != nil {
		return 0, fmt.Errorf("get user elo: %w", err)
	}
	return elo, nil
}

// UpsertSubjectElo creates or updates a user's per-subject ELO.
func (r *Repository) UpsertSubjectElo(ctx context.Context, userID, subjectID uuid.UUID, elo int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_subject_elo (user_id, subject_id, elo_rating)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, subject_id) DO UPDATE
		SET elo_rating = EXCLUDED.elo_rating, updated_at = NOW()
	`, userID, subjectID, elo)
	if err != nil {
		return fmt.Errorf("upsert subject elo: %w", err)
	}
	return nil
}

// GetSubjectElo returns a user's ELO for a specific subject (defaults to 1000).
func (r *Repository) GetSubjectElo(ctx context.Context, userID, subjectID uuid.UUID) (int, error) {
	var elo int
	err := r.pool.QueryRow(ctx, `
		SELECT elo_rating FROM user_subject_elo WHERE user_id = $1 AND subject_id = $2
	`, userID, subjectID).Scan(&elo)
	if err != nil {
		return 1000, nil
	}
	return elo, nil
}

// RefreshLeaderboard recomputes and upserts composite scores in the leaderboard table.
func (r *Repository) RefreshLeaderboard(ctx context.Context, quizID uuid.UUID, timeLimitSeconds int) error {
	rows, err := r.pool.Query(ctx, `
		SELECT a.user_id, u.username, u.elo_rating,
		       a.score, a.time_elapsed_seconds,
		       CASE WHEN (a.correct_answers + a.wrong_answers) > 0
		            THEN a.correct_answers::float / (a.correct_answers + a.wrong_answers)
		            ELSE 0 END AS accuracy,
		       a.started_at
		FROM attempts a
		JOIN users u ON u.id = a.user_id
		WHERE a.quiz_id = $1 AND a.status = 'completed'
	`, quizID)
	if err != nil {
		return fmt.Errorf("query attempts: %w", err)
	}
	defer rows.Close()

	type row struct {
		userID      uuid.UUID
		username    string
		elo         int
		score       int
		timeElapsed int
		accuracy    float64
		startedAt   time.Time
	}
	var rowData []row
	for rows.Next() {
		var rd row
		if err := rows.Scan(
			&rd.userID, &rd.username, &rd.elo,
			&rd.score, &rd.timeElapsed, &rd.accuracy, &rd.startedAt,
		); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}
		rowData = append(rowData, rd)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	var totalQuestions int
	_ = r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM questions WHERE quiz_id = $1", quizID,
	).Scan(&totalQuestions)

	for i, rd := range rowData {
		composite := computeComposite(rd.score, timeLimitSeconds, rd.timeElapsed, totalQuestions)
		rowData[i].score = int(composite * 1000) // store scaled for ranking
		_ = composite
	}

	for _, rd := range rowData {
		composite := computeComposite(rd.score, timeLimitSeconds, rd.timeElapsed, totalQuestions)
		_, err := r.pool.Exec(ctx, `
			INSERT INTO leaderboard (user_id, quiz_id, score, time_elapsed_seconds,
			                         accuracy, composite_score, rank, attempted_at)
			VALUES ($1, $2, $3, $4, $5, $6, 0, $7)
			ON CONFLICT (user_id, quiz_id) DO UPDATE
			SET score              = EXCLUDED.score,
			    time_elapsed_seconds = EXCLUDED.time_elapsed_seconds,
			    accuracy           = EXCLUDED.accuracy,
			    composite_score    = EXCLUDED.composite_score,
			    attempted_at       = EXCLUDED.attempted_at
		`, rd.userID, quizID, rd.score, rd.timeElapsed,
			rd.accuracy, composite, rd.startedAt)
		if err != nil {
			return fmt.Errorf("upsert leaderboard: %w", err)
		}
	}
	return nil
}

// computeComposite calculates a composite ranking score.
func computeComposite(score, timeLimitSeconds, timeElapsed, totalQuestions int) float64 {
	if totalQuestions == 0 || timeLimitSeconds == 0 {
		return 0
	}
	normScore := float64(score) / float64(totalQuestions)
	if normScore > 1 {
		normScore = 1
	}
	if normScore < 0 {
		normScore = 0
	}
	timeRatio := float64(timeElapsed) / float64(timeLimitSeconds)
	if timeRatio > 1 {
		timeRatio = 1
	}
	timeEfficiency := 1.0 - timeRatio
	return 0.6*normScore + 0.25*timeEfficiency + 0.15*normScore
}

// questionSelectorAdapter adapts the question.Selector for use within quiz.
type questionSelectorAdapter struct {
	repo     *question.Repository
	selector *question.Selector
}
