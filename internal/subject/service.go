package subject

import (
	"context"
	"fmt"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
)

// Service contains business logic for subjects.
type Service struct {
	repo *Repository
}

// NewService creates a subject Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns all subjects.
func (s *Service) List(ctx context.Context) ([]*models.Subject, error) {
	return s.repo.List(ctx)
}

// GetByID returns a single subject.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Subject, error) {
	return s.repo.GetByID(ctx, id)
}

// Create creates a new subject (admin operation).
func (s *Service) Create(ctx context.Context, req models.CreateSubjectRequest) (*models.Subject, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("subject name is required")
	}
	sub := &models.Subject{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
	}
	return s.repo.Create(ctx, sub)
}

// GetUserSubjectElo returns a user's ELO for a subject.
func (s *Service) GetUserSubjectElo(ctx context.Context, userID, subjectID uuid.UUID) (int, error) {
	return s.repo.GetUserSubjectElo(ctx, userID, subjectID)
}

// UpsertUserSubjectElo creates or updates a user's subject ELO.
func (s *Service) UpsertUserSubjectElo(ctx context.Context, userID, subjectID uuid.UUID, elo int) error {
	return s.repo.UpsertUserSubjectElo(ctx, userID, subjectID, elo)
}

// GetSubjectLeaderboard returns the ranked leaderboard for a subject.
func (s *Service) GetSubjectLeaderboard(ctx context.Context, subjectID uuid.UUID, limit int) ([]*models.UserSubjectElo, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetSubjectLeaderboard(ctx, subjectID, limit)
}
