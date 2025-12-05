package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultQueryTimeout = 5 * time.Second

// Repository provides database access for authentication concerns.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a new Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateUser persists a new user record.
func (r *Repository) CreateUser(ctx context.Context, email, passwordHash string, displayName *string) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	query := `
INSERT INTO users (email, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING id, email, password_hash, display_name, is_admin, created_at, updated_at;`

	row := r.pool.QueryRow(ctx, query, email, passwordHash, displayName)

	var user User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrEmailAlreadyExists
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}

	return user, nil
}

// FindUserByEmail fetches a user by email.
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	query := `
SELECT id, email, password_hash, display_name, is_admin, created_at, updated_at
FROM users
WHERE email = $1;`

	var user User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("find user: %w", err)
	}

	return user, nil
}

// StoreRefreshToken saves or updates a refresh token hash for the user.
func (r *Repository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	query := `
INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked_at)
VALUES ($1, $2, $3, NULL)
ON CONFLICT (user_id, token_hash)
DO UPDATE SET expires_at = EXCLUDED.expires_at, revoked_at = NULL, created_at = NOW();`

	if _, err := r.pool.Exec(ctx, query, userID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// RevokeToken marks a refresh token as revoked.
func (r *Repository) RevokeToken(ctx context.Context, userID uuid.UUID, tokenHash string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	query := `
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND token_hash = $2;`

	if _, err := r.pool.Exec(ctx, query, userID, tokenHash); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	return nil
}
