package document

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	auditlog "github.com/docmanage_new/internal/auditLog"
	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/documentversion"
	"github.com/docmanage_new/internal/storage"
	"github.com/docmanage_new/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ================================================================== */
/*  Interfaces                                                         */
/* ================================================================== */

// AuditLogger is satisfied by auditlog.Service.
type AuditLogger interface {
	Log(ctx context.Context, event string) error
}

// ShareChecker is satisfied by share.Service.
// Defined here to avoid an import cycle between document and share.
type ShareChecker interface {
	CanAccess(ctx context.Context, documentID, userID int64, action string) (bool, error)
}

/* ================================================================== */
/*  Handler                                                            */
/* ================================================================== */

type Handler struct {
	Service *Service
	audit   AuditLogger
	shares  ShareChecker
}

func NewHandler(
	db *pgxpool.Pool,
	versions *documentversion.Service,
	audit AuditLogger,
	shares ShareChecker,
) *Handler {
	return &Handler{
		Service: NewService(db, versions),
		audit:   audit,
		shares:  shares,
	}
}

/* ================================================================== */
/*  Helpers                                                            */
/* ================================================================== */

func (h *Handler) currentUser(r *http.Request) (*user.User, error) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		return nil, err
	}
	return user.GetUserByID(h.Service.DB, userID)
}

// getDocumentOwner returns the user_id stored on the documents row.
func (h *Handler) getDocumentOwner(ctx context.Context, docID int64) (int64, error) {
	var ownerID int64
	err := h.Service.DB.QueryRow(ctx,
		`SELECT created_by FROM documents WHERE id = $1`, docID,
	).Scan(&ownerID)
	if err != nil {
		return 0, fmt.Errorf("get document owner: %w", err)
	}
	return ownerID, nil
}

// canAccess reports whether the user may perform `action` on the document.
//
// action: "view", "download", "edit", "delete", "share"
//
// Ownership always grants full access. Otherwise the share table decides.
func (h *Handler) canAccess(
	ctx context.Context,
	docID, userID int64,
	action string,
) bool {
	// 1. Owner bypass
	if ownerID, err := h.getDocumentOwner(ctx, docID); err == nil && ownerID == userID {
		return true
	}

	// 2. Share table
	required := map[string]string{
		"view":     "viewer",
		"download": "viewer",
		"edit":     "editor",
		"delete":   "editor",
		"share":    "editor",
	}[action]
	if required == "" {
		return false
	}

	if h.shares == nil {
		// Share module not wired — deny by default.
		return false
	}
	ok, err := h.shares.CanAccess(ctx, docID, userID, required)
	return err == nil && ok
}

/* ================================================================== */
/*  Upload                                                             */
/* ================================================================== */

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	folderID, err := strconv.ParseInt(r.PathValue("folderId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large or invalid", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	description := r.FormValue("description")

	doc, err := h.Service.UploadDocument(
		r.Context(), folderID, currentUser, header, file, description,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentCreate); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":  "File uploaded successfully",
		"document": doc,
	})
}

/* ================================================================== */
/*  List                                                               */
/* ================================================================== */

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	folderID, err := strconv.ParseInt(r.PathValue("folderId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	docs, err := h.Service.ListDocuments(r.Context(), folderID, currentUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentView); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

/* ================================================================== */
/*  Download                                                           */
/* ================================================================== */

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.canAccess(r.Context(), docID, userID, "download") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	doc, err := h.Service.DownloadDocument(r.Context(), docID, currentUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentDownload); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	url, err := storage.GetFileURL(doc.PublicID, true, true)
	if err != nil {
		http.Error(w, "Failed to generate download URL", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

/* ================================================================== */
/*  Delete                                                             */
/* ================================================================== */

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.canAccess(r.Context(), docID, userID, "delete") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.Service.DeleteDocument(r.Context(), docID, currentUser); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentDelete); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document deleted"})
}

/* ================================================================== */
/*  Rename                                                             */
/* ================================================================== */

func (h *Handler) Rename(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.canAccess(r.Context(), docID, userID, "edit") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if err := h.Service.RenameDocument(r.Context(), docID, req.Name, currentUser); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentRename); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document renamed"})
}

/* ================================================================== */
/*  Move                                                               */
/* ================================================================== */

func (h *Handler) Move(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.canAccess(r.Context(), docID, userID, "edit") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FolderID int64 `json:"folder_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.FolderID <= 0 {
		http.Error(w, "Folder ID is required", http.StatusBadRequest)
		return
	}

	if err := h.Service.MoveDocument(r.Context(), docID, req.FolderID, currentUser); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentMove); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document moved"})
}

/* ================================================================== */
/*  Search                                                             */
/* ================================================================== */

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("search")

	var folderID *int64
	if s := r.URL.Query().Get("folder_id"); s != "" {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			folderID = &id
		}
	}

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	docs, err := h.Service.SearchDocuments(r.Context(), searchTerm, folderID, currentUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentSearch); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

/* ================================================================== */
/*  Update content                                                     */
/* ================================================================== */

func (h *Handler) UpdateDocumentContentHandler(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if !h.canAccess(r.Context(), docID, userID, "edit") {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "Missing 'document' file in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	userRole, _ := auth.GetRoleFromContext(r.Context())
	requester := &user.User{ID: userID, Role: userRole}

	if err := h.Service.UpdateDocumentContent(r.Context(), docID, file, header, requester); err != nil {
		if err.Error() == "permission denied" {
			http.Error(w, `{"error":"Permission denied"}`, http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentUpdate); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Document updated successfully"}`))
}

/* ================================================================== */
/*  Status workflow                                                    */
/* ================================================================== */

func (h *Handler) UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if !h.canAccess(r.Context(), docID, userID, "edit") {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	role, err := auth.GetRoleFromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	requester := &user.User{ID: userID, Role: role}

	var req struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateStatus(r.Context(), docID, req.Status, requester, req.Comment); err != nil {
		if err.Error() == "permission denied" {
			http.Error(w, `{"error":"Permission denied"}`, http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentStatusUpdate); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Document status and comment updated successfully"}`))
}

/* ================================================================== */
/*  Version history                                                    */
/* ================================================================== */

func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.canAccess(r.Context(), docID, userID, "view") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	doc, err := h.Service.GetDocument(r.Context(), docID, currentUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	versions, err := h.Service.Versions.GetDocumentVersionsByDocumentID(
		r.Context(), h.Service.DB, docID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type versionResponse struct {
		ID           string `json:"id"`
		Version      int    `json:"version"`
		Filename     string `json:"original_filename"`
		URL          string `json:"url"`
		PreviewURL   string `json:"preview_url"`
		HashedString string `json:"hashed_string"`
		CreatedAt    string `json:"created_at"`
		IsCurrent    bool   `json:"is_current"`
	}

	out := make([]versionResponse, 0, len(versions))
	for _, v := range versions {
		downloadURL, _ := storage.GetFileURL(v.PublicID, true, false)
		previewURL, _ := storage.GetFileURL(v.PublicID, false, false)

		out = append(out, versionResponse{
			ID:           v.ID,
			Version:      v.Version,
			Filename:     v.OriginalFilename,
			URL:          downloadURL,
			PreviewURL:   previewURL,
			HashedString: v.HashedString,
			CreatedAt:    v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			IsCurrent:    v.Version == doc.Version,
		})
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentVersionList); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

/* ================================================================== */
/*  Download a version                                                 */
/* ================================================================== */

func (h *Handler) DownloadVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id")

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	v, err := h.Service.Versions.GetVersionByID(r.Context(), versionID)
	if err != nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	if !h.canAccess(r.Context(), v.DocumentID, userID, "download") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	url, err := storage.GetFileURL(v.PublicID, true, false)
	if err != nil {
		http.Error(w, "Failed to generate URL", http.StatusInternalServerError)
		return
	}

	// JSON response — the frontend uses this to trigger the browser download.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":      url,
		"filename": v.OriginalFilename,
	})
}

/* ================================================================== */
/*  Restore a version                                                  */
/* ================================================================== */

func (h *Handler) RestoreVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id")

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	v, err := h.Service.Versions.GetVersionByID(r.Context(), versionID)
	if err != nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	if !h.canAccess(r.Context(), v.DocumentID, currentUser.ID, "edit") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	doc, err := h.Service.GetDocument(r.Context(), v.DocumentID, currentUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if _, err := h.Service.DB.Exec(r.Context(),
		`UPDATE documents SET version = $1, file_path = $2 WHERE id = $3`,
		v.Version, v.PublicID, doc.ID,
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.audit.Log(r.Context(), auditlog.EventDocumentVersionRestore); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Version restored"})
}
