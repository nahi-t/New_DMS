package documentversion

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Handler struct {
	Service *Service
}

func NewDocumentVersionHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) GetDocumentVersionHandler(w http.ResponseWriter, r *http.Request) {
	// Read from query parameter ?document_id=123
	idStr := r.URL.Query().Get("document_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	versions, err := h.Service.GetDocumentVersionsByDocumentID(r.Context(), h.Service.Repo.DB, id)
	if err != nil {
		http.Error(w, "Failed to retrieve document versions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}
func (h *Handler) DownloadDocumentVersionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id") // This is the ID of the row in document_versions
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid version ID", http.StatusBadRequest)
		return
	}

	// Fetch file path and version number from document_versions
	var filePath string
	var versionNum int
	query := `SELECT file_path, version FROM document_versions WHERE id = $1`
	err = h.Service.Repo.DB.QueryRow(r.Context(), query, id).Scan(&filePath, &versionNum)
	if err != nil {
		http.Error(w, "Version file not found", http.StatusNotFound)
		return
	}

	// Send the file to the client
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"v%d_file\"", versionNum))
	http.ServeFile(w, r, filePath)
}
