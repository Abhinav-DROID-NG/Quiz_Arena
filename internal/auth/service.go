package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
)

// UserRepository is the minimal interface the auth service needs.
type UserRepository interface {
	Create(ctx context.Context, u *models.User) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*models.User, error)
	UpdateGoogleID(ctx context.Context, userID uuid.UUID, googleID, avatarURL string) error
}

// Service handles authentication business logic.
type Service struct {
	users  UserRepository
	tokens *TokenManager
	google *GoogleProvider
}

// NewService creates an auth Service.
func NewService(users UserRepository, tokens *TokenManager, google *GoogleProvider) *Service {
	return &Service{users: users, tokens: tokens, google: google}
}

// RegisterLocal creates a new user with email/password credentials.
func (s *Service) RegisterLocal(ctx context.Context, req models.RegisterRequest) (*models.AuthResponse, error) {
	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Provider:     models.AuthProviderLocal,
		EloRating:    1000,
	}

	created, err := s.users.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.tokens.Generate(created.ID, created.Email, created.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &models.AuthResponse{Token: token, User: created}, nil
}

// LoginLocal authenticates with email/password and returns a JWT.
func (s *Service) LoginLocal(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !CheckPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.tokens.Generate(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &models.AuthResponse{Token: token, User: user}, nil
}

// HandleGoogleCallback processes the Google OAuth callback, creating or updating the user.
func (s *Service) HandleGoogleCallback(ctx context.Context, code string) (*models.AuthResponse, error) {
	if s.google == nil {
		return nil, fmt.Errorf("Google OAuth is not configured")
	}

	token, err := s.google.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("oauth exchange: %w", err)
	}

	info, err := s.google.UserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("fetch user info: %w", err)
	}

	// Try to find user by Google ID first.
	user, err := s.users.GetByGoogleID(ctx, info.ID)
	if err != nil {
		// Not found by Google ID – look up by email.
		user, err = s.users.GetByEmail(ctx, info.Email)
		if err != nil {
			// Create a brand-new user.
			username := deriveUsername(info.Name, info.Email)
			user = &models.User{
				ID:        uuid.New(),
				Username:  username,
				Email:     info.Email,
				GoogleID:  info.ID,
				Provider:  models.AuthProviderGoogle,
				AvatarURL: info.Picture,
				EloRating: 1000,
			}
			user, err = s.users.Create(ctx, user)
			if err != nil {
				return nil, fmt.Errorf("create google user: %w", err)
			}
		} else {
			// Existing local user – link Google ID.
			if err := s.users.UpdateGoogleID(ctx, user.ID, info.ID, info.Picture); err != nil {
				return nil, fmt.Errorf("link google id: %w", err)
			}
			user.GoogleID = info.ID
			user.AvatarURL = info.Picture
		}
	}

	jwtToken, err := s.tokens.Generate(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &models.AuthResponse{Token: jwtToken, User: user}, nil
}

// deriveUsername creates a username from the Google profile name or email.
func deriveUsername(name, email string) string {
	if name != "" {
		// Replace spaces with underscores, keep alphanumeric and underscores.
		clean := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
				return r
			}
			if r == ' ' {
				return '_'
			}
			return -1
		}, name)
		if len(clean) >= 3 {
			if len(clean) > 50 {
				return clean[:50]
			}
			return clean
		}
	}
	// Fall back to the part before @ in the email.
	parts := strings.SplitN(email, "@", 2)
	if len(parts[0]) >= 3 {
		if len(parts[0]) > 50 {
			return parts[0][:50]
		}
		return parts[0]
	}
	return "user_" + uuid.New().String()[:8]
}
