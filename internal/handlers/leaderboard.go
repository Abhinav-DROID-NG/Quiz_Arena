package handlers

import (
	"net/http"
	"strconv"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// LeaderboardHandler handles leaderboard HTTP requests.
type LeaderboardHandler struct {
	lb *repository.LeaderboardRepo
}

// NewLeaderboardHandler creates a new LeaderboardHandler.
func NewLeaderboardHandler(lb *repository.LeaderboardRepo) *LeaderboardHandler {
	return &LeaderboardHandler{lb: lb}
}

// QuizLeaderboard handles GET /api/leaderboard/quiz/:quiz_id.
func (h *LeaderboardHandler) QuizLeaderboard(w http.ResponseWriter, r *http.Request) {
	quizID, err := uuid.Parse(chi.URLParam(r, "quiz_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid quiz id")
		return
	}

	entries, err := h.lb.GetQuizLeaderboard(r.Context(), quizID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch leaderboard")
		return
	}
	if entries == nil {
		entries = []*models.LeaderboardEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

// GlobalLeaderboard handles GET /api/leaderboard/global.
func (h *LeaderboardHandler) GlobalLeaderboard(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	entries, err := h.lb.GetGlobalLeaderboard(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch global leaderboard")
		return
	}
	if entries == nil {
		entries = []*models.GlobalLeaderboardEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}
