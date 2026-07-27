package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"my-go-app/internal/auth"
	"my-go-app/internal/config"
	"my-go-app/internal/handler"
	"my-go-app/internal/openai"
	"my-go-app/internal/repository"
	"my-go-app/internal/router"
	"my-go-app/internal/service"

	"gorm.io/gorm"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	startedAt  time.Time
}

func New(cfg *config.Config, logger *slog.Logger, db *gorm.DB) *Server {
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiry)
	users := repository.NewUserRepository(db)
	tutors := repository.NewTutorRepository(db)
	ai := openai.NewClient(cfg.OpenAIAPIKey, cfg.OpenAIChatModel, cfg.OpenAITTSModel, cfg.OpenAITTSVoice)
	authSvc := service.NewAuthService(users, tokens)
	tutorSvc := service.NewTutorService(users, tutors, ai)
	h := handler.New(logger, db, authSvc, tutorSvc)

	httpHandler := router.New(router.Deps{
		Handler: h,
		Tokens:  tokens,
		Logger:  logger,
	})

	return &Server{
		logger: logger,
		httpServer: &http.Server{
			Addr:         cfg.HTTPAddr,
			Handler:      httpHandler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	s.startedAt = time.Now().UTC()
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) StartedAt() time.Time {
	return s.startedAt
}
