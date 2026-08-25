package user

import (
	"net/http"

	"github.com/docmanage_new/internal/auth"
)

func SetupUserRoutes(mux *http.ServeMux, h *UserHandler, mw *auth.Middleware) {
	// Public endpoint
	mux.HandleFunc("POST /api/users/register", h.Register)

	// Protected endpoint: Any authenticated user
	mux.Handle("GET /api/users/{id}", mw.Authenticate(
		http.HandlerFunc(h.GetUser),
	))

	// Protected endpoint: Admin, Manager, or User
	mux.Handle("PUT /api/users/{id}", mw.Authenticate(
		mw.RequireRole(RoleAdmin, RoleManager, RoleUser)(http.HandlerFunc(h.Update)),
	))

	// Protected endpoint: Any authenticated user
	mux.Handle("PATCH /api/users/{id}/password", mw.Authenticate(
		http.HandlerFunc(h.UpdatePassword),
	))

	// Protected endpoint: Admin only
	mux.Handle("DELETE /api/users/{id}", mw.Authenticate(
		mw.RequireRole(RoleAdmin)(http.HandlerFunc(h.Delete)),
	))
}
