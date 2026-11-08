package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

func (h *Handler) AdminCreateCafe(w http.ResponseWriter, r *http.Request) {
	var cafe store.Cafe

	if err := json.NewDecoder(r.Body).Decode(&cafe); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cafe.Name = strings.TrimSpace(cafe.Name)
	if cafe.Name == "" {
		http.Error(w, "Cafe name is required", http.StatusBadRequest)
		return
	}

	createdCafe, err := h.Store.CreateCafe(r.Context(), cafe)
	if err != nil {
		http.Error(w, "Unable to create cafe", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdCafe)
}

func (h *Handler) AdminUpdateCafe(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Cafe ID is required", http.StatusBadRequest)
		return
	}

	var cafe store.Cafe
	if err := json.NewDecoder(r.Body).Decode(&cafe); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cafe.Name = strings.TrimSpace(cafe.Name)
	if cafe.Name == "" {
		http.Error(w, "Cafe name is required", http.StatusBadRequest)
		return
	}

	updatedCafe, err := h.Store.UpdateCafe(r.Context(), id, cafe)
	if err != nil {
		if errors.Is(err, store.ErrCafeNotFound) {
			http.Error(w, "Cafe not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unable to update cafe", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCafe)
}

func (h *Handler) AdminDeactivateCafe(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Cafe ID is required", http.StatusBadRequest)
		return
	}

	err := h.Store.DeactivateCafe(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrCafeNotFound) {
			http.Error(w, "Cafe not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unable to deactivate cafe", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
