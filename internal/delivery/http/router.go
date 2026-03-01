package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nursu79/go-production-api/internal/config"
	"github.com/nursu79/go-production-api/internal/delivery/http/handler"
	"github.com/nursu79/go-production-api/internal/infrastructure/redis"
	authMiddleware "github.com/nursu79/go-production-api/internal/middleware"
)

// NewRouter initializes and configures the standard chi router mapping explicitly rigid chains strictly.
func NewRouter(dbPool *pgxpool.Pool, redisClient *redis.Client, userHandler *handler.UserHandler, adminHandler *handler.AdminHandler, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// 1. Context Hooks
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// 2. Security Armor
	r.Use(authMiddleware.SecureHeaders)
	r.Use(authMiddleware.CORS(cfg.AppEnv, cfg.CorsOrigins))

	// 3. Reliability & Limitations
	r.Use(authMiddleware.RateLimit(5, 20, redisClient)) // Global Limiter mapped to distributed robust Redis environments natively.
	r.Use(authMiddleware.Timeout(30 * time.Second)) // Hang protection timeout

	// 4. Observability & Panics
	r.Use(authMiddleware.RequestLogger())
	r.Use(middleware.Recoverer)

	// Initialize handlers
	healthHandler := NewHealthHandler(dbPool)

	// Register routes
	r.Get("/health", healthHandler.HealthStatus)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)
			
			// Logout requires a valid token to verify what we are blacklisting!
			r.With(authMiddleware.JWTMiddleware(cfg.JwtSecret, redisClient)).Post("/logout", userHandler.Logout)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(authMiddleware.JWTMiddleware(cfg.JwtSecret, redisClient))
			r.Get("/me", userHandler.GetMe)
			r.Put("/me", userHandler.UpdateMe) // Explicit Profile Patch targeting Cache Drops natively
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(authMiddleware.JWTMiddleware(cfg.JwtSecret, redisClient))
			r.Use(authMiddleware.AuthorizeRole("admin"))

			r.Get("/users", adminHandler.GetAllUsers)
			r.Delete("/users/{id}", adminHandler.DeleteUser)
		})
	})

	return r
}
