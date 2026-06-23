package models

import (
	"time"

	"github.com/google/uuid"
)

// Subject represents a quiz subject/topic category.
type Subject struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserSubjectElo tracks per-subject ELO rating for a user.
type UserSubjectElo struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	SubjectID uuid.UUID `json:"subject_id" db:"subject_id"`
	EloRating int       `json:"elo_rating" db:"elo_rating"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateSubjectRequest holds input for creating a subject.
type CreateSubjectRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
}
