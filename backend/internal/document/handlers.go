package document

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler holds dependencies (service) for HTTP handlers.
type Handler struct {
	Service *Service
}

// NewHandler creates a new document handler with its service.
func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		Service: NewService(db),
	}
}

// Upload handles POST /folders/{folderId}/documents
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	folderIDStr := r.PathValue("folderId")
	folderID, err := strconv.ParseInt(folderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUser, err := user.GetUserByID(h.Service.DB, userID) // still need DB for user fetch
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
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

	doc, err := h.Service.UploadDocument(r.Context(), folderID, currentUser, header, file, description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "File uploaded successfully",
		"document": doc,
	})
}

// List handles GET /folders/{folderId}/documents
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	folderIDStr := r.PathValue("folderId")
	folderID, err := strconv.ParseInt(folderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUser, err := user.GetUserByID(h.Service.DB, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
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

// Download handles GET /documents/{id}
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	docIDStr := r.PathValue("id")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUser, err := user.GetUserByID(h.Service.DB, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	doc, err := h.Service.DownloadDocument(r.Context(), docID, currentUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, doc.FilePath)
}

// Delete handles DELETE /documents/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	docIDStr := r.PathValue("id")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUser, err := user.GetUserByID(h.Service.DB, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if err := h.Service.DeleteDocument(r.Context(), docID, currentUser); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Document deleted"})
}
