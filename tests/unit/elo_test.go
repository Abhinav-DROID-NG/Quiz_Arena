package unit_test

import (
	"testing"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/engine"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
)

func TestComputeElo_CorrectEasyQuestion(t *testing.T) {
	// A player at 1000 Elo answering an easy question (base 1000) correctly
	// should gain rating points.
	result := engine.ComputeElo(1000, models.DifficultyEasy, 1.0)
	if result.Delta <= 0 {
		t.Errorf("expected positive elo delta for correct easy answer, got %d", result.Delta)
	}
	if result.NewRating != result.OldRating+result.Delta {
		t.Errorf("NewRating mismatch: got %d, expected %d", result.NewRating, result.OldRating+result.Delta)
	}
}

func TestComputeElo_WrongHardQuestion(t *testing.T) {
	// A player at 1000 Elo answering a hard question (base 1500) wrongly
	// should lose rating points.
	result := engine.ComputeElo(1000, models.DifficultyHard, 0.0)
	if result.Delta >= 0 {
		t.Errorf("expected negative elo delta for wrong hard answer, got %d", result.Delta)
	}
}

func TestComputeElo_HardKFactor(t *testing.T) {
	// Hard questions have a higher K-factor so the absolute delta should be
	// larger than for normal questions given equal conditions.
	hardResult := engine.ComputeElo(1200, models.DifficultyHard, 1.0)
	easyResult := engine.ComputeElo(1200, models.DifficultyEasy, 1.0)

	if hardResult.Delta <= easyResult.Delta {
		t.Errorf("hard K-factor should produce larger delta: hard=%d, easy=%d",
			hardResult.Delta, easyResult.Delta)
	}
}

func TestComputeElo_Floor(t *testing.T) {
	// Elo should never drop below 100.
	result := engine.ComputeElo(100, models.DifficultyHard, 0.0)
	if result.NewRating < 100 {
		t.Errorf("elo below floor: got %d", result.NewRating)
	}
}

func TestApplyEloToAttempt(t *testing.T) {
	inputs := []struct {
		Difficulty models.Difficulty
		Correct    bool
	}{
		{models.DifficultyEasy, true},
		{models.DifficultyMedium, false},
		{models.DifficultyHard, true},
	}

	result := engine.ApplyEloToAttempt(1200, inputs)
	if result.OldRating != 1200 {
		t.Errorf("expected OldRating=1200, got %d", result.OldRating)
	}
	if result.NewRating < 100 {
		t.Errorf("NewRating below floor: %d", result.NewRating)
	}
	if result.Delta != result.NewRating-result.OldRating {
		t.Errorf("Delta mismatch: delta=%d, new-old=%d", result.Delta, result.NewRating-result.OldRating)
	}
}
