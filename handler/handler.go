package handler

import (
	"os"
	"time"

	"github.com/Warisarapohn/cuppiecafe-backend/auth"
	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

type Handler struct {
	Store  *store.Store
	Issuer *auth.Issuer
}

func New(s *store.Store) *Handler {
	return &Handler{
		Store: s,
		Issuer: &auth.Issuer{
			Secret: []byte(os.Getenv("JWT_SECRET")),
			TTL:    24 * time.Hour,
		},
	}
}
