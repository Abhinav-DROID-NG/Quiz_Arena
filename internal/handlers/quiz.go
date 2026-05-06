package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/middleware"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// QuizHandler handles quiz-related HTTP requests.
type QuizHandler struct {
	quizzes *repository.QuizRepo
}

// NewQuizHandler creates a new QuizHandler.
func NewQuizHandler(quizzes *repository.QuizRepo) *QuizHandler {
	return &QuizHandler{quizzes: quizzes}
}

// List handles GET /api/quizzes — returns all published quizzes.
func (h *QuizHandler) List(w http.ResponseWriter, r *http.Request) {
	quizzes, err := h.quizzes.ListPublished(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch quizzes")
		return
	}
	if quizzes == nil {
		quizzes = []*models.Quiz{}
	}
	writeJSON(w, http.StatusOK, quizzes)
}

// Get handles GET /api/quizzes/:id — returns a quiz with its questions (without answers).
func (h *QuizHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid quiz id")
		return
	}

	quiz, err := h.quizzes.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "quiz not found")
		return
	}

	questions, err := h.quizzes.GetQuestionsByQuiz(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch questions")
		return
	}

	clientQuestions := make([]*models.QuestionForClient, len(questions))
	for i, q := range questions {
		clientQuestions[i] = q.ToClient()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"quiz":      quiz,
		"questions": clientQuestions,
	})
}

// Create handles POST /api/quizzes — admin only, creates a new quiz.
func (h *QuizHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req models.CreateQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.TimeLimit < 60 {
		writeError(w, http.StatusBadRequest, "time limit must be at least 60 seconds")
		return
	}

	quiz := &models.Quiz{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   claims.UserID,
		TimeLimit:   req.TimeLimit,
		IsPublished: true,
	}

	created, err := h.quizzes.Create(r.Context(), quiz)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create quiz")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}
