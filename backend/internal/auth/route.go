package auth

import "net/http"

func SetupAuthRoutes(mux *http.ServeMux, handler *AuthHandler) {
	mux.HandleFunc("POST /api/auth/login", handler.Login)
}
