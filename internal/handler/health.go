package handler

import (
	"net/http"

	"my-go-app/internal/database"
)

func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "my-go-app",
		"status":  "ok",
	})
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	if err := database.Ping(h.db); err != nil {
		h.logger.Error("readiness check failed", "error", err)
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
