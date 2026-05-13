package quiz

import (
	"context"
	"fmt"
	"time"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/elo"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/question"
	"github.com/google/uuid"
)

// Service contains the quiz session business logic.
type Service struct {
	repo     *Repository
	selector *question.Selector
	shuffler *question.Shuffler
	timer    *Timer
}

// NewService creates a quiz Service.
func NewService(repo *Repository, selector *question.Selector, shuffler *question.Shuffler, timer *Timer) *Service {
	return &Service{
		repo:     repo,
		selector: selector,
		shuffler: shuffler,
		timer:    timer,
	}
}

// StartSession creates a new quiz session and returns the first question.
func (s *Service) StartSession(ctx context.Context, userID uuid.UUID, req StartRequest) (*StartResponse, error) {
	quiz, err := s.repo.GetQuizByID(ctx, req.QuizID)
	if err != nil || !quiz.IsPublished {
		return nil, fmt.Errorf("quiz not found or not published")
	}

	userElo, err := s.repo.GetUserElo(ctx, userID)
	if err != nil {
		userElo = elo.DefaultRating
	}

	// If a subject is specified and a subject-specific ELO exists, use that.
	if req.SubjectID != nil {
		subjectElo, err := s.repo.GetSubjectElo(ctx, userID, *req.SubjectID)
		if err == nil {
			userElo = subjectElo
		}
	}

	// Fetch previously seen question IDs to avoid repetition.
	seenIDs := make(map[uuid.UUID]bool)
	if qrepo, ok := s.selector.GetRepo(); ok {
		seenIDs, _ = qrepo.GetSeenIDs(ctx, userID, req.QuizID)
	}

	cfg := question.DefaultConfig(userElo)
	questions, err := s.selector.Select(ctx, req.QuizID, cfg, seenIDs)
	if err != nil || len(questions) == 0 {
		return nil, fmt.Errorf("no questions available for this quiz")
	}

	// Build shuffle maps for every question up-front (server-side only).
	shuffleMaps := make([]map[int]string, len(questions))
	for i, q := range questions {
		_, mapping := s.shuffler.Shuffle(q)
		shuffleMaps[i] = mapping
	}

	// Create attempt record.
	attempt := &models.Attempt{
		ID:     uuid.New(),
		UserID: userID,
		QuizID: req.QuizID,
		Status: models.AttemptStatusActive,
	}
	created, err := s.repo.CreateAttempt(ctx, attempt)
	if err != nil {
		return nil, fmt.Errorf("create attempt: %w", err)
	}

	// Store session state in memory.
	session := &SessionState{
		AttemptID:   created.ID,
		UserID:      userID,
		QuizID:      req.QuizID,
		SubjectID:   req.SubjectID,
		Questions:   questions,
		ShuffleMaps: shuffleMaps,
		CurrentIdx:  0,
		StartedAt:   created.StartedAt,
		QuestionAt:  time.Now(),
		TimeLimitS:  quiz.TimeLimit,
		Answers:     make(map[uuid.UUID]*answerEntry),
	}
	s.repo.StoreSession(session)

	// Return first question (client-safe).
	firstQ, _ := s.shuffler.Shuffle(questions[0])

	return &StartResponse{
		SessionID:  created.ID,
		QuizTitle:  quiz.Title,
		TotalCount: len(questions),
		CurrentIdx: 0,
		Question:   firstQ,
		TimeLimitS: quiz.TimeLimit,
		StartedAt:  created.StartedAt.Format(time.RFC3339),
	}, nil
}

// AnswerQuestion records the user's answer for the current question and advances the session.
func (s *Service) AnswerQuestion(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, req AnswerRequest) (*AnswerResponse, error) {
	session, ok := s.repo.GetSession(sessionID)
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	if !session.IsOwner(userID) {
		return nil, fmt.Errorf("session belongs to another user")
	}
	if session.Completed {
		return nil, fmt.Errorf("session is already completed")
	}

	// Enforce total session timer.
	if session.TimerExpired() {
		_ = s.autoComplete(ctx, session)
		return nil, fmt.Errorf("session time expired")
	}

	currentQ := session.CurrentQuestion()
	if currentQ == nil {
		return nil, fmt.Errorf("no current question")
	}

	// Determine server-side response time (ignore client-reported time for ELO purposes).
	serverResponseMS := s.timer.ResponseTimeMS(session.QuestionAt)

	// Resolve the original option label from the shuffled index.
	var originalOpt string
	var isCorrect bool

	if req.OptionIndex >= 0 && req.OptionIndex <= 3 {
		shuffleMap := session.ShuffleMaps[session.CurrentIdx]
		originalOpt = shuffleMap[req.OptionIndex]
		isCorrect = originalOpt == currentQ.Answer
	} else {
		// Skipped
		originalOpt = ""
		isCorrect = false
	}

	// Store answer in session memory.
	session.Answers[currentQ.ID] = &answerEntry{
		SelectedIdx:  req.OptionIndex,
		OriginalOpt:  originalOpt,
		IsCorrect:    isCorrect,
		ResponseTime: serverResponseMS,
	}

	// Persist to database.
	ans := &models.UserAnswer{
		ID:           uuid.New(),
		AttemptID:    sessionID,
		QuestionID:   currentQ.ID,
		SelectedOpt:  originalOpt,
		IsCorrect:    isCorrect,
		ResponseTime: serverResponseMS,
		AnsweredAt:   time.Now(),
	}
	if err := s.repo.SaveAnswer(ctx, ans); err != nil {
		return nil, fmt.Errorf("save answer: %w", err)
	}

	// Advance to next question.
	hasNext := session.HasNext()
	var nextQ *models.QuestionForClient
	if hasNext {
		session.Advance()
		nextQ, _ = s.shuffler.Shuffle(session.Questions[session.CurrentIdx])
	}

	return &AnswerResponse{
		Recorded:     true,
		HasNext:      hasNext,
		NextQuestion: nextQ,
		CurrentIdx:   session.CurrentIdx,
	}, nil
}

// CompleteSession finalises the session, computes ELO, and updates the leaderboard.
func (s *Service) CompleteSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) (*CompleteResponse, error) {
	session, ok := s.repo.GetSession(sessionID)
	if !ok {
		// Fall back to DB-based completion.
		return s.completeFromDB(ctx, userID, sessionID)
	}
	if !session.IsOwner(userID) {
		return nil, fmt.Errorf("session belongs to another user")
	}
	if session.Completed {
		return nil, fmt.Errorf("session is already completed")
	}

	return s.finalise(ctx, session)
}

// finalise computes the score, updates ELO, persists completion, and cleans up.
func (s *Service) finalise(ctx context.Context, session *SessionState) (*CompleteResponse, error) {
	// Build ELO input from session answers.
	results := make([]elo.QuestionResult, len(session.Questions))
	for i, q := range session.Questions {
		entry, answered := session.Answers[q.ID]
		var result elo.QuestionResult
		result.Difficulty = q.Difficulty
		if answered {
			result.Correct = entry.IsCorrect
			result.Skipped = entry.OriginalOpt == ""
			result.ResponseTime = entry.ResponseTime
		} else {
			result.Skipped = true
		}
		results[i] = result
	}

	// Fetch current user ELO.
	currentElo, err := s.repo.GetUserElo(ctx, session.UserID)
	if err != nil {
		currentElo = elo.DefaultRating
	}

	eloUpdate, score := elo.ApplyToAttempt(currentElo, results)
	timeElapsed := int(time.Since(session.StartedAt).Seconds())

	// Persist completion.
	if err := s.repo.CompleteAttempt(ctx, session.AttemptID,
		score.Score, score.RawScore,
		score.CorrectCount, score.WrongCount, score.UnansweredCount,
		timeElapsed, eloUpdate.Delta,
	); err != nil {
		return nil, fmt.Errorf("complete attempt: %w", err)
	}

	// Update global ELO.
	_ = s.repo.UpdateUserElo(ctx, session.UserID, eloUpdate.NewRating)

	// Update per-subject ELO if the session is subject-specific.
	if session.SubjectID != nil {
		subjectElo, _ := s.repo.GetSubjectElo(ctx, session.UserID, *session.SubjectID)
		subjectUpdate := elo.ComputeQuestion(subjectElo, elo.SelectDifficulty(currentElo), float64(score.Score)/float64(len(session.Questions)+1))
		_ = s.repo.UpsertSubjectElo(ctx, session.UserID, *session.SubjectID, subjectUpdate.NewRating)
	}

	// Mark session as done and clean up.
	session.Completed = true
	s.repo.DeleteSession(session.AttemptID)

	// Asynchronously refresh the leaderboard.
	go func() {
		_ = s.repo.RefreshLeaderboard(context.Background(), session.QuizID, session.TimeLimitS)
	}()

	// Reload attempt from DB for the response.
	attempt, _ := s.repo.GetAttemptByID(ctx, session.AttemptID)

	return &CompleteResponse{
		Attempt:  attempt,
		EloDelta: eloUpdate.Delta,
		NewElo:   eloUpdate.NewRating,
	}, nil
}

// autoComplete handles sessions that exceeded the timer.
func (s *Service) autoComplete(ctx context.Context, session *SessionState) error {
	_, err := s.finalise(ctx, session)
	return err
}

// completeFromDB handles completion when the in-memory session is gone (e.g., server restart).
func (s *Service) completeFromDB(ctx context.Context, userID, sessionID uuid.UUID) (*CompleteResponse, error) {
	attempt, err := s.repo.GetAttemptByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}
	if attempt.UserID != userID {
		return nil, fmt.Errorf("session belongs to another user")
	}
	if attempt.Status == models.AttemptStatusCompleted {
		return &CompleteResponse{Attempt: attempt}, nil
	}
	// Mark as abandoned if not already complete.
	return &CompleteResponse{Attempt: attempt}, nil
}
