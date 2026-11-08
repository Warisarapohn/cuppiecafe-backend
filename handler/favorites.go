package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Warisarapohn/cuppiecafe-backend/mw"
	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

func (h *Handler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(mw.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	favorites, err := h.Store.GetFavoritesByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get favorites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	var request struct {
		CafeID string `json:"cafe_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.CafeID == "" {
		http.Error(w, "Cafe ID is required", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(mw.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.Store.AddFavorite(r.Context(), userID, request.CafeID)
	if err != nil {
		http.Error(w, "Failed to add favorite", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Favorite added successfully",
	})
}

func (h *Handler) DeleteFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(mw.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cafeID := r.PathValue("id")
	if cafeID == "" {
		http.Error(w, "Cafe ID is required", http.StatusBadRequest)
		return
	}

	err := h.Store.DeleteFavorite(r.Context(), userID, cafeID)
	if err != nil {
		if errors.Is(err, store.ErrFavoriteNotFound) {
			http.Error(w, "Favorite not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to delete favorite", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Favorite deleted successfully",
	})
}
