package middleware

import (
	"context"
	"net/http"
	"strings"

	"my-go-app/internal/auth"

	"github.com/google/uuid"
)

const UserIDKey contextKey = "user_id"
const UserEmailKey contextKey = "user_email"

func Auth(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"missing bearer token"}`))
				return
			}

			raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			claims, err := tokens.Parse(raw)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"invalid or expired token"}`))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// Protect wraps a handler func with JWT auth.
func Protect(tokens *auth.TokenManager, next http.HandlerFunc) http.HandlerFunc {
	authMw := Auth(tokens)
	return func(w http.ResponseWriter, r *http.Request) {
		authMw(next).ServeHTTP(w, r)
	}
}
