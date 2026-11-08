package main

import (
	"log"
	"net/http"

	"github.com/Warisarapohn/cuppiecafe-backend/handler"
	"github.com/Warisarapohn/cuppiecafe-backend/mw"
	"github.com/Warisarapohn/cuppiecafe-backend/store"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to database
	db, err := store.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize handlers
	s := store.NewStore(db)
	h := handler.New(s)
	fh := handler.NewFoursquareHandler(s)

	// Setup API routes
	mux := http.NewServeMux()

	// =========================
	// Authentication API
	// =========================
	mux.HandleFunc("/api/auth/register", h.Register)
	mux.HandleFunc("/api/auth/login", h.Login)

	// =========================
	// Cafe API - Public
	// =========================
	mux.HandleFunc("GET /api/cafes", h.GetAllCafes)
	mux.HandleFunc("GET /api/cafes/{id}", h.GetCafeByID)

	// =========================
	// Category API - Public
	// =========================
	mux.HandleFunc("GET /api/categories", h.GetAllCategories)
	mux.HandleFunc("GET /api/categories/{id}", h.GetCategoryByID)

	// =========================
	// Review API - Public
	// =========================
	mux.HandleFunc("GET /api/cafes/{id}/reviews", h.GetReviewsByCafeID)
	mux.HandleFunc("GET /api/reviews/{id}", h.GetReviewByID)

	// =========================
	// Review API - Member only
	// =========================
	mux.Handle(
		"POST /api/reviews",
		mw.Auth(h.Issuer, http.HandlerFunc(h.CreateReview)),
	)
	mux.Handle(
		"PUT /api/reviews/{id}",
		mw.Auth(h.Issuer, http.HandlerFunc(h.UpdateReview)),
	)
	mux.Handle(
		"DELETE /api/reviews/{id}",
		mw.Auth(h.Issuer, http.HandlerFunc(h.DeleteReview)),
	)

	// =========================
	// Favorites API - Member only
	// =========================
	mux.Handle(
		"GET /api/favorites",
		mw.Auth(h.Issuer, http.HandlerFunc(h.GetFavorites)),
	)
	mux.Handle(
		"POST /api/favorites",
		mw.Auth(h.Issuer, http.HandlerFunc(h.AddFavorite)),
	)
	mux.Handle(
		"DELETE /api/favorites/{id}",
		mw.Auth(h.Issuer, http.HandlerFunc(h.DeleteFavorite)),
	)

	// =========================
	// Admin Cafe API
	// =========================
	mux.Handle(
		"POST /api/admin/cafes",
		mw.Auth(h.Issuer,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminCreateCafe)),
		),
	)
	mux.Handle(
		"PUT /api/admin/cafes/{id}",
		mw.Auth(h.Issuer,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminUpdateCafe)),
		),
	)
	mux.Handle(
		"DELETE /api/admin/cafes/{id}",
		mw.Auth(h.Issuer,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminDeactivateCafe)),
		),
	)

	// =========================
	// Admin Review API
	// =========================
	mux.Handle(
		"GET /api/admin/reviews",
		mw.Auth(h.Issuer,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminGetAllReviews)),
		),
	)
	mux.Handle(
		"DELETE /api/admin/reviews/{id}",
		mw.Auth(h.Issuer,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminDeleteReview)),
		),
	)

	// =========================
	// Admin Member API
	// =========================
	mux.Handle(
		"GET /api/admin/members",
		mw.Auth(h.Issuer,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminGetAllMembers)),
		),
	)
	mux.Handle(
		"GET /api/admin/members/{id}",
		mw.Auth(h.Issuer,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminGetMemberByID)),
		),
	)
	mux.Handle(
		"DELETE /api/admin/members/{id}",
		mw.Auth(h.Issuer,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminDeactivateMember)),
		),
	)

	// =========================
	// Foursquare API - Public
	// =========================
	mux.HandleFunc("GET /api/foursquare/cafes", fh.SearchCafes)

	// =========================
	// Middleware
	// =========================
	var server http.Handler = mux
	server = mw.CORS(server)
	server = mw.Logging(server)

	// =========================
	// Start Server
	// =========================
	log.Println("Cuppie API listening on :8080")

	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatal(err)
	}
}
