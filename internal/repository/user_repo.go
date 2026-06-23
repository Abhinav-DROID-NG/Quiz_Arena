package repository

import (
	"context"
	"fmt"

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepo handles database operations for users.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create inserts a new user and returns the created record.
func (r *UserRepo) Create(ctx context.Context, u *models.User) (*models.User, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, username, email, password_hash, google_id, provider, avatar_url, elo_rating, is_admin)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, username, email, password_hash, google_id, provider, avatar_url, elo_rating, is_admin, created_at, updated_at
	`, u.ID, u.Username, u.Email, u.PasswordHash, u.GoogleID, u.Provider, u.AvatarURL, u.EloRating, u.IsAdmin,
	).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.GoogleID, &u.Provider, &u.AvatarURL,
		&u.EloRating, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetByEmail looks up a user by email address.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, google_id, provider, avatar_url, elo_rating, is_admin, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.GoogleID, &u.Provider, &u.AvatarURL,
		&u.EloRating, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// GetByID looks up a user by primary key.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, google_id, provider, avatar_url, elo_rating, is_admin, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.GoogleID, &u.Provider, &u.AvatarURL,
		&u.EloRating, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// GetByGoogleID looks up a user by their Google OAuth ID.
func (r *UserRepo) GetByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, google_id, provider, avatar_url, elo_rating, is_admin, created_at, updated_at
		FROM users WHERE google_id = $1
	`, googleID).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.GoogleID, &u.Provider, &u.AvatarURL,
		&u.EloRating, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by google id: %w", err)
	}
	return u, nil
}

// UpdateGoogleID links a Google ID to an existing user account.
func (r *UserRepo) UpdateGoogleID(ctx context.Context, id uuid.UUID, googleID, avatarURL string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET google_id = $1, avatar_url = $2, provider = $3, updated_at = NOW() WHERE id = $4",
		googleID, avatarURL, models.AuthProviderGoogle, id,
	)
	if err != nil {
		return fmt.Errorf("update google id: %w", err)
	}
	return nil
}

// UpdateElo updates the elo_rating for a user.
func (r *UserRepo) UpdateElo(ctx context.Context, id uuid.UUID, newRating int) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET elo_rating = $1, updated_at = NOW() WHERE id = $2",
		newRating, id,
	)
	if err != nil {
		return fmt.Errorf("update elo: %w", err)
	}
	return nil
}
