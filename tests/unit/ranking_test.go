package unit_test

import (
	"testing"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/ranking"
	"github.com/google/uuid"
)

func TestComputeComposite_PerfectScore(t *testing.T) {
	// All questions correct, finished instantly — should produce the maximum weight for score.
	score := ranking.ComputeComposite(10, 600, 1, 10)
	if score <= 0 || score > 1 {
		t.Errorf("expected composite in (0,1], got %f", score)
	}
}

func TestComputeComposite_ZeroScore(t *testing.T) {
	score := ranking.ComputeComposite(0, 600, 600, 10)
	if score != 0 {
		t.Errorf("expected composite 0 for zero score, got %f", score)
	}
}

func TestComputeComposite_FasterIsBetter(t *testing.T) {
	fast := ranking.ComputeComposite(8, 600, 60, 10)
	slow := ranking.ComputeComposite(8, 600, 550, 10)
	if fast <= slow {
		t.Errorf("faster attempt should have higher composite: fast=%f, slow=%f", fast, slow)
	}
}

func TestRankEntries_Order(t *testing.T) {
	entries := []*models.LeaderboardEntry{
		{UserID: uuid.New(), CompositeScore: 0.5},
		{UserID: uuid.New(), CompositeScore: 0.9},
		{UserID: uuid.New(), CompositeScore: 0.7},
	}
	ranked := ranking.RankEntries(entries)
	if ranked[0].Rank != 1 {
		t.Errorf("top entry should have rank 1, got %d", ranked[0].Rank)
	}
	if ranked[0].CompositeScore != 0.9 {
		t.Errorf("top entry should have highest composite, got %f", ranked[0].CompositeScore)
	}
	if ranked[2].Rank != 3 {
		t.Errorf("bottom entry should have rank 3, got %d", ranked[2].Rank)
	}
}

func TestRankEntries_TiedScores(t *testing.T) {
	entries := []*models.LeaderboardEntry{
		{UserID: uuid.New(), CompositeScore: 0.8},
		{UserID: uuid.New(), CompositeScore: 0.8},
		{UserID: uuid.New(), CompositeScore: 0.5},
	}
	ranked := ranking.RankEntries(entries)
	if ranked[0].Rank != 1 || ranked[1].Rank != 1 {
		t.Errorf("tied entries should share rank 1: got %d and %d", ranked[0].Rank, ranked[1].Rank)
	}
	if ranked[2].Rank != 3 {
		t.Errorf("entry after tie should be rank 3, got %d", ranked[2].Rank)
	}
}

func TestPercentile_BelowAll(t *testing.T) {
	allScores := []float64{0.5, 0.6, 0.7, 0.8, 0.9}
	p := ranking.Percentile(0.3, allScores)
	if p != 0 {
		t.Errorf("score below all should be 0th percentile, got %f", p)
	}
}

func TestPercentile_AboveAll(t *testing.T) {
	allScores := []float64{0.1, 0.2, 0.3}
	p := ranking.Percentile(0.9, allScores)
	if p != 100.0 {
		t.Errorf("score above all should be 100th percentile, got %f", p)
	}
}

func TestPercentile_Empty(t *testing.T) {
	p := ranking.Percentile(0.5, nil)
	if p != 100.0 {
		t.Errorf("percentile with empty slice should be 100, got %f", p)
	}
}
