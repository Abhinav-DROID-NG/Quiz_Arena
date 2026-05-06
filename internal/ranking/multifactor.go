package ranking

import (
	"math"
	"sort"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
)

// weights for composite score computation.
const (
	weightScore    = 0.6
	weightTime     = 0.25
	weightAccuracy = 0.15
)

// ComputeComposite calculates the composite ranking score for a single attempt.
// timeLimitSeconds is the quiz's time limit used to normalise elapsed time.
func ComputeComposite(score, timeLimitSeconds, timeElapsed, totalQuestions int) float64 {
	if totalQuestions == 0 || timeLimitSeconds == 0 {
		return 0
	}

	// Normalised score: raw score / total questions (clamp to [0,1]).
	normScore := math.Max(0, math.Min(1, float64(score)/float64(totalQuestions)))

	// Time efficiency: (1 - fraction of time used), so faster answers rank higher.
	timeRatio := float64(timeElapsed) / float64(timeLimitSeconds)
	if timeRatio > 1 {
		timeRatio = 1
	}
	timeEfficiency := 1.0 - timeRatio

	// Accuracy: correct / total questions, cannot exceed 1.
	accuracy := normScore // same as norm score since score == correct count here

	return weightScore*normScore + weightTime*timeEfficiency + weightAccuracy*accuracy
}

// RankEntry holds the data needed to rank a single attempt.
type RankEntry struct {
	Entry          *models.LeaderboardEntry
	CompositeScore float64
}

// RankEntries assigns ranks to a slice of leaderboard entries sorted by
// composite score descending. Ties share the same rank.
func RankEntries(entries []*models.LeaderboardEntry) []*models.LeaderboardEntry {
	if len(entries) == 0 {
		return entries
	}

	// Sort by composite score descending.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CompositeScore > entries[j].CompositeScore
	})

	rank := 1
	for i, e := range entries {
		if i > 0 && entries[i].CompositeScore < entries[i-1].CompositeScore {
			rank = i + 1
		}
		e.Rank = rank
	}
	return entries
}

// Percentile returns the percentile rank of a score within a sorted slice of scores.
// scoreSlice must be sorted ascending.
func Percentile(score float64, allScores []float64) float64 {
	if len(allScores) == 0 {
		return 100.0
	}
	below := 0
	for _, s := range allScores {
		if s < score {
			below++
		}
	}
	return (float64(below) / float64(len(allScores))) * 100.0
}
