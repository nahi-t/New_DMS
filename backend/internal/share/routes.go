package share

import (
	"net/http"
	"strings"

	"github.com/docmanage_new/internal/auth"
)

// RegisterRoutes wires the share endpoints into the mux.
//
// Auth model:
//   - /api/shares/shared-with-me is open to any authenticated user.
//   - /api/documents/{id}/share[...] is authenticated only at the route
//     level. The service applies the real rule (admin OR manager OR
//     folder owner OR document uploader). Do NOT add RequireRole here.
func RegisterRoutes(mux *http.ServeMux, h *Handler, mw *auth.Middleware) {
	// Shared-with-me: any authenticated user.
	mux.Handle(
		"/api/shares/shared-with-me",
		mw.Authenticate(http.HandlerFunc(h.SharedWithMe)),
	)

	// Document-scoped share management: authenticated, then service decides.
	shareHandler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/share"):
			h.ShareDocument(w, r)
			return

		case r.Method == http.MethodDelete && strings.Contains(path, "/share/"):
			h.UnshareDocument(w, r)
			return

		case r.Method == http.MethodGet && strings.HasSuffix(path, "/shares"):
			h.ListShares(w, r)
			return
		}

		http.NotFound(w, r)
	}))

	mux.Handle("/api/documents/", shareHandler)
}
