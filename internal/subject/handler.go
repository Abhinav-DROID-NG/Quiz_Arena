package subject

import (
	"encoding/json"
	"net/http"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/auth"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for subject endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new subject Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /api/subjects — returns all subjects.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.svc.List(r.Context())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch subjects")
		return
	}
	if subjects == nil {
		subjects = []*models.Subject{}
	}
	utils.WriteJSON(w, http.StatusOK, subjects)
}

// Get handles GET /api/subjects/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid subject id")
		return
	}
	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "subject not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, sub)
}

// Create handles POST /api/subjects — admin only.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil || !claims.IsAdmin {
		utils.WriteError(w, http.StatusForbidden, "admin access required")
		return
	}

	var req models.CreateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := h.svc.Create(r.Context(), req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, sub)
}

// SubjectLeaderboard handles GET /api/subjects/{id}/leaderboard.
func (h *Handler) SubjectLeaderboard(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid subject id")
		return
	}

	entries, err := h.svc.GetSubjectLeaderboard(r.Context(), id, 50)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch leaderboard")
		return
	}
	if entries == nil {
		entries = []*models.UserSubjectElo{}
	}
	utils.WriteJSON(w, http.StatusOK, entries)
}
