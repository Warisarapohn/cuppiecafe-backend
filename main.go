package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Warisarapohn/cuppiecafe-backend/handler"
	"github.com/Warisarapohn/cuppiecafe-backend/mw"
	"github.com/Warisarapohn/cuppiecafe-backend/store"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	db, err := store.NewDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Initialize Store and Handlers
	s := store.NewStore(db)
	h := handler.New(s)
	fh := handler.NewFoursquareHandler(s)

	// Create router
	mux := http.NewServeMux()

	// Health Check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Cuppie Backend is running"))
	})

	// Authentication
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)

	// Cafe
	mux.HandleFunc("GET /api/cafes", h.GetAllCafes)
	mux.HandleFunc("GET /api/cafes/{id}", h.GetCafeByID)

	// Categories
	mux.HandleFunc("GET /api/categories", h.GetAllCategories)
	mux.HandleFunc("GET /api/categories/{id}", h.GetCategoryByID)

	// Reviews - Public
	mux.HandleFunc("GET /api/reviews/cafe/{id}", h.GetReviewsByCafeID)
	mux.HandleFunc("GET /api/reviews/{id}", h.GetReviewByID)

	// Reviews - Member
	mux.Handle("POST /api/reviews",
		mw.Auth(h.Issuer, s, http.HandlerFunc(h.CreateReview)),
	)
	mux.Handle("PUT /api/reviews/{id}",
		mw.Auth(h.Issuer, s, http.HandlerFunc(h.UpdateReview)),
	)
	mux.Handle("DELETE /api/reviews/{id}",
		mw.Auth(h.Issuer, s, http.HandlerFunc(h.DeleteReview)),
	)

	// Favorites - Member
	mux.Handle("GET /api/favorites",
		mw.Auth(h.Issuer, s, http.HandlerFunc(h.GetFavorites)),
	)
	mux.Handle("POST /api/favorites",
		mw.Auth(h.Issuer, s, http.HandlerFunc(h.AddFavorite)),
	)
	mux.Handle("DELETE /api/favorites/{cafeID}",
		mw.Auth(h.Issuer, s, http.HandlerFunc(h.DeleteFavorite)),
	)

	// Admin - Cafe Management
	mux.Handle("POST /api/admin/cafes",
		mw.Auth(h.Issuer, s,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminCreateCafe)),
		),
	)
	mux.Handle("PUT /api/admin/cafes/{id}",
		mw.Auth(h.Issuer, s,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminUpdateCafe)),
		),
	)
	mux.Handle("DELETE /api/admin/cafes/{id}",
		mw.Auth(h.Issuer, s,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminDeactivateCafe)),
		),
	)

	// Admin - Review Management
	mux.Handle("GET /api/admin/reviews",
		mw.Auth(h.Issuer, s,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminGetAllReviews)),
		),
	)
	mux.Handle("DELETE /api/admin/reviews/{id}",
		mw.Auth(h.Issuer, s,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminDeleteReview)),
		),
	)

	// Admin - Member Management
	mux.Handle("GET /api/admin/members",
		mw.Auth(h.Issuer, s,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminGetAllMembers)),
		),
	)
	mux.Handle("GET /api/admin/members/{id}",
		mw.Auth(h.Issuer, s,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminGetMemberByID)),
		),
	)
	mux.Handle("DELETE /api/admin/members/{id}",
		mw.Auth(h.Issuer, s,
			mw.RequireRole("admin", http.HandlerFunc(h.AdminDeactivateMember)),
		),
	)

	// Foursquare
	mux.HandleFunc("GET /api/foursquare/cafes", fh.SearchCafes)

	// Middleware
	app := mw.CORS(mw.Logging(mux))

	// Render Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Cuppie Backend running on port %s", port)

	if err := http.ListenAndServe(":"+port, app); err != nil {
		log.Fatal("Server failed:", err)
	}
}
