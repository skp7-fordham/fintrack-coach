package domain

import (
	"errors"
	"time"
)

var (
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrDemoModeUnavailable = errors.New("demo mode unavailable")
	ErrDemoAILimitReached  = errors.New("demo AI limit reached")
)

// User is the public user representation (never includes password hash).
type User struct {
	ID        string
	Email     string
	IsDemo    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserWithPasswordHash is used internally for login verification.
type UserWithPasswordHash struct {
	User
	PasswordHash string
}

// RegisterUserInput is the validated registration payload for persistence.
type RegisterUserInput struct {
	Email        string
	PasswordHash string
}

// LoginInput is the validated login payload.
type LoginInput struct {
	Email    string
	Password string
}

// AuthResult is returned after successful register or login.
type AuthResult struct {
	User         User
	AccessToken  string
	TokenType    string
	ExpiresInSec int64
}

// TokenClaims holds the authenticated identity extracted from a JWT.
type TokenClaims struct {
	UserID string
	Email  string
	IsDemo bool
}
