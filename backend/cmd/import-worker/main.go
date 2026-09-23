package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/config"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/database"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/importworker"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/queue"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/repository"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	pool, err := database.NewPostgresPool(dbCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("connected to postgres")

	importQueue, err := queue.NewImportQueue(cfg.RedisURL, cfg.ImportQueueName)
	if err != nil {
		logger.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer importQueue.Close()
	logger.Info("connected to redis")

	importRepo := repository.NewImportRepository(pool)
	processor := service.NewImportProcessor(importRepo, cfg.ImportMaxRows, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	importworker.Start(ctx, &wg, importQueue, processor, cfg.ImportWorkerConcurrency, logger)

	<-ctx.Done()
	logger.Info("shutting down import worker")
	wg.Wait()
	logger.Info("import worker stopped")
}
