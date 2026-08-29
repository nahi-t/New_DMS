package user

import (
	"net/http"

	"github.com/docmanage_new/internal/auth"
)

func SetupUserRoutes(mux *http.ServeMux, h *UserHandler, mw *auth.Middleware) {
	// Public
	mux.HandleFunc("POST /api/users/register", h.Register)

	// Admin Only: Fetch all users list
	mux.Handle("GET /api/users", mw.Authenticate(
		mw.RequireRole(RoleAdmin)(http.HandlerFunc(h.GetAll)),
	))

	// Owner or Admin
	mux.Handle("GET /api/users/{id}", mw.Authenticate(http.HandlerFunc(h.GetUser)))
	mux.Handle("PUT /api/users/{id}", mw.Authenticate(http.HandlerFunc(h.Update)))
	mux.Handle("PUT /api/users/{id}/password", mw.Authenticate(http.HandlerFunc(h.UpdatePassword)))

	// Admin Only: Delete user
	mux.Handle("DELETE /api/users/{id}", mw.Authenticate(
		mw.RequireRole(RoleAdmin)(http.HandlerFunc(h.Delete)),
	))
}
