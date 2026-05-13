package quiz

import (
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
)

// StartRequest is the JSON body for POST /api/sessions/start.
type StartRequest struct {
	QuizID    uuid.UUID  `json:"quiz_id"`
	SubjectID *uuid.UUID `json:"subject_id,omitempty"`
}

// StartResponse is returned when a session is created.
// The first question is included; subsequent questions are served one at a time.
type StartResponse struct {
	SessionID   uuid.UUID                  `json:"session_id"`
	QuizTitle   string                     `json:"quiz_title"`
	TotalCount  int                        `json:"total_questions"`
	CurrentIdx  int                        `json:"current_index"`
	Question    *models.QuestionForClient  `json:"question"`
	TimeLimitS  int                        `json:"time_limit_seconds"`
	StartedAt   string                     `json:"started_at"`
}

// AnswerRequest is the JSON body for POST /api/sessions/{id}/answer.
type AnswerRequest struct {
	// OptionIndex is the 0-based index of the selected option from the shuffled Options slice.
	// Use -1 to skip the question.
	OptionIndex  int `json:"option_index"`
	ResponseTime int `json:"response_time_ms"` // client-reported, used for analytics only
}

// AnswerResponse is returned after recording an answer.
type AnswerResponse struct {
	Recorded    bool                      `json:"recorded"`
	HasNext     bool                      `json:"has_next"`
	NextQuestion *models.QuestionForClient `json:"next_question,omitempty"`
	CurrentIdx  int                       `json:"current_index"`
}

// CompleteResponse is returned when a session is finalised.
type CompleteResponse struct {
	Attempt    *models.Attempt `json:"attempt"`
	Rank       int             `json:"rank"`
	Percentile float64         `json:"percentile"`
	EloDelta   int             `json:"elo_delta"`
	NewElo     int             `json:"new_elo"`
}
