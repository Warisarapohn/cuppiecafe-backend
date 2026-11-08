package mw

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Warisarapohn/cuppiecafe-backend/auth"
)

func TestAuthMiddleware(t *testing.T) {
	issuer := &auth.Issuer{
		Secret: []byte("test-secret-key-for-cuppie"),
		TTL:    time.Hour,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Auth(issuer, nextHandler)

	// Create valid token
	validToken, err := issuer.Issue(
		"550e8400-e29b-41d4-a716-446655440000",
		"member",
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "Missing token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid token",
			token:      "invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Valid token",
			token:      validToken,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
