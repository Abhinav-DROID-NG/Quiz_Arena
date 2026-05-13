package question

import (
	"context"
	"fmt"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
)

// Service contains business logic for questions.
type Service struct {
	repo      *Repository
	shuffler  *Shuffler
	validator *Validator
}

// NewService creates a question Service.
func NewService(repo *Repository, shuffler *Shuffler, validator *Validator) *Service {
	return &Service{repo: repo, shuffler: shuffler, validator: validator}
}

// GetByIDForClient returns a client-safe question view with shuffled options.
// The shuffle mapping is also returned for server-side answer validation.
func (s *Service) GetByIDForClient(ctx context.Context, id uuid.UUID) (*models.QuestionForClient, map[int]string, error) {
	q, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	client, mapping := s.shuffler.Shuffle(q)
	return client, mapping, nil
}

// GetByID returns the full question including the answer (for server-side use only).
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Question, error) {
	return s.repo.GetByID(ctx, id)
}

// Create validates and inserts a new question.
func (s *Service) Create(ctx context.Context, q *models.Question) (*models.Question, error) {
	if errs := s.validator.Validate(q); len(errs) > 0 {
		return nil, fmt.Errorf("validation: %v", errs)
	}
	// Auto-detect LaTeX if not explicitly set.
	if !q.LatexEnabled {
		q.LatexEnabled = HasLatex(q.Text)
	}
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return s.repo.Create(ctx, q)
}

// GetByQuiz returns all questions for a quiz.
func (s *Service) GetByQuiz(ctx context.Context, quizID uuid.UUID) ([]*models.Question, error) {
	return s.repo.GetByQuiz(ctx, quizID)
}
