package router

import "net/http"

func registerHealth(mux *http.ServeMux, d Deps) {
	mux.HandleFunc("GET /healthz", d.Handler.Healthz)
	mux.HandleFunc("GET /readyz", d.Handler.Readyz)
	mux.HandleFunc("GET /", d.Handler.Root)
}
