package auditlog

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type AuditHandler struct {
	service *Service
}

func NewAuditHandler(service *Service) *AuditHandler {
	return &AuditHandler{service: service}
}

// GetAuditLogs returns a paginated list of audit events.
//
// @Summary      List audit logs
// @Description  Returns a page of audit events. Only admins can call this endpoint.
// @Description  Supports filtering by user, event type, and time range.
// @Tags         audit
// @Produce      json
// @Security     BearerAuth
// @Param        limit    query     int     false  "Page size (default 50, max 500)"  minimum(1)  maximum(500)
// @Param        offset   query     int     false  "Offset (default 0)"              minimum(0)
// @Param        user_id  query     int     false  "Filter by user ID"
// @Param        event    query     string  false  "Filter by event name (exact match)"  example(user.login)
// @Param        from     query     string  false  "Start time, RFC3339"  example(2026-09-01T00:00:00Z)
// @Param        to       query     string  false  "End time, RFC3339"    example(2026-09-30T23:59:59Z)
// @Success      200      {object}  AuditLogPage
// @Failure      400      {object}  auth.ErrorResponse  "Invalid query parameter"
// @Failure      401      {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403      {object}  auth.ErrorResponse  "Forbidden — admin only"
// @Failure      500      {object}  auth.ErrorResponse  "Internal server error"
// @Router       /auditlogs [get]
func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parse and clamp query params.
	limit := parseIntDefault(r.URL.Query().Get("limit"), 50)
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	if offset < 0 {
		offset = 0
	}

	filter := AuditFilter{
		Limit:  limit,
		Offset: offset,
		UserID: r.URL.Query().Get("user_id"),
		Event:  r.URL.Query().Get("event"),
	}

	if s := r.URL.Query().Get("from"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid 'from' timestamp")
			return
		}
		filter.From = &t
	}
	if s := r.URL.Query().Get("to"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid 'to' timestamp")
			return
		}
		filter.To = &t
	}

	// 2. Fetch.
	logs, total, err := h.service.GetAuditLogs(ctx, filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch audit logs")
		return
	}

	// 3. Marshal BEFORE writing the header so errors are still recoverable.
	resp := AuditLogPage{
		Data:   logs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	body, err := json.Marshal(resp)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to serialize logs")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
