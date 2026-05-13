package question

import (
	"context"
	"fmt"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for questions.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new question Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const questionSelectCols = `
	id, quiz_id, subject_id, text, option_a, option_b, option_c, option_d,
	answer, difficulty, elo_weight, latex_enabled, COALESCE(diagram_url,''),
	COALESCE(year,0), COALESCE(source,'')
`

// scanQuestion scans a full question row.
func scanQuestion(row interface {
	Scan(...any) error
}) (*models.Question, error) {
	q := &models.Question{}
	err := row.Scan(
		&q.ID, &q.QuizID, &q.SubjectID,
		&q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD,
		&q.Answer, &q.Difficulty, &q.EloWeight,
		&q.LatexEnabled, &q.DiagramURL,
		&q.Year, &q.Source,
	)
	if err != nil {
		return nil, err
	}
	return q, nil
}

// GetByID returns a single question by ID (includes answer — server-side only).
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Question, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+questionSelectCols+`
		FROM questions WHERE id = $1
	`, id)
	q, err := scanQuestion(row)
	if err != nil {
		return nil, fmt.Errorf("get question by id: %w", err)
	}
	return q, nil
}

// GetByQuiz returns all questions for a quiz.
func (r *Repository) GetByQuiz(ctx context.Context, quizID uuid.UUID) ([]*models.Question, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+questionSelectCols+`
		FROM questions WHERE quiz_id = $1
	`, quizID)
	if err != nil {
		return nil, fmt.Errorf("get questions by quiz: %w", err)
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		q, err := scanQuestion(rows)
		if err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

// GetByQuizAndDifficulty returns questions filtered by difficulty.
func (r *Repository) GetByQuizAndDifficulty(ctx context.Context, quizID uuid.UUID, difficulty models.Difficulty) ([]*models.Question, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+questionSelectCols+`
		FROM questions WHERE quiz_id = $1 AND difficulty = $2
	`, quizID, difficulty)
	if err != nil {
		return nil, fmt.Errorf("get questions by difficulty: %w", err)
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		q, err := scanQuestion(rows)
		if err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

// GetBySubject returns questions for a subject, optionally filtered by difficulty.
func (r *Repository) GetBySubject(ctx context.Context, subjectID uuid.UUID, difficulty *models.Difficulty, limit int) ([]*models.Question, error) {
	query := `SELECT ` + questionSelectCols + ` FROM questions WHERE subject_id = $1`
	args := []any{subjectID}

	if difficulty != nil {
		query += ` AND difficulty = $2`
		args = append(args, *difficulty)
	}
	if limit > 0 {
		query += fmt.Sprintf(` ORDER BY RANDOM() LIMIT %d`, limit)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get questions by subject: %w", err)
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		q, err := scanQuestion(rows)
		if err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

// Create inserts a new question.
func (r *Repository) Create(ctx context.Context, q *models.Question) (*models.Question, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO questions
			(id, quiz_id, subject_id, text, option_a, option_b, option_c, option_d,
			 answer, difficulty, elo_weight, latex_enabled, diagram_url, year, source)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING `+questionSelectCols,
		q.ID, q.QuizID, q.SubjectID, q.Text,
		q.OptionA, q.OptionB, q.OptionC, q.OptionD,
		q.Answer, q.Difficulty, q.EloWeight,
		q.LatexEnabled, q.DiagramURL, q.Year, q.Source,
	).Scan(
		&q.ID, &q.QuizID, &q.SubjectID,
		&q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD,
		&q.Answer, &q.Difficulty, &q.EloWeight,
		&q.LatexEnabled, &q.DiagramURL,
		&q.Year, &q.Source,
	)
	if err != nil {
		return nil, fmt.Errorf("create question: %w", err)
	}
	return q, nil
}

// CountByQuiz returns the number of questions for a quiz.
func (r *Repository) CountByQuiz(ctx context.Context, quizID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM questions WHERE quiz_id = $1", quizID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count questions: %w", err)
	}
	return count, nil
}

// GetSeenIDs returns the question IDs a user has already answered in previous attempts.
func (r *Repository) GetSeenIDs(ctx context.Context, userID uuid.UUID, quizID uuid.UUID) (map[uuid.UUID]bool, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ua.question_id
		FROM user_answers ua
		JOIN attempts a ON a.id = ua.attempt_id
		WHERE a.user_id = $1 AND a.quiz_id = $2 AND a.status = 'completed'
	`, userID, quizID)
	if err != nil {
		return nil, fmt.Errorf("get seen ids: %w", err)
	}
	defer rows.Close()

	seen := make(map[uuid.UUID]bool)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		seen[id] = true
	}
	return seen, rows.Err()
}
