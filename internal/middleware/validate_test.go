package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"my-go-app/internal/handler"
	"my-go-app/internal/middleware"
)

func TestValidateJSONRegisterSuccess(t *testing.T) {
	var got handler.RegisterRequest
	h := middleware.ValidateJSON[handler.RegisterRequest](func(w http.ResponseWriter, r *http.Request) {
		body, ok := middleware.BodyFromContext[handler.RegisterRequest](r.Context())
		if !ok {
			t.Fatal("expected validated body in context")
		}
		got = body
		w.WriteHeader(http.StatusOK)
	})

	payload := map[string]any{
		"name":          "Riya Sharma",
		"email":         "you@example.com",
		"password":      "secret123",
		"gender":        "Female",
		"mother_name":   "Anita Sharma",
		"father_name":   "Rahul Sharma",
		"mobile_number": "9876543210",
		"child_age":     8,
		"child_class":   "3",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	if got.Gender != "Female" {
		t.Fatalf("gender = %q", got.Gender)
	}
}

func TestValidateJSONInvalidGender(t *testing.T) {
	h := middleware.ValidateJSON[handler.RegisterRequest](func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not run on validation failure")
	})

	payload := map[string]any{
		"name":          "Riya Sharma",
		"email":         "you@example.com",
		"password":      "secret123",
		"gender":        "unknown",
		"mother_name":   "Anita Sharma",
		"father_name":   "Rahul Sharma",
		"mobile_number": "9876543210",
		"child_age":     8,
		"child_class":   "3",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", rr.Code, rr.Body.String())
	}
}
