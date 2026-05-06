package models

import (
	"time"

	"github.com/google/uuid"
)

// LeaderboardEntry represents a single entry in the leaderboard.
type LeaderboardEntry struct {
	Rank           int       `json:"rank" db:"rank"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Username       string    `json:"username" db:"username"`
	EloRating      int       `json:"elo_rating" db:"elo_rating"`
	Score          int       `json:"score" db:"score"`
	TimeElapsed    int       `json:"time_elapsed_seconds" db:"time_elapsed_seconds"`
	Accuracy       float64   `json:"accuracy" db:"accuracy"`
	CompositeScore float64   `json:"composite_score" db:"composite_score"`
	AttemptedAt    time.Time `json:"attempted_at" db:"attempted_at"`
}

// GlobalLeaderboardEntry represents a user in the global ranking.
type GlobalLeaderboardEntry struct {
	Rank      int       `json:"rank" db:"rank"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	EloRating int       `json:"elo_rating" db:"elo_rating"`
	TotalScore int      `json:"total_score" db:"total_score"`
	QuizCount  int      `json:"quiz_count" db:"quiz_count"`
}
