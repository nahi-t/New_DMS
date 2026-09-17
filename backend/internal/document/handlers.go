// package document

// import (
// 	"encoding/json"
// 	"net/http"
// 	"strconv"

// 	"github.com/docmanage_new/internal/auth"
// 	"github.com/docmanage_new/internal/user"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// // Handler holds dependencies (service) for HTTP handlers.
// type Handler struct {
// 	Service *Service
// }

// // NewHandler creates a new document handler with its service.
// func NewHandler(db *pgxpool.Pool) *Handler {
// 	return &Handler{
// 		Service: NewService(db),
// 	}
// }

// // Upload handles POST /folders/{folderId}/documents
// func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
// 	folderIDStr := r.PathValue("folderId")
// 	folderID, err := strconv.ParseInt(folderIDStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
// 		return
// 	}

// 	userID, err := auth.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUser, err := user.GetUserByID(h.Service.DB, userID) // still need DB for user fetch
// 	if err != nil {
// 		http.Error(w, "User not found", http.StatusUnauthorized)
// 		return
// 	}

// 	if err := r.ParseMultipartForm(10 << 20); err != nil {
// 		http.Error(w, "File too large or invalid", http.StatusBadRequest)
// 		return
// 	}
// 	file, header, err := r.FormFile("file")
// 	if err != nil {
// 		http.Error(w, "Missing file", http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	description := r.FormValue("description")

// 	doc, err := h.Service.UploadDocument(r.Context(), folderID, currentUser, header, file, description)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusForbidden)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"message":  "File uploaded successfully",
// 		"document": doc,
// 	})
// }

// // List handles GET /folders/{folderId}/documents
// func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
// 	folderIDStr := r.PathValue("folderId")
// 	folderID, err := strconv.ParseInt(folderIDStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
// 		return
// 	}

// 	userID, err := auth.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUser, err := user.GetUserByID(h.Service.DB, userID)
// 	if err != nil {
// 		http.Error(w, "User not found", http.StatusUnauthorized)
// 		return
// 	}

// 	docs, err := h.Service.ListDocuments(r.Context(), folderID, currentUser)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(docs)
// }

// // Download handles GET /documents/{id}
// func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
// 	docIDStr := r.PathValue("id")
// 	docID, err := strconv.ParseInt(docIDStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid document ID", http.StatusBadRequest)
// 		return
// 	}

// 	userID, err := auth.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUser, err := user.GetUserByID(h.Service.DB, userID)
// 	if err != nil {
// 		http.Error(w, "User not found", http.StatusUnauthorized)
// 		return
// 	}

// 	doc, err := h.Service.DownloadDocument(r.Context(), docID, currentUser)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusForbidden)
// 		return
// 	}

// 	http.ServeFile(w, r, doc.FilePath)
// }

// // Delete handles DELETE /documents/{id}
// func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
// 	docIDStr := r.PathValue("id")
// 	docID, err := strconv.ParseInt(docIDStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid document ID", http.StatusBadRequest)
// 		return
// 	}

// 	userID, err := auth.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUser, err := user.GetUserByID(h.Service.DB, userID)
// 	if err != nil {
// 		http.Error(w, "User not found", http.StatusUnauthorized)
// 		return
// 	}

// 	if err := h.Service.DeleteDocument(r.Context(), docID, currentUser); err != nil {
// 		http.Error(w, err.Error(), http.StatusForbidden)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(map[string]string{"message": "Document deleted"})
// }

// // RenameHandler handles PATCH /documents/{id} – rename document.
// func (h *Handler) Rename(w http.ResponseWriter, r *http.Request) {
// 	docIDStr := r.PathValue("id")
// 	docID, err := strconv.ParseInt(docIDStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid document ID", http.StatusBadRequest)
// 		return
// 	}

// 	userID, err := auth.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUser, err := user.GetUserByID(h.Service.DB, userID)
// 	if err != nil {
// 		http.Error(w, "User not found", http.StatusUnauthorized)
// 		return
// 	}

// 	var req struct {
// 		Name string `json:"name"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}
// 	if req.Name == "" {
// 		http.Error(w, "Name is required", http.StatusBadRequest)
// 		return
// 	}

// 	if err := h.Service.RenameDocument(r.Context(), docID, req.Name, currentUser); err != nil {
// 		http.Error(w, err.Error(), http.StatusForbidden)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(map[string]string{"message": "Document renamed"})
// }

// // MoveHandler handles PATCH /documents/{id}/move – move document to another folder.
// func (h *Handler) Move(w http.ResponseWriter, r *http.Request) {
// 	docIDStr := r.PathValue("id")
// 	docID, err := strconv.ParseInt(docIDStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid document ID", http.StatusBadRequest)
// 		return
// 	}

// 	userID, err := auth.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUser, err := user.GetUserByID(h.Service.DB, userID)
// 	if err != nil {
// 		http.Error(w, "User not found", http.StatusUnauthorized)
// 		return
// 	}

// 	var req struct {
// 		FolderID int64 `json:"folder_id"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}
// 	if req.FolderID <= 0 {
// 		http.Error(w, "Folder ID is required", http.StatusBadRequest)
// 		return
// 	}

// 	if err := h.Service.MoveDocument(r.Context(), docID, req.FolderID, currentUser); err != nil {
// 		http.Error(w, err.Error(), http.StatusForbidden)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(map[string]string{"message": "Document moved"})
// }

// // SearchHandler handles GET /documents?search=query&folder_id=1
// func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
// 	searchTerm := r.URL.Query().Get("search")
// 	folderIDStr := r.URL.Query().Get("folder_id")
// 	var folderID *int64
// 	if folderIDStr != "" {
// 		id, err := strconv.ParseInt(folderIDStr, 10, 64)
// 		if err == nil {
// 			folderID = &id
// 		}
// 	}

// 	userID, err := auth.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUser, err := user.GetUserByID(h.Service.DB, userID)
// 	if err != nil {
// 		http.Error(w, "User not found", http.StatusUnauthorized)
// 		return
// 	}

// 	docs, err := h.Service.SearchDocuments(r.Context(), searchTerm, folderID, currentUser)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(docs)
// }

// func (h *Handler) UpdateDocumentContentHandler(w http.ResponseWriter, r *http.Request) {

// 	docIDStr := r.PathValue("id")
// 	docID, err := strconv.ParseInt(docIDStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid document ID", http.StatusBadRequest)
// 		return
// 	}

// 	if err := r.ParseMultipartForm(10 << 20); err != nil {
// 		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
// 		return
// 	}

// 	file, header, err := r.FormFile("document")
// 	if err != nil {
// 		http.Error(w, "Missing 'document' file in request", http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	// 🔑 USE YOUR AUTH HELPER FUNCTIONS HERE INSTEAD OF "requester"
// 	userID, err := auth.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
// 		return
// 	}

// 	userRole, _ := auth.GetRoleFromContext(r.Context())

// 	// Call service method with extracted ID and role
// 	err = h.Service.updateDocumentContent(r.Context(), docID, file, header, userID, userRole)
// 	if err != nil {
// 		if err.Error() == "permission denied" {
// 			http.Error(w, `{"error": "Permission denied"}`, http.StatusForbidden)
// 			return
// 		}
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"message": "Document updated successfully"}`))
// }

// func (h *Handler) UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
// 	idStr := r.PathValue("id")
// 	docID, err := strconv.ParseInt(idStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid document ID", http.StatusBadRequest)
// 		return
// 	}

// 	// --- 1. Retrieve authenticated user ID and role from context ---
// 	userID, err := auth.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
// 		return
// 	}
// 	role, err := auth.GetRoleFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
// 		return
// 	}

// 	// --- 2. Build the user object (must match the type the service expects) ---
// 	requester := &user.User{
// 		ID:   userID,
// 		Role: role,
// 	}

// 	// --- 3. Decode request body ---
// 	var req struct {
// 		Status  string `json:"status"`
// 		Comment string `json:"comment"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// --- 4. Call service with the valid requester ---
// 	err = h.Service.UpdateStatus(r.Context(), docID, req.Status, requester, req.Comment)
// 	if err != nil {
// 		if err.Error() == "permission denied" {
// 			http.Error(w, `{"error": "Permission denied"}`, http.StatusForbidden)
// 			return
// 		}
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"message": "Document status and comment updated successfully"}`))
// }

package document

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/documentversion"
	"github.com/docmanage_new/internal/storage"
	"github.com/docmanage_new/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ================================================================== */
/*  Handler                                                            */
/* ================================================================== */

// Handler holds dependencies (service) for HTTP handlers.
type Handler struct {
	Service *Service
}

// NewHandler wires up the document service with its dependencies.
func NewHandler(db *pgxpool.Pool, versions *documentversion.Service) *Handler {
	return &Handler{
		Service: NewService(db, versions),
	}
}

// helper — fetch the authenticated user object from the request context.
func (h *Handler) currentUser(r *http.Request) (*user.User, error) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		return nil, err
	}
	return user.GetUserByID(h.Service.DB, userID)
}

/* ================================================================== */
/*  Upload                                                             */
/* ================================================================== */

// Upload handles POST /folders/{folderId}/documents
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

/* ================================================================== */
/*  Download — REDIRECT to Cloudinary                                  */
/* ================================================================== */

// Download handles GET /documents/{id}
//
// With Cloudinary, we no longer serve the file from disk — we build the
// secure download URL from the stored public_id and redirect the browser.
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
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

	// doc.FilePath now holds a Cloudinary public_id.
	// Build a download URL (fl_attachment forces the browser to download
	// rather than display inline).
	url, err := storage.GetFileURL(doc.PublicID, true, true)
	if err != nil {
		http.Error(w, "Failed to generate download URL", http.StatusInternalServerError)
		return
	}
	fmt.Println("Generated download URL:", url)

	// Redirect the client to Cloudinary. The browser handles the download.
	http.Redirect(w, r, url, http.StatusFound)
}

/* ================================================================== */
/*  Delete                                                             */
/* ================================================================== */

// Delete handles DELETE /documents/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document deleted"})
}

/* ================================================================== */
/*  Rename                                                             */
/* ================================================================== */

// Rename handles PATCH /documents/{id}
func (h *Handler) Rename(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document renamed"})
}

/* ================================================================== */
/*  Move                                                               */
/* ================================================================== */

// Move handles PATCH /documents/{id}/move
func (h *Handler) Move(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document moved"})
}

/* ================================================================== */
/*  Search                                                             */
/* ================================================================== */

// Search handles GET /documents?search=query&folder_id=1
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

/* ================================================================== */
/*  Update content — creates a new version                             */
/* ================================================================== */

// UpdateDocumentContentHandler handles PUT /documents/{id}/content
func (h *Handler) UpdateDocumentContentHandler(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
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

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	userRole, _ := auth.GetRoleFromContext(r.Context())

	// Build a requester object so we use the same signature as other handlers.
	requester := &user.User{ID: userID, Role: userRole}

	err = h.Service.UpdateDocumentContent(r.Context(), docID, file, header, requester)
	if err != nil {
		if err.Error() == "permission denied" {
			http.Error(w, `{"error":"Permission denied"}`, http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Document updated successfully"}`))
}

/* ================================================================== */
/*  Status workflow                                                    */
/* ================================================================== */

// UpdateStatusHandler handles PATCH /documents/{id}/status
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

	err = h.Service.UpdateStatus(r.Context(), docID, req.Status, requester, req.Comment)
	if err != nil {
		if err.Error() == "permission denied" {
			http.Error(w, `{"error":"Permission denied"}`, http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Document status and comment updated successfully"}`))
}

/* ================================================================== */
/*  Version history (NEW)                                              */
/* ================================================================== */

// ListVersions handles GET /documents/{id}/versions
//
// Returns every version of the document, newest first, each with a fresh
// Cloudinary download URL.
func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	currentUser, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Permission check via the document service
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// DownloadVersion handles GET /versions/{id}/download
func (h *Handler) DownloadVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id")

	v, err := h.Service.Versions.GetVersionByID(r.Context(), versionID)
	if err != nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	url, err := storage.GetFileURL(v.PublicID, true, false)
	if err != nil {
		http.Error(w, "Failed to generate URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
	// 3. Return JSON — no redirect.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":      url,
		"filename": v.OriginalFilename,
	})
}

// RestoreVersion handles POST /versions/{id}/restore
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

	doc, err := h.Service.GetDocument(r.Context(), v.DocumentID, currentUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Re-point the master row to the old version's public_id.
	_, err = h.Service.DB.Exec(r.Context(),
		`UPDATE documents SET version = $1, file_path = $2 WHERE id = $3`,
		v.Version, v.PublicID, doc.ID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Version restored"})
}
