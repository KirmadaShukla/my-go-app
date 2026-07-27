package handler

import (
	"errors"
	"io"
	"net/http"

	"my-go-app/internal/middleware"
	"my-go-app/internal/service"

	"github.com/google/uuid"
)

// StartTutorSessionRequest starts or resumes a subject voice session.
type StartTutorSessionRequest struct {
	Subject  string `json:"subject" validate:"required,subject"`
	Language string `json:"language" validate:"omitempty,min=2,max=50"`
	ForceNew bool   `json:"force_new"`
}

func (h *Handler) ListTutorSubjects(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"subjects": h.tutor.Subjects(),
		"classes":  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		"mode":     "voice_only",
		"note":     "Voice AI tutor for maths, science, english, and activities. Speaks politely in the student's language.",
	})
}

func (h *Handler) StartTutorSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	req, ok := middleware.BodyFromContext[StartTutorSessionRequest](r.Context())
	if !ok {
		writeError(w, http.StatusBadRequest, "missing validated body")
		return
	}

	result, err := h.tutor.StartSession(r.Context(), service.StartSessionInput{
		UserID:   userID,
		Subject:  req.Subject,
		Language: req.Language,
		ForceNew: req.ForceNew,
	})
	if err != nil {
		h.writeTutorError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"session_id":   result.Session.ID.String(),
		"subject":      result.Session.Subject,
		"language":     result.Session.Language,
		"status":       result.Session.Status,
		"resumed":      result.Resumed,
		"greeting":     result.Greeting,
		"audio_base64": result.AudioBase64,
		"audio_mime":   result.AudioMIME,
	})
}

func (h *Handler) TutorVoice(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessionID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "expected multipart form with audio file")
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		writeError(w, http.StatusBadRequest, "audio file is required (field name: audio)")
		return
	}
	defer file.Close()

	limited := io.LimitReader(file, 10<<20)
	result, err := h.tutor.Voice(r.Context(), service.VoiceInput{
		UserID:    userID,
		SessionID: sessionID,
		Filename:  header.Filename,
		Audio:     limited,
	})
	if err != nil {
		h.writeTutorError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) writeTutorError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation), errors.Is(err, service.ErrInvalidSubject):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrSessionNotFound):
		writeError(w, http.StatusNotFound, "tutor session not found")
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, service.ErrTutorUnavailable):
		h.logger.Error("tutor unavailable", "error", err)
		writeError(w, http.StatusServiceUnavailable, "tutor is temporarily unavailable")
	default:
		h.logger.Error("tutor error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
