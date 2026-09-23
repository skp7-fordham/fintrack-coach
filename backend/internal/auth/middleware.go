package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

type errorBody struct {
	Error string `json:"error"`
}

func Middleware(tokens *TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeUnauthorized(w, "authentication required")
				return
			}

			scheme, token, ok := strings.Cut(header, " ")
			if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			claims, err := tokens.ParseAccessToken(strings.TrimSpace(token))
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			if !isValidUUID(claims.UserID) {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			ctx := WithIdentity(r.Context(), claims.UserID, claims.IsDemo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireNonDemo blocks persistent mutations for signed demo identities.
func RequireNonDemo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsDemoFromContext(r.Context()) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(errorBody{Error: "demo account is read-only"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(errorBody{Error: message})
}

func isValidUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, c := range value {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !isHex(c) {
				return false
			}
		}
	}
	return true
}

func isHex(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}
