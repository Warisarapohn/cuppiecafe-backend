package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Warisarapohn/cuppiecafe-backend/store"
)

type FoursquareHandler struct {
	Client *store.FoursquareClient
	Store  *store.Store
}

// NewFoursquareHandler creates a Foursquare handler.
func NewFoursquareHandler(s *store.Store) *FoursquareHandler {
	return &FoursquareHandler{
		Client: store.NewFoursquareClient(),
		Store:  s,
	}
}

// SearchCafes searches cafes from Foursquare
// and saves or retrieves them from Supabase.
func (h *FoursquareHandler) SearchCafes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := store.NormalizeCafeQuery(r.URL.Query().Get("query"))

	if query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Search cafes from Foursquare.
	places, err := h.Client.SearchCafes(r.Context(), query)
	if err != nil {
		log.Printf("FOURSQUARE SEARCH ERROR: %v", err)
		http.Error(w, "Failed to search cafes", http.StatusBadGateway)
		return
	}

	// Save or retrieve each cafe from Supabase.
	cafes := make([]store.Cafe, 0, len(places))

	for _, place := range places {
		cafe, err := h.Store.GetOrCreateCafeByFSQID(r.Context(), place)
		if err != nil {
			log.Printf(
				"GET OR CREATE CAFE ERROR: fsq_id=%s error=%v",
				place.FSQID,
				err,
			)

			http.Error(
				w,
				"Failed to save cafe data",
				http.StatusInternalServerError,
			)
			return
		}

		cafes = append(cafes, cafe)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(cafes); err != nil {
		log.Printf("ENCODE CAFE RESULTS ERROR: %v", err)
	}
}
