package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/auth"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

type fakeAuthRepository struct {
	user      *domain.UserWithPasswordHash
	findCalls int
}

func (r *fakeAuthRepository) CreateUser(
	ctx context.Context,
	input domain.RegisterUserInput,
) (*domain.User, error) {
	return &domain.User{
		ID:        "11111111-1111-1111-1111-111111111111",
		Email:     input.Email,
		IsDemo:    false,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (r *fakeAuthRepository) FindUserByEmail(
	ctx context.Context,
	normalizedEmail string,
) (*domain.UserWithPasswordHash, error) {
	r.findCalls++
	if r.user == nil || r.user.Email != normalizedEmail {
		return nil, domain.ErrUserNotFound
	}
	return r.user, nil
}

func newPasswordHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}
	return string(hash)
}

func TestDemoLoginDisabled(t *testing.T) {
	repo := &fakeAuthRepository{}
	svc := NewAuthService(
		repo,
		auth.NewTokenManager("test-secret", 15*time.Minute),
		DemoAuthConfig{},
	)

	_, err := svc.DemoLogin(context.Background())
	if !errors.Is(err, domain.ErrDemoModeUnavailable) {
		t.Fatalf("DemoLogin() error = %v", err)
	}
	if repo.findCalls != 0 {
		t.Fatalf("repository called %d times while demo mode disabled", repo.findCalls)
	}
}

func TestDemoLoginSucceedsAndReturnsDemoClaim(t *testing.T) {
	const password = "DemoPassword1"
	repo := &fakeAuthRepository{
		user: &domain.UserWithPasswordHash{
			User: domain.User{
				ID:        "11111111-1111-1111-1111-111111111111",
				Email:     "demo@example.com",
				IsDemo:    true,
				CreatedAt: time.Now().UTC(),
			},
			PasswordHash: newPasswordHash(t, password),
		},
	}
	tokens := auth.NewTokenManager("test-secret", 15*time.Minute)
	svc := NewAuthService(repo, tokens, DemoAuthConfig{
		Enabled:  true,
		Email:    "demo@example.com",
		Password: password,
	})

	result, err := svc.DemoLogin(context.Background())
	if err != nil {
		t.Fatalf("DemoLogin() error: %v", err)
	}
	if !result.User.IsDemo {
		t.Fatal("expected demo user in auth result")
	}
	claims, err := tokens.ParseAccessToken(result.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}
	if !claims.IsDemo || claims.UserID != result.User.ID {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestNormalLoginUnaffectedForNonDemoUser(t *testing.T) {
	const password = "RegularPassword1"
	repo := &fakeAuthRepository{
		user: &domain.UserWithPasswordHash{
			User: domain.User{
				ID:        "22222222-2222-2222-2222-222222222222",
				Email:     "user@example.com",
				IsDemo:    false,
				CreatedAt: time.Now().UTC(),
			},
			PasswordHash: newPasswordHash(t, password),
		},
	}
	svc := NewAuthService(
		repo,
		auth.NewTokenManager("test-secret", 15*time.Minute),
		DemoAuthConfig{},
	)

	result, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "user@example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("Login() error: %v", err)
	}
	if result.User.IsDemo {
		t.Fatal("normal user incorrectly marked as demo")
	}
}
