package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/database"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/demoseed"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	demoEmail := strings.TrimSpace(os.Getenv("DEMO_USER_EMAIL"))
	if demoEmail == "" {
		logger.Error("DEMO_USER_EMAIL is required")
		os.Exit(1)
	}
	demoPassword := os.Getenv("DEMO_USER_PASSWORD")
	if demoPassword == "" {
		logger.Error("DEMO_USER_PASSWORD is required")
		os.Exit(1)
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer connectCancel()

	pool, err := database.NewPostgresPool(connectCtx, databaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer seedCancel()
	if err := demoseed.Seed(seedCtx, pool, demoEmail, demoPassword, time.Now().UTC()); err != nil {
		logger.Error("failed to seed demo data", "err", err)
		os.Exit(1)
	}

	logger.Info("demo data seeded")
}
