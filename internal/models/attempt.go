package models

import (
	"time"

	"github.com/google/uuid"
)

// AttemptStatus represents the lifecycle state of a quiz session.
type AttemptStatus string

const (
	AttemptStatusActive    AttemptStatus = "active"
	AttemptStatusCompleted AttemptStatus = "completed"
	AttemptStatusAbandoned AttemptStatus = "abandoned"
)

// Attempt represents a user's quiz session.
type Attempt struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	UserID          uuid.UUID     `json:"user_id" db:"user_id"`
	QuizID          uuid.UUID     `json:"quiz_id" db:"quiz_id"`
	Status          AttemptStatus `json:"status" db:"status"`
	Score           int           `json:"score" db:"score"`
	RawScore        int           `json:"raw_score" db:"raw_score"`
	CorrectAnswers  int           `json:"correct_answers" db:"correct_answers"`
	WrongAnswers    int           `json:"wrong_answers" db:"wrong_answers"`
	Unanswered      int           `json:"unanswered" db:"unanswered"`
	TimeElapsed     int           `json:"time_elapsed_seconds" db:"time_elapsed_seconds"`
	EloDelta        int           `json:"elo_delta" db:"elo_delta"`
	StartedAt       time.Time     `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty" db:"completed_at"`
}

// Answer records a single question response within an attempt.
type Answer struct {
	ID          uuid.UUID `json:"id" db:"id"`
	AttemptID   uuid.UUID `json:"attempt_id" db:"attempt_id"`
	QuestionID  uuid.UUID `json:"question_id" db:"question_id"`
	SelectedOpt string    `json:"selected_option" db:"selected_option"` // "A","B","C","D" or ""
	IsCorrect   bool      `json:"is_correct" db:"is_correct"`
	AnsweredAt  time.Time `json:"answered_at" db:"answered_at"`
}

// StartSessionRequest holds input for starting a quiz session.
type StartSessionRequest struct {
	QuizID uuid.UUID `json:"quiz_id" validate:"required"`
}

// AnswerRequest holds a single answer submission.
type AnswerRequest struct {
	QuestionID  uuid.UUID `json:"question_id" validate:"required"`
	SelectedOpt string    `json:"selected_option"` // empty string means skipped
}

// CompleteSessionResponse is returned after finishing a quiz.
type CompleteSessionResponse struct {
	Attempt  *Attempt `json:"attempt"`
	Rank     int      `json:"rank"`
	Percentile float64 `json:"percentile"`
}
