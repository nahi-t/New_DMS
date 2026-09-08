package document

import (
	"net/http"

	"github.com/docmanage_new/internal/auth"
)

// RegisterRoutes registers all document routes under the /api prefix.
func RegisterRoutes(mux *http.ServeMux, handler *Handler, authMW *auth.Middleware) {

	mux.Handle("POST /api/folders/{folderId}/documents",
		authMW.Authenticate(http.HandlerFunc(handler.Upload)))
	mux.Handle("GET /api/folders/{folderId}/documents",
		authMW.Authenticate(http.HandlerFunc(handler.List)))
	mux.Handle("GET /api/documents/{id}",
		authMW.Authenticate(http.HandlerFunc(handler.Download)))
	mux.Handle("DELETE /api/documents/{id}",
		authMW.Authenticate(http.HandlerFunc(handler.Delete)))

	mux.Handle("PATCH /api/documents/{id}", authMW.Authenticate(http.HandlerFunc(handler.Rename)))
	mux.Handle("PATCH /api/documents/{id}/move", authMW.Authenticate(http.HandlerFunc(handler.Move)))
	mux.Handle("GET /api/documents", authMW.Authenticate(http.HandlerFunc(handler.Search)))
	mux.Handle("PUT /api/documents/{id}/content", authMW.Authenticate(http.HandlerFunc(handler.UpdateDocumentContentHandler)))
}
