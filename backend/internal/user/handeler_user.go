package user

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/docmanage_new/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	DB *pgxpool.Pool
}

func NewUserHandler(db *pgxpool.Pool) *UserHandler {
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
		// PostgreSQL unique constraint violation error check
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(err.Error(), "duplicate key") {
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
	callerID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		return false
	}
	callerRole, _ := auth.GetRoleFromContext(r.Context())

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

	callerRole, _ := auth.GetRoleFromContext(r.Context())

	var updatedUser *User

	// Admin callers can modify profiles and change roles; standard callers can only update their profile
	if callerRole == RoleAdmin {
		// If role is omitted in request, retain current user role
		if req.Role == "" {
			existing, err := GetUserByID(h.DB, id)
			if err != nil {
				if err.Error() == "user not found" {
					http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
					return
				}
				http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
				return
			}
			req.Role = existing.Role
		}
		updatedUser, err = AdminUpdateUser(h.DB, id, req.Username, req.Email, req.Role)
	} else {
		// Regular user self-profile update (role field is ignored)
		updatedUser, err = UpdateProfile(h.DB, id, req.Username, req.Email)
	}

	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			return
		}
		log.Printf("ERROR: Update user failed: %v", err)
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

	// 🔒 Check ownership / admin privileges
	if !h.canModifyUser(r, id) {
		http.Error(w, `{"error": "Forbidden: You can only delete your own account"}`, http.StatusForbidden)
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

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
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

// Helper to extract {id} parameter from Go 1.22+ net/http ServeMux path
func parseIDParam(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	return strconv.ParseInt(idStr, 10, 64)
}
