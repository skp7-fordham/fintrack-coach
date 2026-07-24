package router

import (
	"net/http"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/handlers"
)

type Handlers struct {
	Transactions *handlers.TransactionHandler
}

func New(h Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /transactions", h.Transactions.Create)
	return mux
}
