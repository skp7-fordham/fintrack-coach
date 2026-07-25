package router

import (
	"net/http"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/handlers"
)

type Handlers struct {
	Auth         *handlers.AuthHandler
	Transactions *handlers.TransactionHandler
	Dashboard    *handlers.DashboardHandler
}

func New(h Handlers, authenticate func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /auth/register", h.Auth.Register)
	mux.HandleFunc("POST /auth/login", h.Auth.Login)

	mux.Handle("POST /transactions", authenticate(http.HandlerFunc(h.Transactions.Create)))
	mux.Handle("GET /transactions", authenticate(http.HandlerFunc(h.Transactions.List)))
	mux.Handle("GET /dashboard/summary", authenticate(http.HandlerFunc(h.Dashboard.Summary)))
	mux.Handle("GET /dashboard/category-spending", authenticate(http.HandlerFunc(h.Dashboard.CategorySpending)))
	mux.Handle("GET /dashboard/monthly-trends", authenticate(http.HandlerFunc(h.Dashboard.MonthlyTrends)))
	mux.Handle("GET /dashboard/recent-transactions", authenticate(http.HandlerFunc(h.Dashboard.RecentTransactions)))

	return mux
}
