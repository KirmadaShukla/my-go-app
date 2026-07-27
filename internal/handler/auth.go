package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"my-go-app/internal/middleware"
	"my-go-app/internal/model"
	"my-go-app/internal/service"
)

// RegisterRequest is the Joi-style schema for POST /auth/register.
type RegisterRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=255"`
	Email        string `json:"email" validate:"required,email,max=255"`
	Password     string `json:"password" validate:"required,min=8,max=128"`
	Gender       string `json:"gender" validate:"required,gender"`
	MotherName   string `json:"mother_name" validate:"required,min=1,max=255"`
	FatherName   string `json:"father_name" validate:"required,min=1,max=255"`
	MobileNumber string `json:"mobile_number" validate:"required,mobile"`
	ChildAge     int    `json:"child_age" validate:"required,gte=1,lte=25"`
	ChildClass   string `json:"child_class" validate:"required,child_class"`
}

// LoginRequest is the Joi-style schema for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type authResponse struct {
	Token     string  `json:"token"`
	ExpiresAt string  `json:"expires_at"`
	User      userDTO `json:"user"`
}

type userDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Gender       string `json:"gender"`
	MotherName   string `json:"mother_name"`
	FatherName   string `json:"father_name"`
	MobileNumber string `json:"mobile_number"`
	ChildAge     int    `json:"child_age"`
	ChildClass   string `json:"child_class"`
}

func toUserDTO(u *model.User) userDTO {
	return userDTO{
		ID:           u.ID.String(),
		Name:         u.Name,
		Email:        u.Email,
		Gender:       u.Gender,
		MotherName:   u.MotherName,
		FatherName:   u.FatherName,
		MobileNumber: u.MobileNumber,
		ChildAge:     u.ChildAge,
		ChildClass:   u.ChildClass,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req, ok := middleware.BodyFromContext[RegisterRequest](r.Context())
	if !ok {
		writeError(w, http.StatusBadRequest, "missing validated body")
		return
	}

	result, err := h.auth.Register(r.Context(), service.RegisterInput{
		Name:         strings.TrimSpace(req.Name),
		Email:        strings.TrimSpace(req.Email),
		Password:     req.Password,
		Gender:       strings.ToLower(strings.TrimSpace(req.Gender)),
		MotherName:   strings.TrimSpace(req.MotherName),
		FatherName:   strings.TrimSpace(req.FatherName),
		MobileNumber: strings.TrimSpace(req.MobileNumber),
		ChildAge:     req.ChildAge,
		ChildClass:   strings.TrimSpace(req.ChildClass),
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
		User:      toUserDTO(result.User),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, ok := middleware.BodyFromContext[LoginRequest](r.Context())
	if !ok {
		writeError(w, http.StatusBadRequest, "missing validated body")
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
		User:      toUserDTO(result.User),
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

	writeJSON(w, http.StatusOK, toUserDTO(user))
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
