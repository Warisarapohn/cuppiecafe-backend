package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Warisarapohn/cuppiecafe-backend/mw"
	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

// GetReviewsByCafeID retrieves reviews of a cafe.
func (h *Handler) GetReviewsByCafeID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cafeID := r.PathValue("id")
	if cafeID == "" {
		http.Error(w, "Cafe ID is required", http.StatusBadRequest)
		return
	}

	reviews, err := h.Store.GetReviewsByCafeID(r.Context(), cafeID)
	if err != nil {
		log.Printf("GET REVIEWS ERROR: %v", err)
		http.Error(w, "Failed to get reviews", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

// GetReviewByID retrieves a review by ID.
func (h *Handler) GetReviewByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}

	review, err := h.Store.GetReviewByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrReviewNotFound) {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		log.Printf("GET REVIEW BY ID ERROR: %v", err)
		http.Error(w, "Failed to get review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(review)
}

// CreateReview creates a new review.
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var review store.Review

	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(mw.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if review.CafeID == "" || review.Rating < 1 || review.Rating > 5 {
		http.Error(w, "Cafe ID and rating (1-5) are required", http.StatusBadRequest)
		return
	}

	// ใช้ User ID จาก JWT Token
	review.UserID = userID

	// บันทึกรีวิวลงฐานข้อมูล
	err := h.Store.CreateReview(r.Context(), review)
	if err != nil {
		log.Printf("CREATE REVIEW ERROR: %v", err)
		http.Error(w, "Failed to create review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Review created successfully",
	})
}

// UpdateReview updates a member's own review.
func (h *Handler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(mw.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var review store.Review

	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	review.ID = r.PathValue("id")
	review.UserID = userID

	if review.Rating < 1 || review.Rating > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	err := h.Store.UpdateReview(r.Context(), review)
	if err != nil {
		if errors.Is(err, store.ErrReviewNotFound) {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		log.Printf("UPDATE REVIEW ERROR: %v", err)
		http.Error(w, "Failed to update review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Review updated successfully",
	})
}

// DeleteReview deletes a member's own review.
func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(mw.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	reviewID := r.PathValue("id")
	if reviewID == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}

	err := h.Store.DeleteReview(r.Context(), reviewID, userID)
	if err != nil {
		if errors.Is(err, store.ErrReviewNotFound) {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		log.Printf("DELETE REVIEW ERROR: %v", err)
		http.Error(w, "Failed to delete review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Review deleted successfully",
	})
}
