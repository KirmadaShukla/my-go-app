package config_test

import (
	"log/slog"
	"testing"

	"my-go-app/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("DATABASE_URL", "postgres://pranjalshukla@localhost:5432/mygoapp?sslmode=disable")
	t.Setenv("JWT_SECRET", "dev-only-change-me")
	t.Setenv("JWT_EXPIRY", "24h")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want development", cfg.Env)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
	if cfg.IsProduction() {
		t.Error("IsProduction() = true, want false")
	}
}

func TestLoadRejectsWeakJWTInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "dev-only-change-me")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected error for weak JWT_SECRET in production")
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "trace")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected error for invalid LOG_LEVEL")
	}
}
