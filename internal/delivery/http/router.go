package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter initializes and configures the standard chi router.
func NewRouter(dbPool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	// Use standard middleware for recovery and logging
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize handlers
	healthHandler := NewHealthHandler(dbPool)

	// Register routes
	r.Get("/health", healthHandler.HealthStatus)

	return r
}
