package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/auth"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/service"
)

type disabledDemoAuthRepository struct{}

func (disabledDemoAuthRepository) CreateUser(
	ctx context.Context,
	input domain.RegisterUserInput,
) (*domain.User, error) {
	return nil, nil
}

func (disabledDemoAuthRepository) FindUserByEmail(
	ctx context.Context,
	normalizedEmail string,
) (*domain.UserWithPasswordHash, error) {
	return nil, domain.ErrUserNotFound
}

func TestDemoEndpointReturnsNotFoundWhenDisabled(t *testing.T) {
	svc := service.NewAuthService(
		disabledDemoAuthRepository{},
		auth.NewTokenManager("test-secret", 15*time.Minute),
		service.DemoAuthConfig{},
	)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewAuthHandler(svc, logger)

	req := httptest.NewRequest(http.MethodPost, "/auth/demo", nil)
	rec := httptest.NewRecorder()
	handler.Demo(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if body := rec.Body.String(); body != "{\"error\":\"demo unavailable\"}\n" {
		t.Fatalf("body = %q", body)
	}
}
