package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"my-go-app/internal/config"
	"my-go-app/internal/database"
	"my-go-app/internal/server"
)

// Set via -ldflags at build time.
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	db, err := database.Connect(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	srv := server.New(cfg, logger, db)

	go func() {
		logger.Info("server starting",
			"addr", cfg.HTTPAddr,
			"env", cfg.Env,
			"version", version,
			"commit", commit,
			"build_time", buildTime,
		)
		if err := srv.Start(); err != nil {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped", "uptime", time.Since(srv.StartedAt()).String())
}
