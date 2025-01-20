package main

import (
	"log/slog"
	"os"
	"strings"
)

func initLogger() *slog.Logger {
	env := strings.ToLower(os.Getenv("ENV"))

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		// Add source file and line number in non-production
		AddSource: env != "production",
	}

	// Use JSON handler in production for better parsing
	var handler slog.Handler
	if env == "production" {
		opts.Level = slog.LevelInfo
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func main() {
	logger := initLogger()
	slog.SetDefault(logger)

	// Structured logging with proper key-value pairs
	slog.Debug("debug message",
		"component", "main",
		"status", "initializing",
	)

	slog.Info("info message",
		"component", "main",
		"user", "john",
		"action", "login",
	)

}
