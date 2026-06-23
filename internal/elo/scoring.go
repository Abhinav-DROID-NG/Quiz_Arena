package elo

import "github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"

// QuestionResult captures the outcome of answering a single question.
type QuestionResult struct {
	Difficulty   models.Difficulty
	Correct      bool
	Skipped      bool
	ResponseTime int // milliseconds
}

// AttemptScore summarises the scoring outcome for a complete attempt.
type AttemptScore struct {
	RawScore        int
	Score           int // max(0, RawScore)
	CorrectCount    int
	WrongCount      int
	UnansweredCount int
	EloDelta        int
}

// ApplyToAttempt computes the aggregate Elo delta for all questions in an attempt,
// applying speed bonuses/penalties to avoid incentivising reckless guessing.
func ApplyToAttempt(currentRating int, results []QuestionResult) (Update, AttemptScore) {
	rating := currentRating
	score := AttemptScore{}

	for _, r := range results {
		var actual float64

		switch {
		case r.Correct:
			actual = CorrectOutcome
			score.CorrectCount++
			score.RawScore += MarkCorrect
		case r.Skipped:
			actual = SkippedOutcome
			score.UnansweredCount++
			// No mark change for skipped questions
		default:
			// Wrong answer: apply speed penalty to discourage random guessing.
			penalty := SpeedPenalty(r.ResponseTime)
			actual = WrongOutcome + penalty // reduces loss slightly for slower wrong answers
			score.WrongCount++
			score.RawScore += MarkWrong
		}

		upd := ComputeQuestion(rating, r.Difficulty, actual)
		rating = upd.NewRating
	}

	if score.Score = score.RawScore; score.Score < 0 {
		score.Score = 0
	}
	score.EloDelta = rating - currentRating

	return Update{
		OldRating: currentRating,
		NewRating: rating,
		Delta:     score.EloDelta,
	}, score
}
