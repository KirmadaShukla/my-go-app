package auth_test

import (
	"testing"
	"time"

	"my-go-app/internal/auth"

	"github.com/google/uuid"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !auth.CheckPassword(hash, "secret123") {
		t.Fatal("expected password to match")
	}
	if auth.CheckPassword(hash, "wrong") {
		t.Fatal("expected password mismatch")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	tm := auth.NewTokenManager("test-secret", time.Hour)
	id := uuid.New()

	token, expiresAt, err := tm.Issue(id, "a@b.com")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("expiresAt should be in the future")
	}

	claims, err := tm.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != id {
		t.Fatalf("UserID = %s, want %s", claims.UserID, id)
	}
	if claims.Email != "a@b.com" {
		t.Fatalf("Email = %q, want a@b.com", claims.Email)
	}
}
