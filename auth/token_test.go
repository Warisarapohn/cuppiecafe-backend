package auth

import (
	"testing"
	"time"
)

func TestTokenIssueAndParse(t *testing.T) {
	issuer := &Issuer{
		Secret: []byte("test-secret-key-for-cuppie"),
		TTL:    time.Hour,
	}

	token, err := issuer.Issue("550e8400-e29b-41d4-a716-446655440000", "member")
	if err != nil {
		t.Fatal(err)
	}

	id, role, err := issuer.Parse(token)
	if err != nil {
		t.Fatal(err)
	}

	if id != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("unexpected member ID: %s", id)
	}

	if role != "member" {
		t.Errorf("unexpected role: %s", role)
	}
}

func TestTokenExpired(t *testing.T) {
	issuer := &Issuer{
		Secret: []byte("test-secret-key-for-cuppie"),
		TTL:    -time.Hour,
	}

	token, err := issuer.Issue("550e8400-e29b-41d4-a716-446655440000", "member")
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = issuer.Parse(token)
	if err == nil {
		t.Fatal("expired token should be rejected")
	}
}
