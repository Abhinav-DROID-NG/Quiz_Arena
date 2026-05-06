package handlers

import (
	"net/http"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/middleware"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/repository"
)

// AnalyticsHandler handles analytics HTTP requests.
type AnalyticsHandler struct {
	attempts *repository.AttemptRepo
	users    *repository.UserRepo
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(attempts *repository.AttemptRepo, users *repository.UserRepo) *AnalyticsHandler {
	return &AnalyticsHandler{attempts: attempts, users: users}
}

// Performance handles GET /api/analytics/performance.
// Returns a summary of the authenticated user's quiz history.
func (h *AnalyticsHandler) Performance(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	history, err := h.attempts.GetUserHistory(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch history")
		return
	}

	// Compute aggregate metrics.
	var totalScore, totalTime, totalCorrect, totalWrong, totalUnanswered int
	for _, a := range history {
		totalScore += a.Score
		totalTime += a.TimeElapsed
		totalCorrect += a.CorrectAnswers
		totalWrong += a.WrongAnswers
		totalUnanswered += a.Unanswered
	}

	total := totalCorrect + totalWrong + totalUnanswered
	accuracy := 0.0
	if total > 0 {
		accuracy = float64(totalCorrect) / float64(total) * 100
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"attempts":          len(history),
		"total_score":       totalScore,
		"average_score":     safeDiv(totalScore, len(history)),
		"total_time":        totalTime,
		"correct_answers":   totalCorrect,
		"wrong_answers":     totalWrong,
		"unanswered":        totalUnanswered,
		"accuracy_percent":  accuracy,
		"recent_attempts":   history,
	})
}

// EloProgression handles GET /api/analytics/elo-progression.
// Returns the user's Elo rating history derived from attempt elo_delta values.
func (h *AnalyticsHandler) EloProgression(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	history, err := h.attempts.GetUserHistory(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch history")
		return
	}

	// Reconstruct Elo over time by working backwards from current rating.
	type point struct {
		AttemptID   string  `json:"attempt_id"`
		CompletedAt any     `json:"completed_at"`
		EloDelta    int     `json:"elo_delta"`
		EloAfter    int     `json:"elo_after"`
	}

	currentElo := user.EloRating
	points := make([]point, len(history))
	for i, a := range history {
		points[i] = point{
			AttemptID:   a.ID.String(),
			CompletedAt: a.CompletedAt,
			EloDelta:    a.EloDelta,
			EloAfter:    currentElo,
		}
		currentElo -= a.EloDelta
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"current_elo": user.EloRating,
		"progression": points,
	})
}

func safeDiv(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
