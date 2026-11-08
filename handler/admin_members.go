package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

// AdminGetAllMembers retrieves all member accounts.
func (h *Handler) AdminGetAllMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.Store.GetAllMembers(r.Context())
	if err != nil {
		http.Error(w, "Unable to retrieve members", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// AdminGetMemberByID retrieves a member account.
func (h *Handler) AdminGetMemberByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "Member ID is required", http.StatusBadRequest)
		return
	}

	member, err := h.Store.GetMemberByIDForAdmin(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrMemberNotFound) {
			http.Error(w, "Member not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Unable to retrieve member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

// AdminDeactivateMember deactivates a member account.
func (h *Handler) AdminDeactivateMember(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "Member ID is required", http.StatusBadRequest)
		return
	}

	err := h.Store.DeactivateMember(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrMemberNotFound) {
			http.Error(w, "Member not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Unable to deactivate member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
