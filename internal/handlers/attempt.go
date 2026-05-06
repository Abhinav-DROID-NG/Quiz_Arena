package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/engine"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/middleware"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/ranking"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AttemptHandler handles quiz session HTTP requests.
type AttemptHandler struct {
	attempts  *repository.AttemptRepo
	quizzes   *repository.QuizRepo
	users     *repository.UserRepo
	pool      *pgxpool.Pool
}

// NewAttemptHandler creates a new AttemptHandler.
func NewAttemptHandler(
	attempts *repository.AttemptRepo,
	quizzes *repository.QuizRepo,
	users *repository.UserRepo,
	pool *pgxpool.Pool,
) *AttemptHandler {
	return &AttemptHandler{
		attempts: attempts,
		quizzes:  quizzes,
		users:    users,
		pool:     pool,
	}
}

// Start handles POST /api/sessions/start — starts a new quiz session.
func (h *AttemptHandler) Start(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req models.StartSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Verify quiz exists and is published.
	quiz, err := h.quizzes.GetByID(r.Context(), req.QuizID)
	if err != nil || !quiz.IsPublished {
		writeError(w, http.StatusNotFound, "quiz not found")
		return
	}

	// Fetch the user to get current Elo rating.
	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	// Select questions based on user Elo.
	difficulty := engine.SelectDifficulty(user.EloRating)
	questions, err := h.quizzes.GetQuestionsByDifficulty(r.Context(), req.QuizID, difficulty)
	if err != nil || len(questions) == 0 {
		// Fall back to all questions if none found for this difficulty tier.
		questions, err = h.quizzes.GetQuestionsByQuiz(r.Context(), req.QuizID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load questions")
			return
		}
	}

	attempt := &models.Attempt{
		ID:     uuid.New(),
		UserID: claims.UserID,
		QuizID: req.QuizID,
		Status: models.AttemptStatusActive,
	}

	created, err := h.attempts.Create(r.Context(), attempt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start session")
		return
	}

	clientQuestions := make([]*models.QuestionForClient, len(questions))
	for i, q := range questions {
		clientQuestions[i] = q.ToClient()
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"session_id": created.ID,
		"quiz":       quiz,
		"questions":  clientQuestions,
		"started_at": created.StartedAt,
	})
}

// Answer handles POST /api/sessions/:id/answer — submits a single answer.
func (h *AttemptHandler) Answer(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	attemptID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	attempt, err := h.attempts.GetByID(r.Context(), attemptID)
	if err != nil || attempt.UserID != claims.UserID {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if attempt.Status != models.AttemptStatusActive {
		writeError(w, http.StatusConflict, "session is not active")
		return
	}

	var req models.AnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	question, err := h.quizzes.GetQuestionByID(r.Context(), req.QuestionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "question not found")
		return
	}

	isCorrect := req.SelectedOpt != "" && req.SelectedOpt == question.Answer

	ans := &models.Answer{
		ID:          uuid.New(),
		AttemptID:   attemptID,
		QuestionID:  req.QuestionID,
		SelectedOpt: req.SelectedOpt,
		IsCorrect:   isCorrect,
		AnsweredAt:  time.Now(),
	}

	if err := h.attempts.SaveAnswer(r.Context(), ans); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save answer")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"recorded": true,
	})
}

// Complete handles POST /api/sessions/:id/complete — finalises a quiz session.
func (h *AttemptHandler) Complete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	attemptID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	attempt, err := h.attempts.GetByID(r.Context(), attemptID)
	if err != nil || attempt.UserID != claims.UserID {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if attempt.Status != models.AttemptStatusActive {
		writeError(w, http.StatusConflict, "session already completed")
		return
	}

	// Load all answers.
	answers, err := h.attempts.GetAnswers(r.Context(), attemptID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load answers")
		return
	}

	// Load questions to count total and get difficulty info.
	questions, err := h.quizzes.GetQuestionsByQuiz(r.Context(), attempt.QuizID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load questions")
		return
	}

	// Build answer lookup.
	answerMap := make(map[uuid.UUID]*models.Answer, len(answers))
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}

	// Compute score.
	var answerResults []engine.AnswerResult
	var eloInputs []struct {
		Difficulty models.Difficulty
		Correct    bool
	}
	for _, q := range questions {
		ans, answered := answerMap[q.ID]
		result := engine.AnswerResult{
			Correct: answered && ans.IsCorrect,
			Skipped: !answered || ans.SelectedOpt == "",
		}
		answerResults = append(answerResults, result)
		eloInputs = append(eloInputs, struct {
			Difficulty models.Difficulty
			Correct    bool
		}{q.Difficulty, result.Correct})
	}

	scoreResult := engine.ComputeScore(answerResults)
	timeElapsed := int(time.Since(attempt.StartedAt).Seconds())

	// Compute Elo delta.
	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	eloUpdate := engine.ApplyEloToAttempt(user.EloRating, eloInputs)

	// Persist completion.
	if err := h.attempts.Complete(r.Context(), attemptID,
		scoreResult.Score, scoreResult.RawScore,
		scoreResult.CorrectCount, scoreResult.WrongCount, scoreResult.UnansweredCount,
		timeElapsed, eloUpdate.Delta,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to complete session")
		return
	}

	// Update user Elo.
	if err := h.users.UpdateElo(r.Context(), claims.UserID, eloUpdate.NewRating); err != nil {
		// Non-fatal: log but don't fail the request.
		_ = err
	}

	// Reload completed attempt.
	completed, err := h.attempts.GetByID(r.Context(), attemptID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reload attempt")
		return
	}

	// Asynchronously refresh leaderboard.
	quiz, _ := h.quizzes.GetByID(r.Context(), attempt.QuizID)
	if quiz != nil {
		go func() {
			ctx := context.Background()
			_ = ranking.RefreshLeaderboard(ctx, h.pool, attempt.QuizID.String(), quiz.TimeLimit)
		}()
	}

	writeJSON(w, http.StatusOK, models.CompleteSessionResponse{
		Attempt:    completed,
		Rank:       0, // will be populated after async leaderboard refresh
		Percentile: 0,
	})
}
