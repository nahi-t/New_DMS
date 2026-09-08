package documentversion

import (
	"net/http"

	"github.com/docmanage_new/internal/auth"
)

func RegisterDocumentVersionRoutes(mux *http.ServeMux, handler *Handler, authMW *auth.Middleware) {
	//only admin and user see all documantion versions
	mux.Handle("GET /api/documentversion", authMW.Authenticate(http.HandlerFunc(handler.GetDocumentVersionHandler)))
	mux.Handle("GET /api/documentversion/{id}/download", authMW.Authenticate(http.HandlerFunc(handler.DownloadDocumentVersionHandler)))

}
