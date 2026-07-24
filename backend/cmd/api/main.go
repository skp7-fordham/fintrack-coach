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

	"github.com/skp7-fordham/fintrack-coach/backend/internal/config"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/database"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/handlers"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/repository"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/router"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	pool, err := database.NewPostgresPool(dbCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("connected to postgres")

	transactionRepo := repository.NewTransactionRepository(pool)
	transactionService := service.NewTransactionService(transactionRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionService, logger)

	srv := &http.Server{
		Addr: ":" + cfg.ServerPort,
		Handler: router.New(router.Handlers{
			Transactions: transactionHandler,
		}),
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
