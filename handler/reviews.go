package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/Warisarapohn/cuppiecafe-backend/mw"
	"github.com/Warisarapohn/cuppiecafe-backend/store"
	"github.com/jackc/pgx/v5/pgconn"
)

// ReviewRequest represents the review data received from the client.
type ReviewRequest struct {
	CafeID           string  `json:"cafe_id"`
	Rating           int     `json:"rating"`
	DrinkRating      *int    `json:"drink_rating"`
	DessertRating    *int    `json:"dessert_rating"`
	AtmosphereRating *int    `json:"atmosphere_rating"`
	ServiceRating    *int    `json:"service_rating"`
	Comment          *string `json:"comment"`
}

// writeJSON sends a JSON response.
func writeReviewJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("WRITE JSON RESPONSE ERROR: %v", err)
	}
}

// decodeReviewRequest decodes JSON and rejects unknown fields or extra JSON values.
func decodeReviewRequest(w http.ResponseWriter, r *http.Request) (ReviewRequest, bool) {
	var req ReviewRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return req, false
	}

	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "Request must contain only one JSON object", http.StatusBadRequest)
		return req, false
	}

	return req, true
}

// validateReview checks review rating values.
func validateReview(req ReviewRequest) string {
	if req.Rating < 1 || req.Rating > 5 {
		return "Rating must be between 1 and 5"
	}

	subRatings := []*int{
		req.DrinkRating,
		req.DessertRating,
		req.AtmosphereRating,
		req.ServiceRating,
	}

	for _, rating := range subRatings {
		if rating != nil && (*rating < 1 || *rating > 5) {
			return "All sub-ratings must be between 1 and 5"
		}
	}

	if req.Comment != nil && len([]rune(*req.Comment)) > 2000 {
		return "Comment must not exceed 2000 characters"
	}

	return ""
}

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

	writeReviewJSON(w, http.StatusOK, reviews)
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

	writeReviewJSON(w, http.StatusOK, review)
}

// CreateReview creates a new review.
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(mw.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req, ok := decodeReviewRequest(w, r)
	if !ok {
		return
	}

	if req.CafeID == "" {
		http.Error(w, "Cafe ID is required", http.StatusBadRequest)
		return
	}

	if message := validateReview(req); message != "" {
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	review := store.Review{
		UserID:           userID,
		CafeID:           req.CafeID,
		Rating:           req.Rating,
		DrinkRating:      req.DrinkRating,
		DessertRating:    req.DessertRating,
		AtmosphereRating: req.AtmosphereRating,
		ServiceRating:    req.ServiceRating,
		Comment:          req.Comment,
	}

	err := h.Store.CreateReview(r.Context(), review)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				http.Error(w, "You have already reviewed this cafe", http.StatusConflict)
				return
			case "23503":
				http.Error(w, "Cafe or user not found", http.StatusBadRequest)
				return
			case "23514":
				http.Error(w, "Review data violates database constraints", http.StatusBadRequest)
				return
			}
		}

		log.Printf("CREATE REVIEW ERROR: %v", err)
		http.Error(w, "Failed to create review", http.StatusInternalServerError)
		return
	}

	writeReviewJSON(w, http.StatusCreated, map[string]string{
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

	reviewID := r.PathValue("id")
	if reviewID == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}

	req, ok := decodeReviewRequest(w, r)
	if !ok {
		return
	}

	if message := validateReview(req); message != "" {
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	review := store.Review{
		ID:               reviewID,
		UserID:           userID,
		Rating:           req.Rating,
		DrinkRating:      req.DrinkRating,
		DessertRating:    req.DessertRating,
		AtmosphereRating: req.AtmosphereRating,
		ServiceRating:    req.ServiceRating,
		Comment:          req.Comment,
	}

	err := h.Store.UpdateReview(r.Context(), review)
	if err != nil {
		if errors.Is(err, store.ErrReviewNotFound) {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			http.Error(w, "Review data violates database constraints", http.StatusBadRequest)
			return
		}

		log.Printf("UPDATE REVIEW ERROR: %v", err)
		http.Error(w, "Failed to update review", http.StatusInternalServerError)
		return
	}

	writeReviewJSON(w, http.StatusOK, map[string]string{
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

	writeReviewJSON(w, http.StatusOK, map[string]string{
		"message": "Review deleted successfully",
	})
}
