package auditlog

import (
	"net/http"

	"github.com/docmanage_new/internal/auth"
)

func SetUpAuditLogDisplay(mux *http.ServeMux, handler *AuditHandler, authMW *auth.Middleware) {
	mux.Handle("GET /api/auditlogs", authMW.Authenticate(authMW.RequireRole("admin")(http.HandlerFunc(handler.GetAuditLogs))))

}
