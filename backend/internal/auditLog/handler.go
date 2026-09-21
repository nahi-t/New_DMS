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

// GetAuditLogs returns a page of audit events.
//
// Query params:
//   - limit   (default 50, max 500)
//   - offset  (default 0)
//   - user_id (optional)
//   - event   (optional)
//   - from    (optional, RFC3339)
//   - to      (optional, RFC3339)
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
	resp := map[string]any{
		"data":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
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
