package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AttemptRepo handles database operations for quiz attempts and answers.
type AttemptRepo struct {
	pool *pgxpool.Pool
}

// NewAttemptRepo creates a new AttemptRepo.
func NewAttemptRepo(pool *pgxpool.Pool) *AttemptRepo {
	return &AttemptRepo{pool: pool}
}

// Create inserts a new attempt record.
func (r *AttemptRepo) Create(ctx context.Context, a *models.Attempt) (*models.Attempt, error) {
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

// GetByID returns an attempt by its ID.
func (r *AttemptRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Attempt, error) {
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

// SaveAnswer records a single question answer for an attempt.
func (r *AttemptRepo) SaveAnswer(ctx context.Context, ans *models.Answer) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO answers (id, attempt_id, question_id, selected_option, is_correct, answered_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (attempt_id, question_id) DO UPDATE
		SET selected_option = EXCLUDED.selected_option,
		    is_correct = EXCLUDED.is_correct,
		    answered_at = EXCLUDED.answered_at
	`, ans.ID, ans.AttemptID, ans.QuestionID, ans.SelectedOpt, ans.IsCorrect, ans.AnsweredAt)
	if err != nil {
		return fmt.Errorf("save answer: %w", err)
	}
	return nil
}

// GetAnswers returns all answers for a given attempt.
func (r *AttemptRepo) GetAnswers(ctx context.Context, attemptID uuid.UUID) ([]*models.Answer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, attempt_id, question_id, selected_option, is_correct, answered_at
		FROM answers WHERE attempt_id = $1
	`, attemptID)
	if err != nil {
		return nil, fmt.Errorf("get answers: %w", err)
	}
	defer rows.Close()

	var answers []*models.Answer
	for rows.Next() {
		a := &models.Answer{}
		if err := rows.Scan(
			&a.ID, &a.AttemptID, &a.QuestionID, &a.SelectedOpt, &a.IsCorrect, &a.AnsweredAt,
		); err != nil {
			return nil, fmt.Errorf("scan answer: %w", err)
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

// Complete finalises an attempt with the computed results.
func (r *AttemptRepo) Complete(ctx context.Context, id uuid.UUID, score, rawScore, correct, wrong, unanswered, timeElapsed, eloDelta int) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE attempts
		SET status = 'completed',
		    score = $2,
		    raw_score = $3,
		    correct_answers = $4,
		    wrong_answers = $5,
		    unanswered = $6,
		    time_elapsed_seconds = $7,
		    elo_delta = $8,
		    completed_at = $9
		WHERE id = $1
	`, id, score, rawScore, correct, wrong, unanswered, timeElapsed, eloDelta, now)
	if err != nil {
		return fmt.Errorf("complete attempt: %w", err)
	}
	return nil
}

// GetUserHistory returns all completed attempts for a user, newest first.
func (r *AttemptRepo) GetUserHistory(ctx context.Context, userID uuid.UUID) ([]*models.Attempt, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, quiz_id, status, score, raw_score,
		       correct_answers, wrong_answers, unanswered,
		       time_elapsed_seconds, elo_delta, started_at, completed_at
		FROM attempts
		WHERE user_id = $1 AND status = 'completed'
		ORDER BY completed_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get user history: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		a := &models.Attempt{}
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.QuizID, &a.Status,
			&a.Score, &a.RawScore, &a.CorrectAnswers, &a.WrongAnswers, &a.Unanswered,
			&a.TimeElapsed, &a.EloDelta, &a.StartedAt, &a.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan attempt: %w", err)
		}
		attempts = append(attempts, a)
	}
	return attempts, rows.Err()
}

// GetAllScoresForQuiz returns composite scores for all completed attempts on a quiz.
func (r *AttemptRepo) GetAllScoresForQuiz(ctx context.Context, quizID uuid.UUID) ([]float64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT composite_score FROM leaderboard WHERE quiz_id = $1
	`, quizID)
	if err != nil {
		return nil, fmt.Errorf("get scores: %w", err)
	}
	defer rows.Close()

	var scores []float64
	for rows.Next() {
		var s float64
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	return scores, rows.Err()
}
