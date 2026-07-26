package router

import (
	"net/http"

	"my-go-app/internal/handler"
	"my-go-app/internal/middleware"
)

func registerAuth(mux *http.ServeMux, d Deps) {
	mux.HandleFunc("POST /auth/register", middleware.ValidateJSON[handler.RegisterRequest](d.Handler.Register))
	mux.HandleFunc("POST /auth/login", middleware.ValidateJSON[handler.LoginRequest](d.Handler.Login))
	mux.HandleFunc("GET /auth/me", middleware.Protect(d.Tokens, d.Handler.Me))
}
