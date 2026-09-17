// package document

// import (
// 	"net/http"

// 	"github.com/docmanage_new/internal/auth"
// )

// // RegisterRoutes registers all document routes under the /api prefix.
// func RegisterRoutes(mux *http.ServeMux, handler *Handler, authMW *auth.Middleware) {

// 	mux.Handle("POST /api/folders/{folderId}/documents",
// 		authMW.Authenticate(http.HandlerFunc(handler.Upload)))
// 	mux.Handle("GET /api/folders/{folderId}/documents",
// 		authMW.Authenticate(http.HandlerFunc(handler.List)))
// 	mux.Handle("GET /api/documents/{id}",
// 		authMW.Authenticate(http.HandlerFunc(handler.Download)))
// 	mux.Handle("DELETE /api/documents/{id}",
// 		authMW.Authenticate(http.HandlerFunc(handler.Delete)))

// 	mux.Handle("PATCH /api/documents/{id}", authMW.Authenticate(http.HandlerFunc(handler.Rename)))
// 	mux.Handle("PATCH /api/documents/{id}/move", authMW.Authenticate(http.HandlerFunc(handler.Move)))
// 	mux.Handle("GET /api/documents", authMW.Authenticate(http.HandlerFunc(handler.Search)))
// 	mux.Handle("PUT /api/documents/{id}/content", authMW.Authenticate(http.HandlerFunc(handler.UpdateDocumentContentHandler)))

// 	mux.Handle("PATCH /api/documents/{id}/status", authMW.Authenticate(http.HandlerFunc(handler.UpdateStatusHandler)))

// }

package document

import (
	"net/http"

	"github.com/docmanage_new/internal/auth"
)

// RegisterRoutes wires every document-level endpoint.
//
// IMPORTANT: version-specific paths (/api/versions/...) and the version
// listing endpoint (/api/documents/{id}/versions) live in the documentversion
// package — do NOT duplicate them here or the ServeMux will panic on startup.
func RegisterRoutes(
	mux *http.ServeMux,
	handler *Handler,
	authMW *auth.Middleware,
) {
	/* -------------------------------------------------------------- */
	/*  Upload + list — scoped to a folder                             */
	/* -------------------------------------------------------------- */

	mux.Handle("POST /api/folders/{folderId}/documents",
		authMW.Authenticate(http.HandlerFunc(handler.Upload)))

	mux.Handle("GET /api/folders/{folderId}/documents",
		authMW.Authenticate(http.HandlerFunc(handler.List)))

	/* -------------------------------------------------------------- */
	/*  Search                                                         */
	/* -------------------------------------------------------------- */

	mux.Handle("GET /api/documents",
		authMW.Authenticate(http.HandlerFunc(handler.Search)))

	/* -------------------------------------------------------------- */
	/*  Single document — metadata, download, delete                   */
	/* -------------------------------------------------------------- */

	mux.Handle("GET /api/documents/{id}",
		authMW.Authenticate(http.HandlerFunc(handler.Download)))

	mux.Handle("DELETE /api/documents/{id}",
		authMW.Authenticate(http.HandlerFunc(handler.Delete)))

	/* -------------------------------------------------------------- */
	/*  Updates — rename, move, status                                 */
	/* -------------------------------------------------------------- */

	mux.Handle("PATCH /api/documents/{id}",
		authMW.Authenticate(http.HandlerFunc(handler.Rename)))

	mux.Handle("PATCH /api/documents/{id}/move",
		authMW.Authenticate(http.HandlerFunc(handler.Move)))

	mux.Handle("PATCH /api/documents/{id}/status",
		authMW.Authenticate(http.HandlerFunc(handler.UpdateStatusHandler)))

	/* -------------------------------------------------------------- */
	/*  Upload a NEW version of an existing document                   */
	/*  (creates v2, v3, ... and re-points the master row)             */
	/* -------------------------------------------------------------- */

	mux.Handle("PUT /api/documents/{id}/content",
		authMW.Authenticate(http.HandlerFunc(handler.UpdateDocumentContentHandler)))
}
