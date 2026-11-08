package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

// AdminGetAllReviews retrieves all reviews.
func (h *Handler) AdminGetAllReviews(w http.ResponseWriter, r *http.Request) {
	reviews, err := h.Store.GetAllReviews(r.Context())
	if err != nil {
		http.Error(w, "Unable to retrieve reviews", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

// AdminDeleteReview deletes a review.
func (h *Handler) AdminDeleteReview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}

	err := h.Store.AdminDeleteReview(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrReviewNotFound) {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Unable to delete review", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
