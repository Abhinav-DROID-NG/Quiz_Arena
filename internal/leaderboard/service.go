package leaderboard

import (
	"context"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
)

// Service handles leaderboard business logic.
type Service struct {
	repo *Repository
}

// NewService creates a leaderboard Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetQuizLeaderboard returns the ranked entries for a quiz.
func (s *Service) GetQuizLeaderboard(ctx context.Context, quizID uuid.UUID) ([]*models.LeaderboardEntry, error) {
	return s.repo.GetQuizLeaderboard(ctx, quizID)
}

// GetGlobalLeaderboard returns top users by global ELO.
func (s *Service) GetGlobalLeaderboard(ctx context.Context, limit int) ([]*models.GlobalLeaderboardEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.GetGlobalLeaderboard(ctx, limit)
}

// GetSubjectLeaderboard returns top users for a subject.
func (s *Service) GetSubjectLeaderboard(ctx context.Context, subjectID uuid.UUID, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.GetSubjectLeaderboard(ctx, subjectID, limit)
}
