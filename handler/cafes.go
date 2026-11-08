package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

func (h *Handler) GetAllCafes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cafes, err := h.Store.GetAllCafes(r.Context())
	if err != nil {
		http.Error(w, "Could not retrieve cafes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cafes)
}

func (h *Handler) GetCafeByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Cafe ID is required", http.StatusBadRequest)
		return
	}

	cafe, err := h.Store.GetCafeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrCafeNotFound) {
			http.Error(w, "Cafe not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Could not retrieve cafe", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cafe)
}
