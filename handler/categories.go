package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

func (h *Handler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := h.Store.GetAllCategories(r.Context())
	if err != nil {
		http.Error(w, "Failed to get categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (h *Handler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	category, err := h.Store.GetCategoryByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrCategoryNotFound) {
			http.Error(w, "Category not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to get category", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}
