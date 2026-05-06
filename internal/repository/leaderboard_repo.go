package repository

import (
	"context"
	"fmt"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LeaderboardRepo handles leaderboard data access.
type LeaderboardRepo struct {
	pool *pgxpool.Pool
}

// NewLeaderboardRepo creates a new LeaderboardRepo.
func NewLeaderboardRepo(pool *pgxpool.Pool) *LeaderboardRepo {
	return &LeaderboardRepo{pool: pool}
}

// GetQuizLeaderboard returns the ranked leaderboard for a specific quiz.
func (r *LeaderboardRepo) GetQuizLeaderboard(ctx context.Context, quizID uuid.UUID) ([]*models.LeaderboardEntry, error) {
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

// GetGlobalLeaderboard returns the global top users ranked by Elo rating.
func (r *LeaderboardRepo) GetGlobalLeaderboard(ctx context.Context, limit int) ([]*models.GlobalLeaderboardEntry, error) {
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

// GetUserRankForQuiz returns the current rank of a user in a specific quiz leaderboard.
func (r *LeaderboardRepo) GetUserRankForQuiz(ctx context.Context, userID, quizID uuid.UUID) (int, error) {
	var rank int
	err := r.pool.QueryRow(ctx,
		"SELECT rank FROM leaderboard WHERE user_id = $1 AND quiz_id = $2",
		userID, quizID,
	).Scan(&rank)
	if err != nil {
		return 0, fmt.Errorf("get user rank: %w", err)
	}
	return rank, nil
}
