package router

import (
	"net/http"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/handlers"
)

type Handlers struct {
	Transactions *handlers.TransactionHandler
	Dashboard    *handlers.DashboardHandler
}

func New(h Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /transactions", h.Transactions.List)
	mux.HandleFunc("POST /transactions", h.Transactions.Create)
	mux.HandleFunc("GET /dashboard/summary", h.Dashboard.Summary)
	mux.HandleFunc("GET /dashboard/category-spending", h.Dashboard.CategorySpending)
	mux.HandleFunc("GET /dashboard/monthly-trends", h.Dashboard.MonthlyTrends)
	mux.HandleFunc("GET /dashboard/recent-transactions", h.Dashboard.RecentTransactions)
	return mux
}
