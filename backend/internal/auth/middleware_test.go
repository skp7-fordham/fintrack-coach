package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareAddsDemoIdentityFromJWT(t *testing.T) {
	tokens := NewTokenManager("test-secret-with-enough-entropy", 15*time.Minute)
	token, _, err := tokens.IssueAccessToken(
		"11111111-1111-1111-1111-111111111111",
		"demo@example.test",
		true,
	)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok || userID != "11111111-1111-1111-1111-111111111111" {
			t.Fatalf("user identity = %q, %v", userID, ok)
		}
		if !IsDemoFromContext(r.Context()) {
			t.Fatal("demo identity was not added to context")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/dashboard/summary", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	Middleware(tokens)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestRequireNonDemoBlocksDemoMutation(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	handler := RequireNonDemo(next)

	req := httptest.NewRequest(http.MethodPost, "/accounts", nil)
	req = req.WithContext(WithIdentity(req.Context(), "11111111-1111-1111-1111-111111111111", true))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if called {
		t.Fatal("protected mutation reached next handler")
	}
	if body := rec.Body.String(); body != "{\"error\":\"demo account is read-only\"}\n" {
		t.Fatalf("body = %q", body)
	}
}

func TestRequireNonDemoAllowsNormalMutation(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	handler := RequireNonDemo(next)

	req := httptest.NewRequest(http.MethodPost, "/accounts", nil)
	req = req.WithContext(WithIdentity(req.Context(), "22222222-2222-2222-2222-222222222222", false))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent || !called {
		t.Fatalf("status = %d, called = %v", rec.Code, called)
	}
}
