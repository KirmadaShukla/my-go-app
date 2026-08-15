package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env             string
	HTTPAddr        string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration

	DatabaseURL string

	JWTSecret string
	JWTExpiry time.Duration

	DefaultChatModel string
	// OpenAI temporarily unused (kept for rollback).
	OpenAIAPIKey    string
	OpenAIChatModel string
	OpenAITTSModel  string
	OpenAITTSVoice  string

	// Gemini (active free tutor provider).
	GeminiAPIKey    string
	GeminiChatModel string
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:              getEnv("APP_ENV", "development"),
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		ShutdownTimeout:  getDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:      getDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:     getDuration("HTTP_WRITE_TIMEOUT", 120*time.Second),
		IdleTimeout:      getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://pranjalshukla@localhost:5432/mygoapp?sslmode=disable"),
		JWTSecret:        getEnv("JWT_SECRET", "dev-only-change-me"),
		JWTExpiry:        getDuration("JWT_EXPIRY", 24*time.Hour),
		DefaultChatModel: getEnv("DEFAULT_CHAT_MODEL", "gemini"),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		OpenAIChatModel:  getEnv("OPENAI_CHAT_MODEL", "gpt-4o-mini"),
		OpenAITTSModel:   getEnv("OPENAI_TTS_MODEL", "tts-1"),
		OpenAITTSVoice:   getEnv("OPENAI_TTS_VOICE", "nova"),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		GeminiChatModel:  getEnv("GEMINI_CHAT_MODEL", "gemini-2.5-flash-lite"),
	}

	level, err := parseLogLevel(getEnv("LOG_LEVEL", "info"))
	if err != nil {
		return nil, err
	}
	cfg.LogLevel = level

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.HTTPAddr == "" {
		return fmt.Errorf("HTTP_ADDR must not be empty")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT must be positive")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL must not be empty")
	}
	if c.JWTExpiry <= 0 {
		return fmt.Errorf("JWT_EXPIRY must be positive")
	}
	if c.IsProduction() && (c.JWTSecret == "" || c.JWTSecret == "dev-only-change-me") {
		return fmt.Errorf("JWT_SECRET must be set to a strong secret in production")
	}
	return nil
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.Env, "production")
}

// ChatProvider returns the normalized tutor backend: "openai" or "gemini".
func (c *Config) ChatProvider() string {
	switch strings.ToLower(strings.TrimSpace(c.DefaultChatModel)) {
	case "openAI", "chatgpt", "gpt":
		return "openai"
	case "gemini", "google", "":
		return "gemini"
	default:
		return strings.ToLower(strings.TrimSpace(c.DefaultChatModel))
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		sec, ierr := strconv.Atoi(v)
		if ierr != nil {
			return fallback
		}
		return time.Duration(sec) * time.Second
	}
	return d
}

func parseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid LOG_LEVEL %q (use debug|info|warn|error)", level)
	}
}
