package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

const issuer = "fintrack-coach"

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *TokenManager) TTL() time.Duration {
	return m.ttl
}

type accessClaims struct {
	Email  string `json:"email"`
	IsDemo bool   `json:"is_demo,omitempty"`
	jwt.RegisteredClaims
}

func (m *TokenManager) IssueAccessToken(userID, email string, isDemo bool) (token string, expiresInSec int64, err error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)

	claims := accessClaims{
		Email:  email,
		IsDemo: isDemo,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}

	return signed, int64(m.ttl.Seconds()), nil
}

func (m *TokenManager) ParseAccessToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&accessClaims{},
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.secret, nil
		},
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("missing subject")
	}

	return &domain.TokenClaims{
		UserID: claims.Subject,
		Email:  claims.Email,
		IsDemo: claims.IsDemo,
	}, nil
}
