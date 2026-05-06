package models

import (
	"time"

	"github.com/google/uuid"
)

// Difficulty represents a question difficulty tier.
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// Quiz represents a collection of questions.
type Quiz struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	TimeLimit   int        `json:"time_limit_seconds" db:"time_limit_seconds"` // seconds
	IsPublished bool       `json:"is_published" db:"is_published"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateQuizRequest holds input for creating a quiz.
type CreateQuizRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description"`
	TimeLimit   int    `json:"time_limit_seconds" validate:"required,min=60"`
}
