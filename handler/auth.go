package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Warisarapohn/cuppiecafe-backend/auth"
	"github.com/Warisarapohn/cuppiecafe-backend/store"
	"github.com/jackc/pgx/v5/pgconn"
)

// Register Request
type registerRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register Response
type registerResponse struct {
	Message string `json:"message"`
}

// Register API
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req registerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.FullName == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, email and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Check duplicate email
	_, err := h.Store.GetMemberByEmail(ctx, req.Email)

	if err == nil {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	if !errors.Is(err, store.ErrMemberNotFound) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Could not process password", http.StatusInternalServerError)
		return
	}

	member := store.Member{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := h.Store.CreateMember(ctx, member); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}

		http.Error(w, "Could not create account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(registerResponse{
		Message: "Registration successful",
	})
}

// Login Request
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login Response
type loginResponse struct {
	Message  string `json:"message"`
	Token    string `json:"token"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// Login API
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req loginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Find active member by email
	member, err := h.Store.GetMemberByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Check password
	if err := auth.CheckPassword(req.Password, member.PasswordHash); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := h.Issuer.Issue(member.ID, member.Role)
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(loginResponse{
		Message:  "Login successful",
		Token:    token,
		FullName: member.FullName,
		Email:    member.Email,
		Role:     member.Role,
	})
}
