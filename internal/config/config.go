package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	AppPort          string   `env:"APP_PORT,required"`
	DBUrl            string   `env:"DB_URL,required"`
	AppEnv           string   `env:"APP_ENV" envDefault:"development"`
	JwtSecret        string   `env:"JWT_SECRET,required"`
	JwtRefreshSecret string   `env:"JWT_REFRESH_SECRET,required"`
	CorsOrigins      []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000"`
}

// String explicitly masks secrets ensuring configuration dumps never leak sensitive properties into structured logs safely.
func (c *Config) String() string {
	return fmt.Sprintf("AppPort:%s | AppEnv:%s | DBUrl:[REDACTED] | JwtSecret:[REDACTED] | JwtRefreshSecret:[REDACTED] | CorsOrigins:[%v]", c.AppPort, c.AppEnv, c.CorsOrigins)
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
