package quiz

import "time"

// PerQuestionTimeLimitS is the default time allowed for a single question (seconds).
// A value of 0 disables per-question limits.
const PerQuestionTimeLimitS = 120 // 2 minutes per question

// Timer validates timing constraints for quiz sessions.
type Timer struct {
	perQuestionS int
}

// NewTimer creates a Timer with the given per-question limit.
func NewTimer(perQuestionS int) *Timer {
	return &Timer{perQuestionS: perQuestionS}
}

// DefaultTimer returns a Timer with the default per-question limit.
func DefaultTimer() *Timer {
	return NewTimer(PerQuestionTimeLimitS)
}

// QuestionExpired returns true if the question served at questionStarted has
// exceeded the per-question time limit.
func (t *Timer) QuestionExpired(questionStarted time.Time) bool {
	if t.perQuestionS <= 0 {
		return false
	}
	return time.Since(questionStarted) > time.Duration(t.perQuestionS)*time.Second
}

// SessionExpired returns true if the session started at sessionStarted has
// exceeded the total time limit.
func (t *Timer) SessionExpired(sessionStarted time.Time, totalLimitS int) bool {
	return time.Since(sessionStarted) > time.Duration(totalLimitS)*time.Second
}

// ResponseTimeMS returns the elapsed time in milliseconds since the question
// was served, capped at the per-question limit.
func (t *Timer) ResponseTimeMS(questionStarted time.Time) int {
	elapsed := time.Since(questionStarted).Milliseconds()
	if t.perQuestionS > 0 {
		max := int64(t.perQuestionS) * 1000
		if elapsed > max {
			return int(max)
		}
	}
	return int(elapsed)
}
