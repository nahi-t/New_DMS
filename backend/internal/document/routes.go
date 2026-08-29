package document

import (
	"net/http"

	"github.com/docmanage_new/internal/auth"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler, authMW *auth.Middleware) {
	mux.Handle("POST /folders/{folderId}/documents", authMW.Authenticate(handler.Upload))
	mux.Handle("GET /folders/{folderId}/documents", authMW.Authenticate(handler.List))
	mux.Handle("GET /documents/{id}", authMW.Authenticate(handler.Download))
	mux.Handle("DELETE /documents/{id}", authMW.Authenticate(handler.Delete))
}
