package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/aiprovider"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/auth"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/coach"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/config"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/database"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/handlers"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/queue"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/repository"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/router"
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

	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTokenTTL)
	authenticate := auth.Middleware(tokenManager)

	authRepo := repository.NewAuthRepository(pool)
	authService := service.NewAuthService(authRepo, tokenManager)
	authHandler := handlers.NewAuthHandler(authService, logger)

	transactionRepo := repository.NewTransactionRepository(pool)
	transactionService := service.NewTransactionService(transactionRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionService, logger)

	dashboardRepo := repository.NewDashboardRepository(pool)
	dashboardService := service.NewDashboardService(dashboardRepo)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService, logger)

	accountRepo := repository.NewAccountRepository(pool)
	accountService := service.NewAccountService(accountRepo)
	accountHandler := handlers.NewAccountHandler(accountService, logger)

	categoryRepo := repository.NewCategoryRepository(pool)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService, logger)

	importRepo := repository.NewImportRepository(pool)
	importService := service.NewImportService(importRepo, importQueue, cfg.ImportUploadDir, cfg.ImportMaxFileSize)
	importHandler := handlers.NewImportHandler(importService, logger, cfg.ImportMaxFileSize)

	var llm coach.LLM
	if cfg.AIAPIKey != "" {
		llm = aiprovider.NewClient(cfg.AIAPIKey, cfg.AIBaseURL, cfg.AITimeout)
		logger.Info("coach llm configured", "model", cfg.AIModel, "base_url", cfg.AIBaseURL)
	} else {
		logger.Warn("AI_API_KEY not set; coach chat will return unavailable")
	}

	toolRegistry := coach.NewToolRegistry(dashboardService, accountService, importService, logger)
	coachAgent := coach.NewAgent(llm, toolRegistry, cfg.AIModel, cfg.AIMaxToolIterations, logger)
	coachRepo := repository.NewCoachRepository(pool)
	coachService := coach.NewService(coachRepo, coachAgent, logger)
	coachHandler := handlers.NewCoachHandler(coachService, logger)

	srv := &http.Server{
		Addr: ":" + cfg.ServerPort,
		Handler: router.New(router.Handlers{
			Auth:         authHandler,
			Transactions: transactionHandler,
			Dashboard:    dashboardHandler,
			Accounts:     accountHandler,
			Categories:   categoryHandler,
			Imports:      importHandler,
			Coach:        coachHandler,
		}, authenticate),
	}

	go func() {
		logger.Info("starting server", "addr", srv.Addr, "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
