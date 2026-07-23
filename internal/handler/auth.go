package handler

import (
	"errors"
	"net/http"
	"time"

	"my-go-app/internal/middleware"
	"my-go-app/internal/service"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token     string  `json:"token"`
	ExpiresAt string  `json:"expires_at"`
	User      userDTO `json:"user"`
}

type userDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	result, err := h.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
		User: userDTO{
			ID:    result.User.ID.String(),
			Email: result.User.Email,
		},
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	result, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
		User: userDTO{
			ID:    result.User.ID.String(),
			Email: result.User.Email,
		},
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.auth.Me(r.Context(), userID)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userDTO{
		ID:    user.ID.String(),
		Email: user.Email,
	})
}

func (h *Handler) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrEmailTaken):
		writeError(w, http.StatusConflict, "email already registered")
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid email or password")
	default:
		h.logger.Error("auth error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
