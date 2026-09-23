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
/*  Swagger helper types                                               */
/* ================================================================== */

// MessageResponse is a simple {message: "..."} envelope.
type MessageResponse struct {
	Message string `json:"message" example:"Document deleted"`
}

// UploadResponse is returned by Upload.
type UploadResponse struct {
	Message  string      `json:"message"  example:"File uploaded successfully"`
	Document interface{} `json:"document"`
}

// RenameDocumentRequest is the body for PATCH /documents/{id}.
type RenameDocumentRequest struct {
	Name string `json:"name" example:"Contract-2026.pdf"`
}

// MoveDocumentRequest is the body for PATCH /documents/{id}/move.
type MoveDocumentRequest struct {
	FolderID int64 `json:"folder_id" example:"3"`
}

// UpdateStatusRequest is the body for PATCH /documents/{id}/status.
type UpdateStatusRequest struct {
	Status  string `json:"status"  example:"approved" enums:"approved,rejected"`
	Comment string `json:"comment" example:"Looks good"`
}

// DocumentVersionResponse is a single version entry.
type DocumentVersionResponse struct {
	ID           string `json:"id"                example:"2f0c8e5b-..."`
	Version      int    `json:"version"           example:"3"`
	Filename     string `json:"original_filename" example:"Contract-2026.pdf"`
	URL          string `json:"url"               example:"https://res.cloudinary.com/..."`
	PreviewURL   string `json:"preview_url"       example:"https://res.cloudinary.com/..."`
	HashedString string `json:"hashed_string"     example:"a1b2c3..."`
	CreatedAt    string `json:"created_at"        example:"2026-09-21T14:48:32Z"`
	IsCurrent    bool   `json:"is_current"        example:"true"`
}

// VersionDownloadResponse is returned by DownloadVersion.
type VersionDownloadResponse struct {
	URL      string `json:"url"      example:"https://res.cloudinary.com/..."`
	Filename string `json:"filename" example:"Contract-2026.pdf"`
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

func (h *Handler) canAccess(
	ctx context.Context,
	docID, userID int64,
	action string,
) bool {
	if ownerID, err := h.getDocumentOwner(ctx, docID); err == nil && ownerID == userID {
		return true
	}

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
		return false
	}
	ok, err := h.shares.CanAccess(ctx, docID, userID, required)
	return err == nil && ok
}

/* ================================================================== */
/*  Upload                                                             */
/* ================================================================== */

// Upload handles POST /folders/{folderId}/documents
//
// @Summary      Upload a document
// @Description  Uploads a new file into the given folder and creates version 1.
// @Tags         documents
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        folderId     path      int     true   "Folder ID"
// @Param        file         formData  file    true   "File to upload"
// @Param        description  formData  string  false  "Optional description"
// @Success      200  {object}  UploadResponse
// @Failure      400  {object}  auth.ErrorResponse  "Invalid folder ID or missing file"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Router       /folders/{folderId}/documents [post]
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

// List handles GET /folders/{folderId}/documents
//
// @Summary      List documents in a folder
// @Tags         documents
// @Produce      json
// @Security     BearerAuth
// @Param        folderId  path  int  true  "Folder ID"
// @Success      200  {array}   document.Document
// @Failure      400  {object}  auth.ErrorResponse  "Invalid folder ID"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  auth.ErrorResponse  "Internal server error"
// @Router       /folders/{folderId}/documents [get]
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

// Download handles GET /documents/{id}
//
// @Summary      Download the latest version
// @Description  Redirects (302) to a Cloudinary URL with fl_attachment.
// @Tags         documents
// @Security     BearerAuth
// @Param        id  path  int  true  "Document ID"
// @Success      302  "Redirect to the Cloudinary download URL"
// @Failure      400  {object}  auth.ErrorResponse  "Invalid document ID"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Failure      500  {object}  auth.ErrorResponse  "Failed to generate URL"
// @Router       /documents/{id} [get]
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

// Delete handles DELETE /documents/{id}
//
// @Summary      Delete a document
// @Description  Deletes the document and every version of it.
// @Tags         documents
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "Document ID"
// @Success      200  {object}  MessageResponse
// @Failure      400  {object}  auth.ErrorResponse  "Invalid document ID"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Router       /documents/{id} [delete]
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

// Rename handles PATCH /documents/{id}
//
// @Summary      Rename a document
// @Description  Updates the display name. File content is unchanged.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  int                    true  "Document ID"
// @Param        body  body  RenameDocumentRequest  true  "New name"
// @Success      200  {object}  MessageResponse
// @Failure      400  {object}  auth.ErrorResponse  "Invalid request"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Router       /documents/{id} [patch]
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

// Move handles PATCH /documents/{id}/move
//
// @Summary      Move a document to another folder
// @Tags         documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  int                  true  "Document ID"
// @Param        body  body  MoveDocumentRequest  true  "Target folder"
// @Success      200  {object}  MessageResponse
// @Failure      400  {object}  auth.ErrorResponse  "Invalid request"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Router       /documents/{id}/move [patch]
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

// Search handles GET /documents?search=query&folder_id=1
//
// @Summary      Search documents
// @Description  Search by name or description, optionally limited to a folder.
// @Tags         documents
// @Produce      json
// @Security     BearerAuth
// @Param        search     query  string  false  "Search term"  example(contract)
// @Param        folder_id  query  int     false  "Limit to a folder"
// @Success      200  {array}   document.Document
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  auth.ErrorResponse  "Internal server error"
// @Router       /documents [get]
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

// UpdateDocumentContentHandler handles PUT /documents/{id}/content
//
// @Summary      Upload a new version
// @Description  Creates a new version of the document (v2, v3, ...).
// @Tags         documents
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id        path      int   true  "Document ID"
// @Param        document  formData  file  true  "New file content"
// @Success      200  {object}  MessageResponse
// @Failure      400  {object}  auth.ErrorResponse  "Invalid request"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Router       /documents/{id}/content [put]
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

// UpdateStatusHandler handles PATCH /documents/{id}/status
//
// @Summary      Approve or reject a document
// @Description  Only managers and admins can call this. Comment is optional.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  int                  true  "Document ID"
// @Param        body  body  UpdateStatusRequest  true  "Decision + comment"
// @Success      200  {object}  MessageResponse
// @Failure      400  {object}  auth.ErrorResponse  "Invalid request"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Router       /documents/{id}/status [patch]
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

// ListVersions handles GET /documents/{id}/versions
//
// @Summary      List every version of a document
// @Description  Returns versions newest first, each with a download URL.
// @Tags         document-versions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "Document ID"
// @Success      200  {array}   document.DocumentVersionResponse
// @Failure      400  {object}  auth.ErrorResponse  "Invalid document ID"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Router       /documents/{id}/versions [get]
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

// DownloadVersion handles GET /versions/{id}/download
//
// @Summary      Get a download URL for a specific version
// @Description  Returns JSON with url + filename. The frontend triggers the download.
// @Tags         document-versions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Version UUID"
// @Success      200  {object}  VersionDownloadResponse
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Failure      404  {object}  auth.ErrorResponse  "Version not found"
// @Router       /versions/{id}/download [get]
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":      url,
		"filename": v.OriginalFilename,
	})
}

/* ================================================================== */
/*  Restore a version                                                  */
/* ================================================================== */

// RestoreVersion handles POST /versions/{id}/restore
//
// @Summary      Restore an old version
// @Description  Re-points the master document to the given version.
// @Tags         document-versions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Version UUID"
// @Success      200  {object}  MessageResponse
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden"
// @Failure      404  {object}  auth.ErrorResponse  "Version not found"
// @Failure      500  {object}  auth.ErrorResponse  "Internal server error"
// @Router       /versions/{id}/restore [post]
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
