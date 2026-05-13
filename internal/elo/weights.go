package elo

import "github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"

// Rating boundaries and defaults.
const (
	RatingFloor   = 100  // minimum achievable Elo
	DefaultRating = 1000 // starting Elo for new users
)

// K-factors control the magnitude of Elo changes per question.
const (
	KFactorEasy   = 24.0
	KFactorMedium = 32.0
	KFactorHard   = 48.0
)

// Outcome values passed as the "actual" score in the Elo formula.
const (
	CorrectOutcome = 1.0
	WrongOutcome   = 0.0
	SkippedOutcome = 0.0
)

// GATE-style marking scheme.
const (
	MarkCorrect = 1
	MarkWrong   = -1 // negative marking
)

// Speed thresholds for the guessing penalty.
const (
	// If a wrong answer is submitted faster than this (ms), a speed penalty is applied.
	GuessThresholdMS = 3000 // 3 seconds
)

// KFactor returns the K-factor for a given difficulty tier.
func KFactor(d models.Difficulty) float64 {
	switch d {
	case models.DifficultyHard:
		return KFactorHard
	case models.DifficultyMedium:
		return KFactorMedium
	default:
		return KFactorEasy
	}
}

// SpeedPenalty returns an additional "actual" offset applied to wrong answers
// that were submitted very quickly, to penalise guessing.
// A fast wrong answer receives a larger penalty (more negative Elo shift).
func SpeedPenalty(responseTimeMS int) float64 {
	if responseTimeMS < GuessThresholdMS && responseTimeMS > 0 {
		// Scale penalty from 0.0 (at threshold) to -0.2 (at 0 ms).
		fraction := 1.0 - float64(responseTimeMS)/float64(GuessThresholdMS)
		return -0.2 * fraction
	}
	return 0.0
}
