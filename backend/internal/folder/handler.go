package folder

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/docmanage_new/internal/auth"
)

/* ================================================================== */
/*  Swagger helper types                                               */
/* ================================================================== */

// CreateFolderRequest is the body for POST /folders.
type CreateFolderRequest struct {
	Name string `json:"name" example:"Contracts"`
}

// UpdateFolderRequest is the body for PUT /folders/{id}.
type UpdateFolderRequest struct {
	Name string `json:"name" example:"Contracts 2026"`
}

// MessageResponse is a {message: "..."} envelope.
type MessageResponse struct {
	Message string `json:"message" example:"Folder deleted"`
}

/* ================================================================== */
/*  Handler                                                            */
/* ================================================================== */

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

/* ================================================================== */
/*  POST /api/folders                                                  */
/* ================================================================== */

// Create creates a new folder.
//
// @Summary      Create a folder
// @Description  Creates a folder owned by the authenticated user.
// @Tags         folders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateFolderRequest  true  "Folder payload"
// @Success      201   {object}  folder.Folder
// @Failure      400   {object}  folder.ErrorResponse  "Invalid payload"
// @Failure      401   {object}  folder.ErrorResponse  "Unauthorized"
// @Router       /folders [post]
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

/* ================================================================== */
/*  GET /api/folders                                                   */
/* ================================================================== */

// GetAll lists every folder in the workspace.
//
// @Summary      List folders
// @Tags         folders
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   folder.Folder
// @Failure      401  {object}  folder.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  folder.ErrorResponse  "Failed to retrieve folders"
// @Router       /folders [get]
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

/* ================================================================== */
/*  PUT /api/folders/{id}                                              */
/* ================================================================== */

// Update renames an existing folder.
//
// @Summary      Rename a folder
// @Tags         folders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                  true  "Folder ID"
// @Param        body  body      UpdateFolderRequest  true  "New name"
// @Success      200   {object}  folder.Folder
// @Failure      400   {object}  folder.ErrorResponse  "Invalid folder ID or payload"
// @Failure      401   {object}  folder.ErrorResponse  "Unauthorized"
// @Failure      403   {object}  folder.ErrorResponse  "Forbidden"
// @Failure      404   {object}  folder.ErrorResponse  "Folder not found"
// @Router       /folders/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	var req UpdateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		writeJSONError(w, "Name is required", http.StatusBadRequest)
		return
	}
	name := req.Name

	updated, err := h.service.Update(r.Context(), id, UpdateFolderRequest{name})
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

/* ================================================================== */
/*  DELETE /api/folders/{id}                                           */
/* ================================================================== */

// Delete removes a folder and everything inside it.
//
// @Summary      Delete a folder
// @Description  Deletes the folder and every document it contains.
// @Tags         folders
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "Folder ID"
// @Success      200  {object}  folder.MessageResponse
// @Failure      400  {object}  folder.ErrorResponse  "Invalid folder ID"
// @Failure      401  {object}  folder.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  folder.ErrorResponse  "Failed to delete folder"
// @Router       /folders/{id} [delete]
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

/* ================================================================== */
/*  Error envelope                                                     */
/* ================================================================== */

// ErrorResponse is a generic error envelope.
type ErrorResponse struct {
	Error string `json:"error" example:"Failed to delete folder"`
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
