package ranking

import (
	"context"
	"fmt"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RefreshLeaderboard recomputes and persists composite scores for all completed
// attempts for the given quiz, then assigns ranks.
// This is called asynchronously after a session completes.
func RefreshLeaderboard(ctx context.Context, pool *pgxpool.Pool, quizID string, timeLimitSeconds int) error {
	rows, err := pool.Query(ctx, `
		SELECT a.id, a.user_id, u.username, u.elo_rating,
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

	var entries []*models.LeaderboardEntry
	var totalQuestions int
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM questions WHERE quiz_id = $1", quizID,
	).Scan(&totalQuestions)
	if err != nil {
		return fmt.Errorf("count questions: %w", err)
	}

	for rows.Next() {
		e := &models.LeaderboardEntry{}
		if err := rows.Scan(
			&e.UserID, &e.UserID, &e.Username, &e.EloRating,
			&e.Score, &e.TimeElapsed, &e.Accuracy, &e.AttemptedAt,
		); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}
		e.CompositeScore = ComputeComposite(e.Score, timeLimitSeconds, e.TimeElapsed, totalQuestions)
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate rows: %w", err)
	}

	RankEntries(entries)

	// Upsert leaderboard rows.
	for _, e := range entries {
		_, err := pool.Exec(ctx, `
			INSERT INTO leaderboard (user_id, quiz_id, score, time_elapsed_seconds,
			                         accuracy, composite_score, rank, attempted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (user_id, quiz_id) DO UPDATE
			SET score = EXCLUDED.score,
			    time_elapsed_seconds = EXCLUDED.time_elapsed_seconds,
			    accuracy = EXCLUDED.accuracy,
			    composite_score = EXCLUDED.composite_score,
			    rank = EXCLUDED.rank,
			    attempted_at = EXCLUDED.attempted_at
		`, e.UserID, quizID, e.Score, e.TimeElapsed,
			e.Accuracy, e.CompositeScore, e.Rank, e.AttemptedAt)
		if err != nil {
			return fmt.Errorf("upsert leaderboard: %w", err)
		}
	}
	return nil
}
