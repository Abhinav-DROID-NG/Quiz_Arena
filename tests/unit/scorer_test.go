package unit_test

import (
	"testing"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/engine"
)

func TestMark_Correct(t *testing.T) {
	if got := engine.Mark(true, false); got != 1 {
		t.Errorf("expected 1 for correct answer, got %d", got)
	}
}

func TestMark_Wrong(t *testing.T) {
	if got := engine.Mark(false, false); got != -1 {
		t.Errorf("expected -1 for wrong answer, got %d", got)
	}
}

func TestMark_Skipped(t *testing.T) {
	if got := engine.Mark(false, true); got != 0 {
		t.Errorf("expected 0 for skipped answer, got %d", got)
	}
}

func TestComputeScore_AllCorrect(t *testing.T) {
	answers := []engine.AnswerResult{
		{Correct: true, Skipped: false},
		{Correct: true, Skipped: false},
		{Correct: true, Skipped: false},
	}
	result := engine.ComputeScore(answers)
	if result.Score != 3 {
		t.Errorf("expected score 3, got %d", result.Score)
	}
	if result.CorrectCount != 3 {
		t.Errorf("expected 3 correct, got %d", result.CorrectCount)
	}
	if result.WrongCount != 0 {
		t.Errorf("expected 0 wrong, got %d", result.WrongCount)
	}
}

func TestComputeScore_NegativeMarkingFloor(t *testing.T) {
	// More wrong answers than correct — raw score negative but Score clamped to 0.
	answers := []engine.AnswerResult{
		{Correct: false, Skipped: false},
		{Correct: false, Skipped: false},
		{Correct: true, Skipped: false},
	}
	result := engine.ComputeScore(answers)
	if result.RawScore != -1 {
		t.Errorf("expected raw score -1, got %d", result.RawScore)
	}
	if result.Score != 0 {
		t.Errorf("expected displayed score 0, got %d", result.Score)
	}
}

func TestComputeScore_Mixed(t *testing.T) {
	answers := []engine.AnswerResult{
		{Correct: true, Skipped: false},  // +1
		{Correct: false, Skipped: false}, // -1
		{Correct: false, Skipped: true},  // 0
		{Correct: true, Skipped: false},  // +1
	}
	result := engine.ComputeScore(answers)
	if result.RawScore != 1 {
		t.Errorf("expected raw score 1, got %d", result.RawScore)
	}
	if result.Score != 1 {
		t.Errorf("expected score 1, got %d", result.Score)
	}
	if result.CorrectCount != 2 {
		t.Errorf("expected 2 correct, got %d", result.CorrectCount)
	}
	if result.WrongCount != 1 {
		t.Errorf("expected 1 wrong, got %d", result.WrongCount)
	}
	if result.UnansweredCount != 1 {
		t.Errorf("expected 1 unanswered, got %d", result.UnansweredCount)
	}
}

func TestComputeScore_AllSkipped(t *testing.T) {
	answers := []engine.AnswerResult{
		{Correct: false, Skipped: true},
		{Correct: false, Skipped: true},
	}
	result := engine.ComputeScore(answers)
	if result.Score != 0 {
		t.Errorf("expected score 0 for all skipped, got %d", result.Score)
	}
	if result.UnansweredCount != 2 {
		t.Errorf("expected 2 unanswered, got %d", result.UnansweredCount)
	}
}
