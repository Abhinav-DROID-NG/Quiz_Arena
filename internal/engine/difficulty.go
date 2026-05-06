package engine

import "github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"

// SelectDifficulty returns the appropriate question difficulty tier
// for a user with the given Elo rating.
//
// Thresholds:
//   - Elo < 1100  → easy
//   - 1100 ≤ Elo < 1350 → medium
//   - Elo ≥ 1350  → hard
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
