package auth

import "errors"

var (
	// ErrEmailAlreadyExists indicates the email is already registered.
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials is returned when authentication fails.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserNotFound signals that the user could not be located.
	ErrUserNotFound = errors.New("user not found")
	// ErrUnauthorized represents missing or invalid authentication tokens.
	ErrUnauthorized = errors.New("unauthorized")
)
