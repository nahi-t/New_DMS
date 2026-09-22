package share

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/httpx"
)

type Handler struct {
	service *Service
	audit   AuditLogger
}

func NewHandler(service *Service, audit AuditLogger) *Handler {
	return &Handler{service: service, audit: audit}
}

/* ------------------------------------------------------------------ */
/*  POST /api/documents/{id}/share                                     */
/* ------------------------------------------------------------------ */

func (h *Handler) ShareDocument(w http.ResponseWriter, r *http.Request) {
	actorID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.GetRoleFromContext(r.Context())

	docID, err := extractDocID(r.URL.Path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid document id")
		return
	}

	var req struct {
		UserID     int64  `json:"user_id"`
		Permission string `json:"permission"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	err = h.service.Share(r.Context(), docID, req.UserID, actorID, role, req.Permission)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidPermission):
			writeJSONError(w, http.StatusBadRequest, "permission must be viewer or editor")
		case errors.Is(err, ErrSelfShare):
			writeJSONError(w, http.StatusBadRequest, "cannot share with yourself")
		case errors.Is(err, ErrForbidden):
			writeJSONError(w, http.StatusForbidden, "forbidden")
		default:
			log.Printf("share error: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to share")
		}
		return
	}

	// Audit
	if aerr := h.audit.LogLogin(
		r.Context(),
		&actorID,
		"",
		"document.share",
		httpx.ClientIP(r),
		r.Header.Get("User-Agent"),
		true,
	); aerr != nil {
		log.Printf("audit: document.share failed: %v", aerr)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "shared"})
}

/* ------------------------------------------------------------------ */
/*  DELETE /api/documents/{id}/share/{userId}                          */
/* ------------------------------------------------------------------ */

func (h *Handler) UnshareDocument(w http.ResponseWriter, r *http.Request) {
	actorID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.GetRoleFromContext(r.Context())

	docID, err := extractDocID(r.URL.Path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid document id")
		return
	}

	targetUserID, err := extractUserIDFromPath(r.URL.Path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.service.Unshare(r.Context(), docID, targetUserID, actorID, role); err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			writeJSONError(w, http.StatusForbidden, "forbidden")
		default:
			log.Printf("unshare error: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to unshare")
		}
		return
	}

	if aerr := h.audit.LogLogin(
		r.Context(),
		&actorID,
		"",
		"document.unshare",
		httpx.ClientIP(r),
		r.Header.Get("User-Agent"),
		true,
	); aerr != nil {
		log.Printf("audit: document.unshare failed: %v", aerr)
	}

	w.WriteHeader(http.StatusNoContent)
}

/* ------------------------------------------------------------------ */
/*  GET /api/documents/{id}/shares                                     */
/* ------------------------------------------------------------------ */

func (h *Handler) ListShares(w http.ResponseWriter, r *http.Request) {
	actorID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.GetRoleFromContext(r.Context())

	docID, err := extractDocID(r.URL.Path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid document id")
		return
	}

	shares, err := h.service.ListShares(r.Context(), docID, actorID, role)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			writeJSONError(w, http.StatusForbidden, "forbidden")
			return
		}
		log.Printf("list shares error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to list shares")
		return
	}

	writeJSON(w, http.StatusOK, shares)
}

/* ------------------------------------------------------------------ */
/*  GET /api/shares/shared-with-me                                     */
/* ------------------------------------------------------------------ */

func (h *Handler) SharedWithMe(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	shares, err := h.service.GetSharedWithMe(r.Context(), userID)
	if err != nil {
		log.Printf("shared-with-me error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch shared documents")
		return
	}

	writeJSON(w, http.StatusOK, shares)
}

/* ------------------------------------------------------------------ */
/*  Helpers                                                            */
/* ------------------------------------------------------------------ */

// extractDocID parses "/api/documents/{id}/share" or "/api/documents/{id}/shares".
func extractDocID(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// parts = ["api", "documents", "<id>", "share"...]
	if len(parts) < 3 {
		return 0, fmt.Errorf("bad path: %s", path)
	}
	return strconv.ParseInt(parts[2], 10, 64)
}

// extractUserIDFromPath parses "/api/documents/{id}/share/{userId}".
func extractUserIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// parts = ["api", "documents", "<id>", "share", "<userId>"]
	if len(parts) < 5 {
		return 0, fmt.Errorf("bad path: %s", path)
	}
	return strconv.ParseInt(parts[4], 10, 64)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
