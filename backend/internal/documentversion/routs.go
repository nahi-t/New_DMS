// package documentversion

// import (
// 	"net/http"

// 	"github.com/docmanage_new/internal/auth"
// )

// func RegisterDocumentVersionRoutes(mux *http.ServeMux, handler *Handler, authMW *auth.Middleware) {
// 	//only admin and user see all documantion versions
// 	mux.Handle("GET /api/documentversion", authMW.Authenticate(http.HandlerFunc(handler.GetDocumentVersionHandler)))
// 	mux.Handle("GET /api/documentversion/{id}/download", authMW.Authenticate(http.HandlerFunc(handler.DownloadDocumentVersionHandler)))

// }

package documentversion

import (
	"net/http"

	"github.com/docmanage_new/internal/auth"
)

// RegisterDocumentVersionRoutes wires every version-level endpoint.
//
// IMPORTANT: document/routes.go must NOT register any of these paths.
// Each HTTP pattern can only be registered once per ServeMux.
func RegisterDocumentVersionRoutes(
	mux *http.ServeMux,
	handler *Handler,
	authMW *auth.Middleware,
) {
	/* -------------------------------------------------------------- */
	/*  Version history of a document                                  */
	/* -------------------------------------------------------------- */

	// List every version of a single document.
	// Permission checked inside the handler (owner, manager, admin).
	mux.Handle("GET /api/documents/{id}/versions",
		authMW.Authenticate(http.HandlerFunc(handler.ListVersionsForDocument)))

	/* -------------------------------------------------------------- */
	/*  Global version listing (admin-only)                            */
	/* -------------------------------------------------------------- */

	mux.Handle("GET /api/versions",
		authMW.Authenticate(
			authMW.RequireRole("admin")(
				http.HandlerFunc(handler.GetAllVersions),
			),
		))

	/* -------------------------------------------------------------- */
	/*  Single version — read, download, restore, delete               */
	/* -------------------------------------------------------------- */

	// Get a single version's metadata.
	mux.Handle("GET /api/versions/{id}",
		authMW.Authenticate(http.HandlerFunc(handler.GetVersion)))

	// Download a specific version.
	mux.Handle("GET /api/versions/{id}/download",
		authMW.Authenticate(http.HandlerFunc(handler.DownloadVersion)))

	// Restore a version — re-points the master document to it.
	mux.Handle("POST /api/versions/{id}/restore",
		authMW.Authenticate(http.HandlerFunc(handler.RestoreVersion)))

	// Delete a specific version (admin-only).
	mux.Handle("DELETE /api/versions/{id}",
		authMW.Authenticate(
			authMW.RequireRole("admin")(
				http.HandlerFunc(handler.DeleteVersion),
			),
		))
}
