package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/auth"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

type authRepository interface {
	CreateUser(ctx context.Context, input domain.RegisterUserInput) (*domain.User, error)
	FindUserByEmail(ctx context.Context, normalizedEmail string) (*domain.UserWithPasswordHash, error)
}

type AuthService struct {
	repo       authRepository
	tokens     *auth.TokenManager
	demoConfig DemoAuthConfig
}

type DemoAuthConfig struct {
	Enabled  bool
	Email    string
	Password string
}

func NewAuthService(repo authRepository, tokens *auth.TokenManager, demoConfig DemoAuthConfig) *AuthService {
	return &AuthService{
		repo:       repo,
		tokens:     tokens,
		demoConfig: demoConfig,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*domain.AuthResult, error) {
	email, err := normalizeAndValidateEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, domain.RegisterUserInput{
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, err
	}

	return s.issueAuthResult(*user)
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*domain.AuthResult, error) {
	email, err := normalizeAndValidateEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if req.Password == "" {
		return nil, &domain.ValidationError{Message: "password is required"}
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.issueAuthResult(user.User)
}

func (s *AuthService) DemoLogin(ctx context.Context) (*domain.AuthResult, error) {
	if !s.demoConfig.Enabled {
		return nil, domain.ErrDemoModeUnavailable
	}

	email := strings.ToLower(strings.TrimSpace(s.demoConfig.Email))
	if email == "" || s.demoConfig.Password == "" {
		return nil, domain.ErrDemoModeUnavailable
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrDemoModeUnavailable
		}
		return nil, err
	}
	if !user.IsDemo {
		return nil, domain.ErrDemoModeUnavailable
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(s.demoConfig.Password),
	); err != nil {
		return nil, domain.ErrDemoModeUnavailable
	}

	return s.issueAuthResult(user.User)
}

func (s *AuthService) issueAuthResult(user domain.User) (*domain.AuthResult, error) {
	token, expiresIn, err := s.tokens.IssueAccessToken(user.ID, user.Email, user.IsDemo)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResult{
		User:         user,
		AccessToken:  token,
		TokenType:    "Bearer",
		ExpiresInSec: expiresIn,
	}, nil
}

func normalizeAndValidateEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", &domain.ValidationError{Message: "email is required"}
	}
	if len(email) > 255 {
		return "", &domain.ValidationError{Message: "email must be at most 255 characters"}
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || !strings.EqualFold(parsed.Address, email) || parsed.Address != email {
		return "", &domain.ValidationError{Message: "email must be a valid email address"}
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 || !strings.Contains(email[at+1:], ".") {
		return "", &domain.ValidationError{Message: "email must be a valid email address"}
	}

	return email, nil
}

func validatePassword(password string) error {
	if password == "" {
		return &domain.ValidationError{Message: "password is required"}
	}
	if len(password) > 72 {
		return &domain.ValidationError{Message: "password must be at most 72 bytes"}
	}
	if utf8.RuneCountInString(password) < 8 {
		return &domain.ValidationError{Message: "password must be at least 8 characters"}
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return &domain.ValidationError{
			Message: "password must include at least one uppercase letter, one lowercase letter, and one digit",
		}
	}

	return nil
}
