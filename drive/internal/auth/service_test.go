package auth

import (
	"context"
	"testing"
	"time"

	"github.com/abduss/godrive/internal/config"
	"github.com/google/uuid"
)

func TestRegisterSuccess(t *testing.T) {
	store := newMemoryStore()
	cfg := config.AuthConfig{
		AccessTokenSecret:  "access-secret",
		RefreshTokenSecret: "refresh-secret",
		AccessTokenTTL:     time.Minute,
		RefreshTokenTTL:    time.Hour,
		BcryptCost:         4,
	}

	service := NewService(store, cfg)
	result, err := service.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "StrongPass1!",
	})

	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}

	if result.User.PasswordHash != "" {
		t.Fatalf("expected password hash to be stripped from response")
	}

	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
		t.Fatalf("expected tokens to be issued")
	}

	if len(store.users) != 1 {
		t.Fatalf("expected user stored; got %d", len(store.users))
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	store := newMemoryStore()
	cfg := config.AuthConfig{
		AccessTokenSecret:  "access-secret",
		RefreshTokenSecret: "refresh-secret",
		AccessTokenTTL:     time.Minute,
		RefreshTokenTTL:    time.Hour,
		BcryptCost:         4,
	}

	service := NewService(store, cfg)
	_, err := service.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "StrongPass1!",
	})
	if err != nil {
		t.Fatalf("initial registration returned error: %v", err)
	}

	_, err = service.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "AnotherPass2!",
	})

	if err == nil || err != ErrEmailAlreadyExists {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestLogin(t *testing.T) {
	store := newMemoryStore()
	cfg := config.AuthConfig{
		AccessTokenSecret:  "access-secret",
		RefreshTokenSecret: "refresh-secret",
		AccessTokenTTL:     time.Minute,
		RefreshTokenTTL:    time.Hour,
		BcryptCost:         4,
	}

	service := NewService(store, cfg)
	_, err := service.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "StrongPass1!",
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}

	result, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "StrongPass1!",
	})

	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	if result.Tokens.AccessToken == "" {
		t.Fatalf("expected access token")
	}
	if result.Tokens.RefreshToken == "" {
		t.Fatalf("expected refresh token")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	store := newMemoryStore()
	cfg := config.AuthConfig{
		AccessTokenSecret:  "access-secret",
		RefreshTokenSecret: "refresh-secret",
		AccessTokenTTL:     time.Minute,
		RefreshTokenTTL:    time.Hour,
		BcryptCost:         4,
	}

	service := NewService(store, cfg)
	_, err := service.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "StrongPass1!",
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}

	_, err = service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "WrongPass",
	})

	if err == nil || err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

// memoryStore implements userStore for tests.
type memoryStore struct {
	users         map[string]User
	refreshTokens map[string]time.Time
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users:         make(map[string]User),
		refreshTokens: make(map[string]time.Time),
	}
}

func (m *memoryStore) CreateUser(ctx context.Context, email, passwordHash string, displayName *string) (User, error) {
	if _, ok := m.users[email]; ok {
		return User{}, ErrEmailAlreadyExists
	}
	user := User{
		ID:           uuid.New(),
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.users[email] = user
	return user, nil
}

func (m *memoryStore) FindUserByEmail(ctx context.Context, email string) (User, error) {
	user, ok := m.users[email]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

func (m *memoryStore) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	m.refreshTokens[tokenHash] = expiresAt
	return nil
}

func (m *memoryStore) RevokeToken(ctx context.Context, userID uuid.UUID, tokenHash string) error {
	delete(m.refreshTokens, tokenHash)
	return nil
}
