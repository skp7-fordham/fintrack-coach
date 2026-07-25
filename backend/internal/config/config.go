package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	ServerPort        string
	Environment       string
	DatabaseURL       string
	JWTSecret         string
	JWTAccessTokenTTL time.Duration
}

func Load() (Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	ttlRaw := getEnv("JWT_ACCESS_TOKEN_TTL", "15m")
	ttl, err := time.ParseDuration(ttlRaw)
	if err != nil {
		return Config{}, fmt.Errorf("JWT_ACCESS_TOKEN_TTL is invalid: %w", err)
	}
	if ttl <= 0 {
		return Config{}, fmt.Errorf("JWT_ACCESS_TOKEN_TTL must be greater than zero")
	}

	return Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		Environment:       getEnv("APP_ENV", "development"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://fintrack:fintrack@localhost:5433/fintrack?sslmode=disable"),
		JWTSecret:         secret,
		JWTAccessTokenTTL: ttl,
	}, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
