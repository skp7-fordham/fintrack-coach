package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort              string
	Environment             string
	DatabaseURL             string
	JWTSecret               string
	JWTAccessTokenTTL       time.Duration
	RedisURL                string
	ImportQueueName         string
	ImportUploadDir         string
	ImportMaxFileSize       int64
	ImportMaxRows           int
	ImportWorkerConcurrency int
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

	maxFileSize, err := getEnvInt64("IMPORT_MAX_FILE_SIZE", 5<<20)
	if err != nil {
		return Config{}, err
	}
	maxRows, err := getEnvInt("IMPORT_MAX_ROWS", 10000)
	if err != nil {
		return Config{}, err
	}
	concurrency, err := getEnvInt("IMPORT_WORKER_CONCURRENCY", 2)
	if err != nil {
		return Config{}, err
	}
	if concurrency < 1 {
		return Config{}, fmt.Errorf("IMPORT_WORKER_CONCURRENCY must be at least 1")
	}

	return Config{
		ServerPort:              getEnv("SERVER_PORT", "8080"),
		Environment:             getEnv("APP_ENV", "development"),
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://fintrack:fintrack@localhost:5433/fintrack?sslmode=disable"),
		JWTSecret:               secret,
		JWTAccessTokenTTL:       ttl,
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379/0"),
		ImportQueueName:         getEnv("IMPORT_QUEUE_NAME", "transaction_import_jobs"),
		ImportUploadDir:         getEnv("IMPORT_UPLOAD_DIR", "./var/imports"),
		ImportMaxFileSize:       maxFileSize,
		ImportMaxRows:           maxRows,
		ImportWorkerConcurrency: concurrency,
	}, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %w", key, err)
	}
	return value, nil
}

func getEnvInt64(key string, fallback int64) (int64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %w", key, err)
	}
	return value, nil
}
