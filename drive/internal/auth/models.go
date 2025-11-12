package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents an application user.
type User struct {
	ID           uuid.UUID
	Email        string
	DisplayName  *string
	IsAdmin      bool
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SafeUser removes sensitive fields for response payloads.
func (u User) SafeUser() User {
	u.PasswordHash = ""
	return u
}

// TokenPair bundles access and refresh tokens.
type TokenPair struct {
	AccessToken        string
	AccessTokenExpiry  time.Time
	RefreshToken       string
	RefreshTokenExpiry time.Time
}
