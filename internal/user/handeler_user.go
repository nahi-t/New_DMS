package user

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type UserHandler struct {
	DB *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{DB: db}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // Optional: defaults to "user"
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
	}

	if req.Role == "" {
		req.Role = RoleUser
	}

	if req.Role != RoleAdmin && req.Role != RoleManager && req.Role != RoleUser {
		http.Error(w, `{"error": "Invalid role specified"}`, http.StatusBadRequest)
		return
	}

	user, err := RegisterUser(h.DB, req.Username, req.Email, req.Password, req.Role)
	if err != nil {
		// Print the exact underlying error to your server logs
		log.Printf("ERROR: Registration failed: %v", err)

		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, `{"error": "Username or email already exists"}`, http.StatusConflict)
			return
		}

		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	resp := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

}
