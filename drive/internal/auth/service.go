package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abduss/godrive/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	refreshTokenLength = 48
	maxPasswordLength  = 72 // bcrypt limit
)

// userStore abstracts the persistence layer.
type userStore interface {
	CreateUser(ctx context.Context, email, passwordHash string, displayName *string) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	RevokeToken(ctx context.Context, userID uuid.UUID, tokenHash string) error
}

// Service encapsulates authentication use cases.
type Service struct {
	store    userStore
	cfg      config.AuthConfig
	nowFunc  func() time.Time
	idIssuer string
	parser   *jwt.Parser
}

// NewService creates a Service with dependencies.
func NewService(store userStore, cfg config.AuthConfig) *Service {
	return &Service{
		store:    store,
		cfg:      cfg,
		nowFunc:  time.Now,
		idIssuer: "godrive",
		parser:   jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name})),
	}
}

// RegisterInput carries data for user registration.
type RegisterInput struct {
	Email       string
	Password    string
	DisplayName *string
}

// LoginInput carries login credentials.
type LoginInput struct {
	Email    string
	Password string
}

// AuthResult contains user and token information.
type AuthResult struct {
	User   User
	Tokens TokenPair
}

// UserClaims describes the validated identity extracted from an access token.
type UserClaims struct {
	UserID    uuid.UUID
	Email     string
	IsAdmin   bool
	ExpiresAt time.Time
	IssuedAt  time.Time
}

// Register creates a new user, hashing the password and issuing tokens.
func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	if err := validateCredentials(input.Email, input.Password); err != nil {
		return AuthResult{}, err
	}

	hashedPassword, err := hashPassword(input.Password, s.cfg.BcryptCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.store.CreateUser(ctx, strings.ToLower(input.Email), hashedPassword, input.DisplayName)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return AuthResult{}, ErrEmailAlreadyExists
		}
		return AuthResult{}, fmt.Errorf("create user: %w", err)
	}

	result, err := s.issueTokens(ctx, user)
	if err != nil {
		return AuthResult{}, err
	}

	return result, nil
}

// Login authenticates credentials and issues a fresh token pair.
func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	if err := validateCredentials(input.Email, input.Password); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	user, err := s.store.FindUserByEmail(ctx, strings.ToLower(input.Email))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user)
}

// ValidateAccessToken verifies the token signature and extracts user claims.
func (s *Service) ValidateAccessToken(tokenString string) (UserClaims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return UserClaims{}, ErrUnauthorized
	}

	parsed, err := s.parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.AccessTokenSecret), nil
	})
	if err != nil || !parsed.Valid {
		return UserClaims{}, ErrUnauthorized
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return UserClaims{}, ErrUnauthorized
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return UserClaims{}, ErrUnauthorized
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return UserClaims{}, ErrUnauthorized
	}

	email, _ := claims["email"].(string)
	isAdmin, _ := claims["is_admin"].(bool)

	expFloat, okExp := claims["exp"].(float64)
	if !okExp {
		return UserClaims{}, ErrUnauthorized
	}
	exp := time.Unix(int64(expFloat), 0)

	iat := time.Time{}
	if iatFloat, ok := claims["iat"].(float64); ok {
		iat = time.Unix(int64(iatFloat), 0)
	}

	if exp.Before(s.nowFunc()) {
		return UserClaims{}, ErrUnauthorized
	}

	return UserClaims{
		UserID:    userID,
		Email:     email,
		IsAdmin:   isAdmin,
		ExpiresAt: exp,
		IssuedAt:  iat,
	}, nil
}

func (s *Service) issueTokens(ctx context.Context, user User) (AuthResult, error) {
	now := s.nowFunc()

	accessToken, accessExpiry, err := s.generateAccessToken(user, now)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, refreshExpiry, err := s.generateRefreshToken(now)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshHash := hashRefreshToken(refreshToken, s.cfg.RefreshTokenSecret)
	if err := s.store.StoreRefreshToken(ctx, user.ID, refreshHash, refreshExpiry); err != nil {
		return AuthResult{}, fmt.Errorf("store refresh token: %w", err)
	}

	return AuthResult{
		User: user.SafeUser(),
		Tokens: TokenPair{
			AccessToken:        accessToken,
			AccessTokenExpiry:  accessExpiry,
			RefreshToken:       refreshToken,
			RefreshTokenExpiry: refreshExpiry,
		},
	}, nil
}

func (s *Service) generateAccessToken(user User, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(s.cfg.AccessTokenTTL)
	claims := jwt.MapClaims{
		"sub":      user.ID.String(),
		"iss":      s.idIssuer,
		"aud":      "godrive-api",
		"iat":      now.Unix(),
		"exp":      expiresAt.Unix(),
		"email":    user.Email,
		"is_admin": user.IsAdmin,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.AccessTokenSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (s *Service) generateRefreshToken(now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(s.cfg.RefreshTokenTTL)

	raw := make([]byte, refreshTokenLength)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, expiresAt, nil
}

func hashPassword(password string, cost int) (string, error) {
	if len(password) > maxPasswordLength {
		return "", fmt.Errorf("password exceeds maximum length of %d characters", maxPasswordLength)
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func hashRefreshToken(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func validateCredentials(email, password string) error {
	if len(strings.TrimSpace(email)) == 0 || len(strings.TrimSpace(password)) == 0 {
		return ErrInvalidCredentials
	}

	if len(password) < 8 || len(password) > maxPasswordLength {
		return ErrInvalidCredentials
	}
	return nil
}
