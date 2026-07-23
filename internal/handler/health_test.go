package handler_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"my-go-app/internal/handler"
)

func TestHealthz(t *testing.T) {
	h := handler.New(slog.New(slog.NewTextHandler(os.Stderr, nil)), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	h.Healthz(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "alive" {
		t.Errorf("status = %q, want alive", body["status"])
	}
}
