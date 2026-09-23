package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	RunImportWorkerInAPI    bool
	AIAPIKey                string
	AIBaseURL               string
	AIModel                 string
	AITimeout               time.Duration
	AIMaxToolIterations     int
	CORSAllowedOrigins      []string
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

	runWorkerInAPI, err := getEnvBool("RUN_IMPORT_WORKER_IN_API", false)
	if err != nil {
		return Config{}, err
	}

	aiTimeoutRaw := getEnv("AI_TIMEOUT", "30s")
	aiTimeout, err := time.ParseDuration(aiTimeoutRaw)
	if err != nil {
		return Config{}, fmt.Errorf("AI_TIMEOUT is invalid: %w", err)
	}
	if aiTimeout <= 0 {
		return Config{}, fmt.Errorf("AI_TIMEOUT must be greater than zero")
	}

	aiMaxIterations, err := getEnvInt("AI_MAX_TOOL_ITERATIONS", 5)
	if err != nil {
		return Config{}, err
	}
	if aiMaxIterations < 1 {
		return Config{}, fmt.Errorf("AI_MAX_TOOL_ITERATIONS must be at least 1")
	}

	origins, err := parseCORSOrigins(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		ServerPort:              listenPort(),
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
		RunImportWorkerInAPI:    runWorkerInAPI,
		AIAPIKey:                os.Getenv("AI_API_KEY"),
		AIBaseURL:               getEnv("AI_BASE_URL", "https://api.openai.com/v1"),
		AIModel:                 getEnv("AI_MODEL", "gpt-4o-mini"),
		AITimeout:               aiTimeout,
		AIMaxToolIterations:     aiMaxIterations,
		CORSAllowedOrigins:      origins,
	}, nil
}

// listenPort prefers Render's PORT over the local SERVER_PORT default.
func listenPort() string {
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return port
	}
	return getEnv("SERVER_PORT", "8080")
}

func parseCORSOrigins(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part == "*" {
			return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS must not include * for authenticated APIs")
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS must include at least one origin")
	}
	return out, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s is invalid: %w", key, err)
	}
	return value, nil
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
