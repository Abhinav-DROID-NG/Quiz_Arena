package quiz

import (
	"time"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
)

// SessionState holds the server-side state for an active quiz session.
// This is the authoritative source of truth — clients cannot tamper with it.
type SessionState struct {
	AttemptID   uuid.UUID
	UserID      uuid.UUID
	QuizID      uuid.UUID
	SubjectID   *uuid.UUID
	Questions   []*models.Question      // ordered question list (with answers, server-side only)
	ShuffleMaps []map[int]string        // per-question shuffle mapping: shuffled index → original label
	CurrentIdx  int                     // index of the current question
	StartedAt   time.Time
	QuestionAt  time.Time               // when the current question was served
	TimeLimitS  int                     // per-quiz time limit in seconds
	Answers     map[uuid.UUID]*answerEntry // questionID → recorded answer
	Completed   bool
}

// answerEntry holds a recorded answer within a session.
type answerEntry struct {
	SelectedIdx  int   // shuffled option index selected by client (0-3)
	OriginalOpt  string // resolved original label (A/B/C/D or "skip")
	IsCorrect    bool
	ResponseTime int   // ms
}

// CurrentQuestion returns the question at the current index, or nil if done.
func (s *SessionState) CurrentQuestion() *models.Question {
	if s.CurrentIdx >= len(s.Questions) {
		return nil
	}
	return s.Questions[s.CurrentIdx]
}

// HasNext returns true if there are more questions to serve.
func (s *SessionState) HasNext() bool {
	return s.CurrentIdx < len(s.Questions)-1
}

// Advance moves to the next question and records when it was served.
func (s *SessionState) Advance() {
	s.CurrentIdx++
	s.QuestionAt = time.Now()
}

// TimerExpired returns true if the total quiz time has been exceeded.
func (s *SessionState) TimerExpired() bool {
	return time.Since(s.StartedAt) > time.Duration(s.TimeLimitS)*time.Second
}

// QuestionTimerExpired returns true if the per-question timer has been exceeded.
// questionLimitS is the per-question time in seconds (0 = no limit).
func (s *SessionState) QuestionTimerExpired(questionLimitS int) bool {
	if questionLimitS <= 0 {
		return false
	}
	return time.Since(s.QuestionAt) > time.Duration(questionLimitS)*time.Second
}

// IsOwner returns true if the session belongs to userID.
func (s *SessionState) IsOwner(userID uuid.UUID) bool {
	return s.UserID == userID
}
