package repository

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // register postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // register file driver
)

// RunMigrations executes 'up' migrations utilizing golang-migrate/migrate.
func RunMigrations(dbURL string, sourceURL string) error {
	slog.Info("Running database migrations...")
	
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to init migration: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No new migrations to apply")
			return nil
		}
		// Dirty state or other error
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Database migrations applied successfully")
	return nil
}
