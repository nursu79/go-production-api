package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nursu79/go-production-api/internal/config"
	deliveryHttp "github.com/nursu79/go-production-api/internal/delivery/http"
	"github.com/nursu79/go-production-api/internal/delivery/http/handler"
	"github.com/nursu79/go-production-api/internal/infrastructure/redis"
	"github.com/nursu79/go-production-api/internal/repository"
	"github.com/nursu79/go-production-api/internal/usecase"
	"github.com/nursu79/go-production-api/pkg/logger"
)

func main() {
	// Initialize structured logger
	logger.Init()
	slog.Info("Starting API server")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize database connection pool with retry logic
	dbPool, err := initDB(context.Background(), cfg.DBUrl)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Run migrations
	sourceURL := "file://migrations"
	if err := repository.RunMigrations(cfg.DBUrl, sourceURL); err != nil {
		slog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}

	// Initialize Dependency Injection
	redisClient := redis.NewRedisClient(context.Background(), cfg.RedisUrl, cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)

	userRepo := repository.NewUserRepository(dbPool)
	userUsecase := usecase.NewUserUsecase(userRepo, redisClient, cfg.JwtSecret, cfg.JwtRefreshSecret)
	userHandler := handler.NewUserHandler(userUsecase)
	adminHandler := handler.NewAdminHandler(userUsecase)

	// Initialize routing
	router := deliveryHttp.NewRouter(dbPool, redisClient, userHandler, adminHandler, cfg)

	// Configure HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		slog.Info("Server listening", "port", cfg.AppPort)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		slog.Error("Error starting server", "error", err)

	case sig := <-shutdown:
		slog.Info("Graceful shutdown started", "signal", sig)

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Asking listener to shutdown
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Graceful shutdown failed", "error", err)
			err = srv.Close()
			if err != nil {
				slog.Error("Error closing server", "error", err)
			}
		}
		slog.Info("Graceful shutdown completed")
	}
}

// initDB handles the database connection and the "wait-for-db" retry logic.
func initDB(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	maxRetries := 5
	retryDelay := 2 * time.Second

	var pool *pgxpool.Pool
	var err error

	for i := 0; i < maxRetries; i++ {
		pool, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				slog.Info("Successfully connected to database")
				return pool, nil
			}
		}

		slog.Warn("Failed to connect to database, retrying...", "attempt", i+1, "error", err)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("could not connect to database after %d retries: %w", maxRetries, err)
}
