package question

import (
	"context"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/elo"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/utils"
	"github.com/google/uuid"
)

// Selector selects a smart question pool for a quiz session.
// It avoids recently seen questions and varies difficulty based on user ELO.
type Selector struct {
	repo *Repository
}

// NewSelector creates a Selector.
func NewSelector(repo *Repository) *Selector {
	return &Selector{repo: repo}
}

// SelectionConfig configures the question selection strategy.
type SelectionConfig struct {
	// Total number of questions to select.
	Count int
	// EasyFraction, MediumFraction, HardFraction must sum to 1.0 and represent
	// the target proportion of questions at each difficulty.
	EasyFraction   float64
	MediumFraction float64
	HardFraction   float64
}

// DefaultConfig returns a balanced config suitable for a 10-question session.
func DefaultConfig(userElo int) SelectionConfig {
	d := elo.SelectDifficulty(userElo)
	switch d {
	case models.DifficultyHard:
		return SelectionConfig{Count: 10, EasyFraction: 0.2, MediumFraction: 0.3, HardFraction: 0.5}
	case models.DifficultyMedium:
		return SelectionConfig{Count: 10, EasyFraction: 0.3, MediumFraction: 0.4, HardFraction: 0.3}
	default:
		return SelectionConfig{Count: 10, EasyFraction: 0.5, MediumFraction: 0.3, HardFraction: 0.2}
	}
}

// Select chooses a question set for a quiz session.
// It respects the difficulty distribution in cfg and avoids questions the user
// has already answered in previous sessions (seenIDs).
func (s *Selector) Select(ctx context.Context, quizID uuid.UUID, cfg SelectionConfig, seenIDs map[uuid.UUID]bool) ([]*models.Question, error) {
	// Load all questions for the quiz grouped by difficulty.
	allQuestions, err := s.repo.GetByQuiz(ctx, quizID)
	if err != nil {
		return nil, err
	}

	easy := filterDifficulty(allQuestions, models.DifficultyEasy, seenIDs)
	medium := filterDifficulty(allQuestions, models.DifficultyMedium, seenIDs)
	hard := filterDifficulty(allQuestions, models.DifficultyHard, seenIDs)

	nEasy := int(float64(cfg.Count) * cfg.EasyFraction)
	nMedium := int(float64(cfg.Count) * cfg.MediumFraction)
	nHard := cfg.Count - nEasy - nMedium

	selected := make([]*models.Question, 0, cfg.Count)
	selected = append(selected, utils.Sample(easy, nEasy)...)
	selected = append(selected, utils.Sample(medium, nMedium)...)
	selected = append(selected, utils.Sample(hard, nHard)...)

	// If we didn't get enough questions (small corpus), fill with unseen questions regardless of difficulty.
	if len(selected) < cfg.Count {
		remaining := filterSeen(allQuestions, seenIDs)
		extras := utils.Sample(remaining, cfg.Count-len(selected))
		selected = append(selected, extras...)
	}

	// Final fallback: if still short, allow repeats.
	if len(selected) < cfg.Count && len(allQuestions) > 0 {
		extras := utils.Sample(allQuestions, cfg.Count-len(selected))
		selected = append(selected, extras...)
	}

	// Shuffle the final set so difficulty order is not predictable.
	shuffled, _ := utils.Shuffle(selected)
	return shuffled, nil
}

// GetRepo returns the underlying repository for direct use.
func (s *Selector) GetRepo() (*Repository, bool) {
	return s.repo, s.repo != nil
}
func filterDifficulty(questions []*models.Question, d models.Difficulty, seen map[uuid.UUID]bool) []*models.Question {
	result := make([]*models.Question, 0)
	for _, q := range questions {
		if q.Difficulty == d && !seen[q.ID] {
			result = append(result, q)
		}
	}
	return result
}

// filterSeen returns questions the user hasn't seen regardless of difficulty.
func filterSeen(questions []*models.Question, seen map[uuid.UUID]bool) []*models.Question {
	result := make([]*models.Question, 0)
	for _, q := range questions {
		if !seen[q.ID] {
			result = append(result, q)
		}
	}
	return result
}
