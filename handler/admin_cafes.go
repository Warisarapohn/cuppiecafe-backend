package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/Warisarapohn/cuppiecafe-backend/store"
	"github.com/jackc/pgx/v5/pgconn"
)

type AdminCafeRequest struct {
	CategoryID   *string  `json:"category_id"`
	Name         string   `json:"name"`
	Description  *string  `json:"description"`
	Address      *string  `json:"address"`
	PriceMin     *float64 `json:"price_min"`
	PriceMax     *float64 `json:"price_max"`
	OpeningHours *string  `json:"opening_hours"`
	ContactInfo  *string  `json:"contact_info"`
}

func decodeAdminCafeRequest(r *http.Request) (AdminCafeRequest, error) {
	var req AdminCafeRequest

	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		return req, err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return req, errors.New("request must contain only one JSON object")
	}

	return req, nil
}

func validateAdminCafe(req *AdminCafeRequest) error {
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		return errors.New("Cafe name is required")
	}

	if utf8.RuneCountInString(req.Name) > 200 {
		return errors.New("Cafe name must not exceed 200 characters")
	}

	if req.PriceMin != nil && *req.PriceMin < 0 {
		return errors.New("Minimum price cannot be negative")
	}

	if req.PriceMax != nil && *req.PriceMax < 0 {
		return errors.New("Maximum price cannot be negative")
	}

	return nil
}

func (req AdminCafeRequest) toStoreCafe() store.Cafe {
	return store.Cafe{
		CategoryID:   req.CategoryID,
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		PriceMin:     req.PriceMin,
		PriceMax:     req.PriceMax,
		OpeningHours: req.OpeningHours,
		ContactInfo:  req.ContactInfo,
	}
}

func handleCafeDatabaseError(w http.ResponseWriter, err error) {
	log.Printf("Admin cafe database error: %v", err)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		case "23514":
			http.Error(w, "Cafe data violates database constraints", http.StatusBadRequest)
			return
		case "23505":
			http.Error(w, "Cafe already exists", http.StatusConflict)
			return
		}
	}

	http.Error(w, "Unable to process cafe", http.StatusInternalServerError)
}

// POST /api/admin/cafes
func (h *Handler) AdminCreateCafe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req, err := decodeAdminCafeRequest(r)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateAdminCafe(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdCafe, err := h.Store.CreateCafe(r.Context(), req.toStoreCafe())
	if err != nil {
		handleCafeDatabaseError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createdCafe)
}

// PUT /api/admin/cafes/{id}
func (h *Handler) AdminUpdateCafe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Cafe ID is required", http.StatusBadRequest)
		return
	}

	req, err := decodeAdminCafeRequest(r)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateAdminCafe(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedCafe, err := h.Store.UpdateCafe(
		r.Context(),
		id,
		req.toStoreCafe(),
	)

	if err != nil {
		if errors.Is(err, store.ErrCafeNotFound) {
			http.Error(w, "Cafe not found", http.StatusNotFound)
			return
		}

		handleCafeDatabaseError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updatedCafe)
}

// DELETE /api/admin/cafes/{id}
func (h *Handler) AdminDeactivateCafe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

		handleCafeDatabaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
