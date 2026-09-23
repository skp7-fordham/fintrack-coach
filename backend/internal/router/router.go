package router

import (
	"net/http"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/auth"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/handlers"
)

type Handlers struct {
	Auth         *handlers.AuthHandler
	Transactions *handlers.TransactionHandler
	Dashboard    *handlers.DashboardHandler
	Accounts     *handlers.AccountHandler
	Categories   *handlers.CategoryHandler
	Imports      *handlers.ImportHandler
	Coach        *handlers.CoachHandler
}

func New(h Handlers, authenticate func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()
	protectMutation := func(handler http.Handler) http.Handler {
		return authenticate(auth.RequireNonDemo(handler))
	}

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /auth/register", h.Auth.Register)
	mux.HandleFunc("POST /auth/login", h.Auth.Login)
	mux.HandleFunc("POST /auth/demo", h.Auth.Demo)

	mux.Handle("POST /transactions", protectMutation(http.HandlerFunc(h.Transactions.Create)))
	mux.Handle("GET /transactions", authenticate(http.HandlerFunc(h.Transactions.List)))
	mux.Handle("GET /dashboard/summary", authenticate(http.HandlerFunc(h.Dashboard.Summary)))
	mux.Handle("GET /dashboard/category-spending", authenticate(http.HandlerFunc(h.Dashboard.CategorySpending)))
	mux.Handle("GET /dashboard/monthly-trends", authenticate(http.HandlerFunc(h.Dashboard.MonthlyTrends)))
	mux.Handle("GET /dashboard/recent-transactions", authenticate(http.HandlerFunc(h.Dashboard.RecentTransactions)))

	mux.Handle("POST /accounts", protectMutation(http.HandlerFunc(h.Accounts.Create)))
	mux.Handle("GET /accounts", authenticate(http.HandlerFunc(h.Accounts.List)))
	mux.Handle("PATCH /accounts/{id}", protectMutation(http.HandlerFunc(h.Accounts.Update)))
	mux.Handle("DELETE /accounts/{id}", protectMutation(http.HandlerFunc(h.Accounts.Delete)))

	mux.Handle("POST /categories", protectMutation(http.HandlerFunc(h.Categories.Create)))
	mux.Handle("GET /categories", authenticate(http.HandlerFunc(h.Categories.List)))
	mux.Handle("PATCH /categories/{id}", protectMutation(http.HandlerFunc(h.Categories.Update)))
	mux.Handle("DELETE /categories/{id}", protectMutation(http.HandlerFunc(h.Categories.Delete)))

	mux.Handle("POST /imports/transactions", protectMutation(http.HandlerFunc(h.Imports.Create)))
	mux.Handle("GET /imports", authenticate(http.HandlerFunc(h.Imports.List)))
	mux.Handle("GET /imports/{id}", authenticate(http.HandlerFunc(h.Imports.Get)))
	mux.Handle("GET /imports/{id}/errors", authenticate(http.HandlerFunc(h.Imports.ListErrors)))

	mux.Handle("POST /coach/chat", authenticate(http.HandlerFunc(h.Coach.Chat)))
	mux.Handle("GET /coach/conversations", authenticate(http.HandlerFunc(h.Coach.ListConversations)))
	mux.Handle("GET /coach/conversations/{id}", authenticate(http.HandlerFunc(h.Coach.GetConversation)))
	mux.Handle("DELETE /coach/conversations/{id}", protectMutation(http.HandlerFunc(h.Coach.DeleteConversation)))

	return mux
}
