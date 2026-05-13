package leaderboard

import (
	"context"
	"fmt"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for leaderboards.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new leaderboard Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetQuizLeaderboard returns the ranked leaderboard for a specific quiz.
func (r *Repository) GetQuizLeaderboard(ctx context.Context, quizID uuid.UUID) ([]*models.LeaderboardEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT l.rank, l.user_id, u.username, u.elo_rating,
		       l.score, l.time_elapsed_seconds, l.accuracy, l.composite_score, l.attempted_at
		FROM leaderboard l
		JOIN users u ON u.id = l.user_id
		WHERE l.quiz_id = $1
		ORDER BY l.rank ASC
	`, quizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []*models.LeaderboardEntry
	for rows.Next() {
		e := &models.LeaderboardEntry{}
		if err := rows.Scan(
			&e.Rank, &e.UserID, &e.Username, &e.EloRating,
			&e.Score, &e.TimeElapsed, &e.Accuracy, &e.CompositeScore, &e.AttemptedAt,
		); err != nil {
			return nil, fmt.Errorf("scan leaderboard entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetGlobalLeaderboard returns the global top users ranked by ELO.
func (r *Repository) GetGlobalLeaderboard(ctx context.Context, limit int) ([]*models.GlobalLeaderboardEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ROW_NUMBER() OVER (ORDER BY u.elo_rating DESC) AS rank,
		       u.id, u.username, u.elo_rating,
		       COALESCE(SUM(a.score), 0) AS total_score,
		       COUNT(a.id) AS quiz_count
		FROM users u
		LEFT JOIN attempts a ON a.user_id = u.id AND a.status = 'completed'
		GROUP BY u.id, u.username, u.elo_rating
		ORDER BY u.elo_rating DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("get global leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []*models.GlobalLeaderboardEntry
	for rows.Next() {
		e := &models.GlobalLeaderboardEntry{}
		if err := rows.Scan(
			&e.Rank, &e.UserID, &e.Username, &e.EloRating,
			&e.TotalScore, &e.QuizCount,
		); err != nil {
			return nil, fmt.Errorf("scan global entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetSubjectLeaderboard returns users ranked by per-subject ELO.
func (r *Repository) GetSubjectLeaderboard(ctx context.Context, subjectID uuid.UUID, limit int) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ROW_NUMBER() OVER (ORDER BY se.elo_rating DESC) AS rank,
		       u.id, u.username, se.elo_rating
		FROM user_subject_elo se
		JOIN users u ON u.id = se.user_id
		WHERE se.subject_id = $1
		ORDER BY se.elo_rating DESC
		LIMIT $2
	`, subjectID, limit)
	if err != nil {
		return nil, fmt.Errorf("get subject leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []map[string]any
	for rows.Next() {
		var (
			rank      int
			userID    uuid.UUID
			username  string
			eloRating int
		)
		if err := rows.Scan(&rank, &userID, &username, &eloRating); err != nil {
			return nil, fmt.Errorf("scan subject entry: %w", err)
		}
		entries = append(entries, map[string]any{
			"rank":       rank,
			"user_id":    userID,
			"username":   username,
			"elo_rating": eloRating,
		})
	}
	return entries, rows.Err()
}
