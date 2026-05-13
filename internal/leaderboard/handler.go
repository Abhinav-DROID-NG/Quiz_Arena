package leaderboard

import (
	"net/http"
	"strconv"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for leaderboard endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a leaderboard Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// QuizLeaderboard handles GET /api/leaderboard/quiz/{quiz_id}.
func (h *Handler) QuizLeaderboard(w http.ResponseWriter, r *http.Request) {
	quizID, err := uuid.Parse(chi.URLParam(r, "quiz_id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid quiz id")
		return
	}

	entries, err := h.svc.GetQuizLeaderboard(r.Context(), quizID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch leaderboard")
		return
	}
	if entries == nil {
		entries = []*models.LeaderboardEntry{}
	}
	utils.WriteJSON(w, http.StatusOK, entries)
}

// GlobalLeaderboard handles GET /api/leaderboard/global.
func (h *Handler) GlobalLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	entries, err := h.svc.GetGlobalLeaderboard(r.Context(), limit)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch global leaderboard")
		return
	}
	if entries == nil {
		entries = []*models.GlobalLeaderboardEntry{}
	}
	utils.WriteJSON(w, http.StatusOK, entries)
}

// SubjectLeaderboard handles GET /api/leaderboard/subject/{subject_id}.
func (h *Handler) SubjectLeaderboard(w http.ResponseWriter, r *http.Request) {
	subjectID, err := uuid.Parse(chi.URLParam(r, "subject_id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid subject id")
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	entries, err := h.svc.GetSubjectLeaderboard(r.Context(), subjectID, limit)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch subject leaderboard")
		return
	}
	if entries == nil {
		entries = []map[string]any{}
	}
	utils.WriteJSON(w, http.StatusOK, entries)
}
