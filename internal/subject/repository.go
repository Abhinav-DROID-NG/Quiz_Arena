package subject

import (
	"context"
	"fmt"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for subjects.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new subject Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// List returns all subjects ordered by name.
func (r *Repository) List(ctx context.Context) ([]*models.Subject, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, description, created_at
		FROM subjects
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		s := &models.Subject{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}
		subjects = append(subjects, s)
	}
	return subjects, rows.Err()
}

// GetByID returns a subject by its primary key.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subject, error) {
	s := &models.Subject{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, created_at
		FROM subjects WHERE id = $1
	`, id).Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get subject: %w", err)
	}
	return s, nil
}

// Create inserts a new subject.
func (r *Repository) Create(ctx context.Context, s *models.Subject) (*models.Subject, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO subjects (id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, created_at
	`, s.ID, s.Name, s.Description).Scan(
		&s.ID, &s.Name, &s.Description, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create subject: %w", err)
	}
	return s, nil
}

// GetUserSubjectElo returns a user's ELO rating for a specific subject.
// Returns 1000 (default) if no record exists.
func (r *Repository) GetUserSubjectElo(ctx context.Context, userID, subjectID uuid.UUID) (int, error) {
	var elo int
	err := r.pool.QueryRow(ctx, `
		SELECT elo_rating FROM user_subject_elo
		WHERE user_id = $1 AND subject_id = $2
	`, userID, subjectID).Scan(&elo)
	if err != nil {
		return 1000, nil // default ELO
	}
	return elo, nil
}

// UpsertUserSubjectElo creates or updates a user's subject ELO.
func (r *Repository) UpsertUserSubjectElo(ctx context.Context, userID, subjectID uuid.UUID, elo int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_subject_elo (user_id, subject_id, elo_rating)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, subject_id) DO UPDATE
		SET elo_rating = EXCLUDED.elo_rating,
		    updated_at = NOW()
	`, userID, subjectID, elo)
	if err != nil {
		return fmt.Errorf("upsert subject elo: %w", err)
	}
	return nil
}

// GetSubjectLeaderboard returns the top users for a subject ordered by subject ELO.
func (r *Repository) GetSubjectLeaderboard(ctx context.Context, subjectID uuid.UUID, limit int) ([]*models.UserSubjectElo, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, subject_id, elo_rating, updated_at
		FROM user_subject_elo
		WHERE subject_id = $1
		ORDER BY elo_rating DESC
		LIMIT $2
	`, subjectID, limit)
	if err != nil {
		return nil, fmt.Errorf("get subject leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []*models.UserSubjectElo
	for rows.Next() {
		e := &models.UserSubjectElo{}
		if err := rows.Scan(&e.UserID, &e.SubjectID, &e.EloRating, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan elo entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
