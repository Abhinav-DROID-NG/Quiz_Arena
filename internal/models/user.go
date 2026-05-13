package models

import (
	"time"

	"github.com/google/uuid"
)

// AuthProvider identifies how the user authenticated.
type AuthProvider string

const (
	AuthProviderLocal  AuthProvider = "local"
	AuthProviderGoogle AuthProvider = "google"
)

// User represents an application user with authentication details and Elo rating.
type User struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	Username     string       `json:"username" db:"username"`
	Email        string       `json:"email" db:"email"`
	PasswordHash string       `json:"-" db:"password_hash"`
	GoogleID     string       `json:"-" db:"google_id"`
	Provider     AuthProvider `json:"provider" db:"provider"`
	AvatarURL    string       `json:"avatar_url,omitempty" db:"avatar_url"`
	EloRating    int          `json:"elo_rating" db:"elo_rating"`
	IsAdmin      bool         `json:"is_admin" db:"is_admin"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
}

// RegisterRequest holds user registration input.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest holds user login input.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is returned after a successful login or register.
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
