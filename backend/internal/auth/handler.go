package auth

import (
	"encoding/json"
	"log"

	"net/http"

	"github.com/docmanage_new/internal/httpx"
)

type AuthHandler struct {
	Service *AuthService
	Audit   AuditLogger
}

func NewAuthHandler(service *AuthService, audit AuditLogger) *AuthHandler {
	return &AuthHandler{Service: service, Audit: audit}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string    `json:"token"`
	User  *AuthUser `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON body"}`, http.StatusBadRequest)

	}

	ip := httpx.ClientIP(r)
	ua := r.Header.Get("User-Agent")

	user, token, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		if aerr := h.Audit.LogLogin(
			r.Context(),
			nil,       // no user id — failed login
			req.Email, // record the attempted identifier
			"user.login.failed",
			ip, ua, false,
		); aerr != nil {
			log.Printf("audit: login.failed event failed: %v", aerr)
		}
		http.Error(w, `{"message":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Successful login
	if aerr := h.Audit.LogLogin(
		r.Context(),
		&user.ID, // real user id
		user.Username,
		"user.login",
		ip, ua, true,
	); aerr != nil {
		log.Printf("audit: login event failed: %v", aerr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Token: token,
		User:  user,
	})
}
