package folder

import (
	"net/http"

	"github.com/docmanage_new/internal/auth"
)

func SetupFolderRoutes(mux *http.ServeMux, h *Handler, mw *auth.Middleware) {
	// All authenticated users (user, manager, admin) can view folders
	mux.Handle("GET /api/folders", mw.Authenticate(http.HandlerFunc(h.GetAll)))

	// Only Managers and Admins can create folders
	mux.Handle("POST /api/folders", mw.Authenticate(
		mw.RequireRole("manager", "admin")(http.HandlerFunc(h.Create)),
	))

	// Only Admins can delete folders
	mux.Handle("DELETE /api/folders/{id}", mw.Authenticate(
		mw.RequireRole("admin")(http.HandlerFunc(h.Delete)),
	))
}
