package user

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/docmanage_new/internal/auth"
)

type UserHandler struct {
	DB *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// Request and Response DTOs

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UpdatePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

func toUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

// Handlers

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
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
		log.Printf("ERROR: Registration failed: %v", err)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, `{"error": "Username or email already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toUserResponse(user))
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	user, err := GetUserByID(h.DB, id)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			return
		}
		log.Printf("ERROR: Fetch user failed: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toUserResponse(user))
}

// Helper to check if the caller is the resource owner OR an admin
func (h *UserHandler) canModifyUser(r *http.Request, targetID int64) bool {
	// Extract caller's UserID and Role from r.Context() (set by auth.Middleware)
	callerID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		return false
	}
	callerRole, _ := auth.GetRoleFromContext(r.Context())

	// Admin can modify anyone; regular users can only modify themselves
	return callerRole == RoleAdmin || callerID == targetID
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// 🔒 Check ownership / admin privileges
	if !h.canModifyUser(r, id) {
		http.Error(w, `{"error": "Forbidden: You can only update your own account"}`, http.StatusForbidden)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Only admins are allowed to change user roles
	callerRole, _ := auth.GetRoleFromContext(r.Context())
	if req.Role != "" && callerRole != RoleAdmin {
		http.Error(w, `{"error": "Forbidden: Only admins can alter user roles"}`, http.StatusForbidden)
		return
	}

	updatedUser, err := UpdateUser(h.DB, id, req.Username, req.Email, req.Role)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toUserResponse(updatedUser))
}

func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// 🔒 Check ownership / admin privileges
	if !h.canModifyUser(r, id) {
		http.Error(w, `{"error": "Forbidden: You can only update your own password"}`, http.StatusForbidden)
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, `{"error": "New password is required"}`, http.StatusBadRequest)
		return
	}

	err = UpdatePassword(h.DB, id, req.NewPassword)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Password updated successfully"}`))
}
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	err = DeleteUser(h.DB, id)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			return
		}
		log.Printf("ERROR: Delete user failed: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "User deleted successfully"}`))
}

// Helper to extract {id} parameter from Go 1.22+ net/http ServeMux path
func parseIDParam(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	return strconv.ParseInt(idStr, 10, 64)
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Query all users from your DB layer

	users, err := GetAllUsers(h.DB)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch users"}`, http.StatusInternalServerError)
		return
	}

	resp := make([]UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toUserResponse(&u))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
