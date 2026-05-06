package repository

import (
	"context"
	"fmt"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QuizRepo handles database operations for quizzes and questions.
type QuizRepo struct {
	pool *pgxpool.Pool
}

// NewQuizRepo creates a new QuizRepo.
func NewQuizRepo(pool *pgxpool.Pool) *QuizRepo {
	return &QuizRepo{pool: pool}
}

// ListPublished returns all published quizzes ordered by creation date descending.
func (r *QuizRepo) ListPublished(ctx context.Context) ([]*models.Quiz, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, description, created_by, time_limit_seconds, is_published, created_at, updated_at
		FROM quizzes
		WHERE is_published = true
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list quizzes: %w", err)
	}
	defer rows.Close()

	var quizzes []*models.Quiz
	for rows.Next() {
		q := &models.Quiz{}
		if err := rows.Scan(
			&q.ID, &q.Title, &q.Description, &q.CreatedBy,
			&q.TimeLimit, &q.IsPublished, &q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan quiz: %w", err)
		}
		quizzes = append(quizzes, q)
	}
	return quizzes, rows.Err()
}

// GetByID returns a single quiz by its ID.
func (r *QuizRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
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

// Create inserts a new quiz.
func (r *QuizRepo) Create(ctx context.Context, q *models.Quiz) (*models.Quiz, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO quizzes (id, title, description, created_by, time_limit_seconds, is_published)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, created_by, time_limit_seconds, is_published, created_at, updated_at
	`, q.ID, q.Title, q.Description, q.CreatedBy, q.TimeLimit, q.IsPublished,
	).Scan(
		&q.ID, &q.Title, &q.Description, &q.CreatedBy,
		&q.TimeLimit, &q.IsPublished, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create quiz: %w", err)
	}
	return q, nil
}

// GetQuestionsByQuiz returns all questions for a quiz.
func (r *QuizRepo) GetQuestionsByQuiz(ctx context.Context, quizID uuid.UUID) ([]*models.Question, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, quiz_id, text, option_a, option_b, option_c, option_d, answer, difficulty, elo_weight
		FROM questions WHERE quiz_id = $1
	`, quizID)
	if err != nil {
		return nil, fmt.Errorf("get questions: %w", err)
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		q := &models.Question{}
		if err := rows.Scan(
			&q.ID, &q.QuizID, &q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD,
			&q.Answer, &q.Difficulty, &q.EloWeight,
		); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

// GetQuestionsByDifficulty returns questions for a quiz filtered by difficulty.
func (r *QuizRepo) GetQuestionsByDifficulty(ctx context.Context, quizID uuid.UUID, difficulty models.Difficulty) ([]*models.Question, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, quiz_id, text, option_a, option_b, option_c, option_d, answer, difficulty, elo_weight
		FROM questions WHERE quiz_id = $1 AND difficulty = $2
	`, quizID, difficulty)
	if err != nil {
		return nil, fmt.Errorf("get questions by difficulty: %w", err)
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		q := &models.Question{}
		if err := rows.Scan(
			&q.ID, &q.QuizID, &q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD,
			&q.Answer, &q.Difficulty, &q.EloWeight,
		); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

// GetQuestionByID returns a single question including the answer.
func (r *QuizRepo) GetQuestionByID(ctx context.Context, id uuid.UUID) (*models.Question, error) {
	q := &models.Question{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, quiz_id, text, option_a, option_b, option_c, option_d, answer, difficulty, elo_weight
		FROM questions WHERE id = $1
	`, id).Scan(
		&q.ID, &q.QuizID, &q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD,
		&q.Answer, &q.Difficulty, &q.EloWeight,
	)
	if err != nil {
		return nil, fmt.Errorf("get question by id: %w", err)
	}
	return q, nil
}

// CountQuestions returns the number of questions for a quiz.
func (r *QuizRepo) CountQuestions(ctx context.Context, quizID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM questions WHERE quiz_id = $1", quizID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count questions: %w", err)
	}
	return count, nil
}
