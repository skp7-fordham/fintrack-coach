package router

import (
	"net/http"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/handlers"
)

func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health)
	return mux
}
