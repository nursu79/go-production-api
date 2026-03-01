package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nursu79/go-production-api/internal/delivery/http/handler"
	authMiddleware "github.com/nursu79/go-production-api/internal/middleware"
)

// NewRouter initializes and configures the standard chi router.
func NewRouter(dbPool *pgxpool.Pool, userHandler *handler.UserHandler, jwtSecret string) *chi.Mux {
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

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(authMiddleware.JWTMiddleware(jwtSecret))
			r.Get("/me", userHandler.GetMe)
		})
	})

	return r
}
