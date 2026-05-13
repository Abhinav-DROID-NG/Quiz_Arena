package quiz

import (
	"encoding/json"
	"net/http"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/auth"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for quiz session endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a quiz Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Start handles POST /api/sessions/start.
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.QuizID == uuid.Nil {
		utils.WriteError(w, http.StatusBadRequest, "quiz_id is required")
		return
	}

	resp, err := h.svc.StartSession(r.Context(), claims.UserID, req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// Answer handles POST /api/sessions/{id}/answer.
func (h *Handler) Answer(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	var req AnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.AnswerQuestion(r.Context(), claims.UserID, sessionID, req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Complete handles POST /api/sessions/{id}/complete.
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	resp, err := h.svc.CompleteSession(r.Context(), claims.UserID, sessionID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
