package logger

import (
	"log/slog"
	"os"
)

// Init initializes the default logger with structured JSON formatting.
func Init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
