package router

import (
	"net/http"

	"my-go-app/internal/middleware"
)

func registerAuth(mux *http.ServeMux, d Deps) {
	mux.HandleFunc("POST /auth/register", d.Handler.Register)
	mux.HandleFunc("POST /auth/login", d.Handler.Login)
	mux.HandleFunc("GET /auth/me", middleware.Protect(d.Tokens, d.Handler.Me))
}
