package mw

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Warisarapohn/cuppiecafe-backend/auth"
	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

// MemberStore defines the methods required by Auth middleware.
type MemberStore interface {
	GetMemberByID(ctx context.Context, id string) (store.Member, error)
}

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

// Auth validates JWT and verifies that the account is still active.
func Auth(issuer *auth.Issuer, s MemberStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Missing authorization token", http.StatusUnauthorized)
			return
		}

		// Validate Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Validate JWT
		userID, _, err := issuer.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Verify that the account exists and is active
		member, err := s.GetMemberByID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, store.ErrMemberNotFound) {
				http.Error(
					w,
					"Account is inactive or not found",
					http.StatusUnauthorized,
				)
				return
			}

			http.Error(
				w,
				"Unable to verify account",
				http.StatusInternalServerError,
			)
			return
		}

		// Use current user ID and role from database
		ctx := context.WithValue(r.Context(), UserIDKey, member.ID)
		ctx = context.WithValue(ctx, RoleKey, member.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
