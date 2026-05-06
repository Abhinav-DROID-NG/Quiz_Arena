package engine

// AnswerResult captures the outcome of a single answered question.
type AnswerResult struct {
	Correct bool
	Skipped bool // true if no option was selected
}

// ScoreResult holds the computed score breakdown for an attempt.
type ScoreResult struct {
	RawScore       int // sum of per-question marks (can be negative)
	Score          int // max(0, RawScore) — floored at zero for display
	CorrectCount   int
	WrongCount     int
	UnansweredCount int
}

// ComputeScore applies GATE-style negative marking to a slice of answer results
// and returns the aggregated score breakdown.
func ComputeScore(answers []AnswerResult) ScoreResult {
	var res ScoreResult
	for _, a := range answers {
		mark := Mark(a.Correct, a.Skipped)
		res.RawScore += mark
		switch {
		case a.Correct:
			res.CorrectCount++
		case a.Skipped:
			res.UnansweredCount++
		default:
			res.WrongCount++
		}
	}
	res.Score = res.RawScore
	if res.Score < 0 {
		res.Score = 0
	}
	return res
}
