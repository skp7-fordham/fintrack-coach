package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/config"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/database"
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
	for i := 0; i < cfg.ImportWorkerConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			logger.Info("import worker started", "worker_id", workerID)
			for {
				jobID, err := importQueue.Dequeue(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, redis.ErrClosed) {
						logger.Info("import worker stopping", "worker_id", workerID)
						return
					}
					logger.Error("failed to dequeue import job", "worker_id", workerID, "err", err)
					select {
					case <-ctx.Done():
						return
					case <-time.After(time.Second):
					}
					continue
				}

				processor.ProcessJob(ctx, jobID)
			}
		}(i + 1)
	}

	<-ctx.Done()
	logger.Info("shutting down import worker")
	wg.Wait()
	logger.Info("import worker stopped")
}
