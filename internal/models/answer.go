package models

import (
	"time"

	"github.com/google/uuid"
)

// UserAnswer records a single question response within an attempt, with response timing.
type UserAnswer struct {
	ID           uuid.UUID `json:"id" db:"id"`
	AttemptID    uuid.UUID `json:"attempt_id" db:"attempt_id"`
	QuestionID   uuid.UUID `json:"question_id" db:"question_id"`
	SelectedOpt  string    `json:"selected_option" db:"selected_option"` // "A","B","C","D" or ""
	IsCorrect    bool      `json:"is_correct" db:"is_correct"`
	ResponseTime int       `json:"response_time_ms" db:"response_time_ms"` // milliseconds
	AnsweredAt   time.Time `json:"answered_at" db:"answered_at"`
}

// AnswerSubmission is sent by the client to submit an answer for a question.
type AnswerSubmission struct {
	QuestionID   uuid.UUID `json:"question_id" validate:"required"`
	SelectedOpt  string    `json:"selected_option"`  // empty string means skipped
	ResponseTime int       `json:"response_time_ms"` // client-reported time in ms (validated server-side)
}

// AnswerResponse is returned after recording an answer (does NOT reveal correctness).
type AnswerResponse struct {
	Recorded    bool   `json:"recorded"`
	NextQuestion bool  `json:"has_next"`
	Message     string `json:"message,omitempty"`
}
