package mw

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Warisarapohn/cuppiecafe-backend/auth"
	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

// fakeMemberStore simulates member database operations.
type fakeMemberStore struct {
	member store.Member
	err    error
}

func (f *fakeMemberStore) GetMemberByID(
	ctx context.Context,
	id string,
) (store.Member, error) {
	if f.err != nil {
		return store.Member{}, f.err
	}

	if f.member.ID != id {
		return store.Member{}, store.ErrMemberNotFound
	}

	return f.member, nil
}

func TestAuthMiddleware(t *testing.T) {
	issuer := &auth.Issuer{
		Secret: []byte("test-secret-key-for-cuppie"),
		TTL:    time.Hour,
	}

	userID := "550e8400-e29b-41d4-a716-446655440000"

	activeStore := &fakeMemberStore{
		member: store.Member{
			ID:       userID,
			Role:     "member",
			IsActive: true,
		},
	}

	inactiveStore := &fakeMemberStore{
		err: store.ErrMemberNotFound,
	}

	databaseErrorStore := &fakeMemberStore{
		err: errors.New("database connection failed"),
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a valid JWT
	validToken, err := issuer.Issue(userID, "member")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		token       string
		memberStore MemberStore
		wantStatus  int
	}{
		{
			name:        "Missing token",
			token:       "",
			memberStore: activeStore,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "Invalid token",
			token:       "invalid-token",
			memberStore: activeStore,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "Valid token with active member",
			token:       validToken,
			memberStore: activeStore,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "Token with inactive member",
			token:       validToken,
			memberStore: inactiveStore,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "Database error",
			token:       validToken,
			memberStore: databaseErrorStore,
			wantStatus:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Auth(issuer, tt.memberStore, nextHandler)

			req := httptest.NewRequest(
				http.MethodGet,
				"/protected",
				nil,
			)

			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf(
					"got status %d, want %d",
					rec.Code,
					tt.wantStatus,
				)
			}
		})
	}
}
