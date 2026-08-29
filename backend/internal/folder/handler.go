package folder

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/docmanage_new/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		writeJSONError(w, "Unauthorized context", http.StatusUnauthorized)
		return
	}

	var req CreateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	folder, err := h.service.CreateFolder(r.Context(), req.Name, userID)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(folder)
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	folders, err := h.service.ListFolders(r.Context())
	if err != nil {
		writeJSONError(w, "Failed to retrieve folders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(folders)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteFolder(r.Context(), id); err != nil {
		writeJSONError(w, "Failed to delete folder", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Folder deleted"})
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
