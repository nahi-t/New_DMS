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

// LoginRequest is the body for POST /api/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    example:"admin@example.com"`
	Password string `json:"password" example:"secret123"`
}

// LoginResponse is returned by POST /api/auth/login.
type LoginResponse struct {
	Token string    `json:"token" example:"eyJhbGciOiJIUzI1..."`
	User  *AuthUser `json:"user"`
}

// Login authenticates a user and returns a JWT.
//
// @Summary      Log in
// @Description  Validate credentials and return a JWT plus the user profile.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Login credentials"
// @Success      200   {object}  LoginResponse
// @Failure      400   {object}  ErrorResponse  "Invalid JSON body"
// @Failure      401   {object}  ErrorResponse  "Invalid email or password"
// @Failure      500   {object}  ErrorResponse  "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	ip := httpx.ClientIP(r)
	ua := r.Header.Get("User-Agent")

	user, token, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		if aerr := h.Audit.LogLogin(
			r.Context(),
			nil,
			req.Email,
			"user.login.failed",
			ip, ua, false,
		); aerr != nil {
			log.Printf("audit: login.failed event failed: %v", aerr)
		}
		http.Error(w, `{"message":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	if aerr := h.Audit.LogLogin(
		r.Context(),
		&user.ID,
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

// ErrorResponse is a generic error envelope.
type ErrorResponse struct {
	Error   string `json:"error,omitempty"   example:"invalid request"`
	Message string `json:"message,omitempty" example:"invalid email or password"`
}
