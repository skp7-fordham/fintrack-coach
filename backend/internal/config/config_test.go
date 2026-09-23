package config

import (
	"strings"
	"testing"
	"time"
)

func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("PORT", "")
	t.Setenv("SERVER_PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("RUN_IMPORT_WORKER_IN_API", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	t.Setenv("AI_TIMEOUT", "")
	t.Setenv("IMPORT_UPLOAD_DIR", "")
}

func TestLoadPORTTakesPriorityOverServerPort(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("PORT", "10000")
	t.Setenv("SERVER_PORT", "8080")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.ServerPort != "10000" {
		t.Fatalf("ServerPort = %q, want 10000", cfg.ServerPort)
	}
}

func TestLoadServerPortWhenPORTUnset(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("SERVER_PORT", "8080")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.ServerPort != "8080" {
		t.Fatalf("ServerPort = %q, want 8080", cfg.ServerPort)
	}
}

func TestLoadRunImportWorkerInAPI(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("RUN_IMPORT_WORKER_IN_API", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !cfg.RunImportWorkerInAPI {
		t.Fatal("expected RunImportWorkerInAPI true")
	}
}

func TestLoadRunImportWorkerInAPIDefaultFalse(t *testing.T) {
	setBaseEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.RunImportWorkerInAPI {
		t.Fatal("expected RunImportWorkerInAPI false")
	}
}

func TestLoadCORSOriginsTrimAndRejectWildcard(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", " http://localhost:3000 , https://fintrack-coach.vercel.app ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("origins = %#v", cfg.CORSAllowedOrigins)
	}
	if cfg.CORSAllowedOrigins[0] != "http://localhost:3000" {
		t.Fatalf("first origin = %q", cfg.CORSAllowedOrigins[0])
	}
	if cfg.CORSAllowedOrigins[1] != "https://fintrack-coach.vercel.app" {
		t.Fatalf("second origin = %q", cfg.CORSAllowedOrigins[1])
	}

	t.Setenv("CORS_ALLOWED_ORIGINS", "*")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for wildcard CORS origin")
	}
}

func TestLoadPreservesDatabaseSSLMode(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("DATABASE_URL", "postgres://user:pass@ep-example.neon.tech/fintrack?sslmode=require")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !strings.Contains(cfg.DatabaseURL, "sslmode=require") {
		t.Fatalf("DATABASE_URL lost sslmode: %s", cfg.DatabaseURL)
	}
	if strings.Contains(cfg.DatabaseURL, "sslmode=disable") {
		t.Fatal("DATABASE_URL should not force sslmode=disable")
	}
}

func TestLoadAITimeoutDefault(t *testing.T) {
	setBaseEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.AITimeout != 30*time.Second {
		t.Fatalf("AITimeout = %s, want 30s", cfg.AITimeout)
	}
	if cfg.ImportUploadDir != "./var/imports" {
		t.Fatalf("ImportUploadDir = %q", cfg.ImportUploadDir)
	}
}
