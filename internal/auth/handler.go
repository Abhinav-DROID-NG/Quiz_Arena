package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/utils"
)

// Handler handles HTTP requests for authentication endpoints.
type Handler struct {
	svc    *Service
	google *GoogleProvider
}

// NewHandler creates an auth Handler.
func NewHandler(svc *Service, google *GoogleProvider) *Handler {
	return &Handler{svc: svc, google: google}
}

// Register handles POST /auth/register — local email/password registration.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "username, email and password are required")
		return
	}

	resp, err := h.svc.RegisterLocal(r.Context(), req)
	if err != nil {
		utils.WriteError(w, http.StatusConflict, "email or username already taken")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// Login handles POST /auth/login — local email/password login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.LoginLocal(r.Context(), req)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GoogleLogin handles GET /auth/google — redirects the user to Google's OAuth consent screen.
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		utils.WriteError(w, http.StatusNotImplemented, "Google OAuth is not configured")
		return
	}

	state, err := generateState()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not generate state")
		return
	}

	// Store state in a short-lived cookie for CSRF validation.
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		Path:     "/",
	})

	http.Redirect(w, r, h.google.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// GoogleCallback handles GET /auth/google/callback — processes the OAuth2 callback.
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		utils.WriteError(w, http.StatusNotImplemented, "Google OAuth is not configured")
		return
	}

	// Validate state to prevent CSRF.
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		utils.WriteError(w, http.StatusBadRequest, "invalid OAuth state")
		return
	}

	// Clear the state cookie.
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	resp, err := h.svc.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "OAuth authentication failed")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// generateState creates a cryptographically random state string for CSRF protection.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
