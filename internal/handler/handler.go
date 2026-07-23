package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"my-go-app/internal/service"

	"gorm.io/gorm"
)

type Handler struct {
	logger *slog.Logger
	db     *gorm.DB
	auth   *service.AuthService
}

func New(logger *slog.Logger, db *gorm.DB, auth *service.AuthService) *Handler {
	return &Handler{logger: logger, db: db, auth: auth}
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
