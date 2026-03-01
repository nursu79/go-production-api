package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string `env:"APP_PORT,required"`
	DBUrl   string `env:"DB_URL,required"`
	AppEnv  string `env:"APP_ENV" envDefault:"development"`
}

// Load reads configuration from .env and maps it to the Config struct.
func Load() (*Config, error) {
	// Ignore error if .env doesn't exist, we might be relying entirely on environment variables (e.g., in Docker)
	err := godotenv.Load()
	if err != nil {
		slog.Warn("No .env file found or unable to load it. Relying on system environment variables.")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return cfg, nil
}
