package engine

import (
	"math"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
)

const (
	kFactorNormal = 32.0
	kFactorHard   = 48.0
)

// EloUpdate holds the Elo delta computed after an attempt.
type EloUpdate struct {
	OldRating int
	NewRating int
	Delta     int
}

// expectedScore returns the expected score of the player against an opponent
// with the given opponent rating, using the standard Elo formula.
func expectedScore(playerRating, opponentRating int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(opponentRating-playerRating)/400.0))
}

// kFactor returns the K-factor for the given question difficulty.
func kFactor(difficulty models.Difficulty) float64 {
	if difficulty == models.DifficultyHard {
		return kFactorHard
	}
	return kFactorNormal
}

// ComputeElo updates a user's Elo rating after answering a question.
// The question's implicit Elo (baseQuestionElo) represents how hard it is
// (easy ≈ 1000, medium ≈ 1200, hard ≈ 1500).
// actual should be 1.0 for a correct answer and 0.0 for wrong/skipped.
func ComputeElo(currentRating int, difficulty models.Difficulty, actual float64) EloUpdate {
	questionElo := difficultyBaseElo(difficulty)
	exp := expectedScore(currentRating, questionElo)
	k := kFactor(difficulty)
	delta := int(math.Round(k * (actual - exp)))
	newRating := currentRating + delta
	if newRating < 100 {
		newRating = 100 // floor to avoid negative ratings
	}
	return EloUpdate{
		OldRating: currentRating,
		NewRating: newRating,
		Delta:     newRating - currentRating,
	}
}

// difficultyBaseElo maps difficulty tiers to representative Elo values.
func difficultyBaseElo(d models.Difficulty) int {
	switch d {
	case models.DifficultyHard:
		return 1500
	case models.DifficultyMedium:
		return 1200
	default: // easy
		return 1000
	}
}

// ApplyEloToAttempt computes the aggregate Elo delta across all questions in an attempt.
// results maps question difficulty to whether the user answered correctly.
func ApplyEloToAttempt(currentRating int, results []struct {
	Difficulty models.Difficulty
	Correct    bool
}) EloUpdate {
	rating := currentRating
	for _, r := range results {
		actual := 0.0
		if r.Correct {
			actual = 1.0
		}
		upd := ComputeElo(rating, r.Difficulty, actual)
		rating = upd.NewRating
	}
	return EloUpdate{
		OldRating: currentRating,
		NewRating: rating,
		Delta:     rating - currentRating,
	}
}
