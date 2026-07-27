package router

import (
	"log/slog"
	"net/http"

	"my-go-app/internal/auth"
	"my-go-app/internal/handler"
	"my-go-app/internal/middleware"
)

// Deps holds everything route groups need to register.
type Deps struct {
	Handler *handler.Handler
	Tokens  *auth.TokenManager
	Logger  *slog.Logger
}

// New builds the root mux and applies global middleware.
// Add a new feature by creating registerX in its own file and calling it here.
func New(d Deps) http.Handler {
	mux := http.NewServeMux()

	registerHealth(mux, d)
	registerAuth(mux, d)
	registerTutor(mux, d)
	// registerUsers(mux, d)
	// registerProducts(mux, d)

	var h http.Handler = mux
	h = middleware.Logging(d.Logger, h)
	h = middleware.RequestID(h)
	h = middleware.Recovery(d.Logger, h)
	return h
}
