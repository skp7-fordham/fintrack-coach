package router

import (
	"net/http"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/handlers"
)

type Handlers struct {
	Auth         *handlers.AuthHandler
	Transactions *handlers.TransactionHandler
	Dashboard    *handlers.DashboardHandler
	Accounts     *handlers.AccountHandler
	Categories   *handlers.CategoryHandler
	Imports      *handlers.ImportHandler
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

	mux.Handle("POST /accounts", authenticate(http.HandlerFunc(h.Accounts.Create)))
	mux.Handle("GET /accounts", authenticate(http.HandlerFunc(h.Accounts.List)))
	mux.Handle("PATCH /accounts/{id}", authenticate(http.HandlerFunc(h.Accounts.Update)))
	mux.Handle("DELETE /accounts/{id}", authenticate(http.HandlerFunc(h.Accounts.Delete)))

	mux.Handle("POST /categories", authenticate(http.HandlerFunc(h.Categories.Create)))
	mux.Handle("GET /categories", authenticate(http.HandlerFunc(h.Categories.List)))
	mux.Handle("PATCH /categories/{id}", authenticate(http.HandlerFunc(h.Categories.Update)))
	mux.Handle("DELETE /categories/{id}", authenticate(http.HandlerFunc(h.Categories.Delete)))

	mux.Handle("POST /imports/transactions", authenticate(http.HandlerFunc(h.Imports.Create)))
	mux.Handle("GET /imports", authenticate(http.HandlerFunc(h.Imports.List)))
	mux.Handle("GET /imports/{id}", authenticate(http.HandlerFunc(h.Imports.Get)))
	mux.Handle("GET /imports/{id}/errors", authenticate(http.HandlerFunc(h.Imports.ListErrors)))

	return mux
}
