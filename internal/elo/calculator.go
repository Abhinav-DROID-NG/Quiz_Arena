package elo

import (
	"math"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
)

// Update holds the Elo delta computed for a single question or full attempt.
type Update struct {
	OldRating int
	NewRating int
	Delta     int
}

// ComputeQuestion updates a user's Elo after answering a single question.
// actual must be 1.0 for correct and 0.0 for wrong/skipped.
func ComputeQuestion(currentRating int, difficulty models.Difficulty, actual float64) Update {
	questionElo := BaseEloForDifficulty(difficulty)
	exp := expectedScore(currentRating, questionElo)
	k := KFactor(difficulty)
	delta := int(math.Round(k * (actual - exp)))
	newRating := currentRating + delta
	if newRating < RatingFloor {
		newRating = RatingFloor
	}
	return Update{
		OldRating: currentRating,
		NewRating: newRating,
		Delta:     newRating - currentRating,
	}
}

// expectedScore returns the classical Elo expected score for a player vs. an opponent.
func expectedScore(playerRating, opponentRating int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(opponentRating-playerRating)/400.0))
}
