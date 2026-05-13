package elo

import "github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"

// BaseEloForDifficulty maps a difficulty tier to a representative Elo value
// used as the "opponent" in the Elo calculation.
//
//   - Easy   → 1000
//   - Medium → 1200
//   - Hard   → 1500
func BaseEloForDifficulty(d models.Difficulty) int {
	switch d {
	case models.DifficultyHard:
		return 1500
	case models.DifficultyMedium:
		return 1200
	default: // easy
		return 1000
	}
}

// SelectDifficulty returns the appropriate question difficulty tier for a
// user with the given Elo rating.
//
//   - Elo < 1100         → easy
//   - 1100 ≤ Elo < 1350  → medium
//   - Elo ≥ 1350         → hard
func SelectDifficulty(eloRating int) models.Difficulty {
	switch {
	case eloRating >= 1350:
		return models.DifficultyHard
	case eloRating >= 1100:
		return models.DifficultyMedium
	default:
		return models.DifficultyEasy
	}
}
