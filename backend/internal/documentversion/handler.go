package documentversion

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/storage"
	"github.com/jackc/pgx/v5"
)

/* ================================================================== */
/*  Swagger helper types                                               */
/* ================================================================== */

// VersionResponse is a single version entry with fresh Cloudinary URLs.
type VersionResponse struct {
	ID               string `json:"id"                example:"2f0c8e5b-..."`
	DocumentID       int64  `json:"document_id"       example:"12"`
	UserID           int64  `json:"user_id"           example:"3"`
	Version          int    `json:"version"           example:"4"`
	OriginalFilename string `json:"original_filename" example:"Contract-2026.pdf"`
	URL              string `json:"url"               example:"https://res.cloudinary.com/..."`
	PreviewURL       string `json:"preview_url"       example:"https://res.cloudinary.com/..."`
	HashedString     string `json:"hashed_string"     example:"a1b2c3..."`
	CreatedAt        string `json:"created_at"        example:"2026-09-21T14:48:32Z"`
}

// VersionDownloadResponse is returned by DownloadVersion when JSON is requested.
type VersionDownloadResponse struct {
	URL      string `json:"url"      example:"https://res.cloudinary.com/..."`
	Filename string `json:"filename" example:"Contract-2026.pdf"`
}

// VersionRestoreResponse is returned by RestoreVersion.
type VersionRestoreResponse struct {
	Message string `json:"message" example:"Version restored"`
	Version int    `json:"version" example:"4"`
}

// VersionMessageResponse is a simple {message:"..."} envelope.
type VersionMessageResponse struct {
	Message string `json:"message" example:"Version deleted"`
}

/* ================================================================== */
/*  Handler                                                            */
/* ================================================================== */

type Handler struct {
	Service *Service
}

func NewDocumentVersionHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

/* ================================================================== */
/*  Helpers                                                            */
/* ================================================================== */

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	http.Error(w, msg, status)
}

/* ================================================================== */
/*  GET /api/documents/{id}/versions                                   */
/* ================================================================== */

// ListVersionsForDocument returns every version of a document.
//
// @Summary      List versions of a document
// @Description  Returns all versions newest first, each with fresh Cloudinary URLs.
// @Tags         document-versions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "Document ID"
// @Success      200  {array}   documentversion.VersionResponse
// @Failure      400  {object}  auth.ErrorResponse  "Invalid document ID"
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  auth.ErrorResponse  "Internal server error"
// @Router       /documents/{id}/versions [get]
func (h *Handler) ListVersionsForDocument(w http.ResponseWriter, r *http.Request) {
	docID, err := parsePathInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	requester, err := currentUser(r, h)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	_ = requester

	versions, err := h.Service.GetDocumentVersionsByDocumentID(
		r.Context(), h.Service.Repo.DB, docID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]VersionResponse, 0, len(versions))
	for _, v := range versions {
		downloadURL, _ := storage.GetFileURL(v.PublicID, true, false)
		previewURL, _ := storage.GetFileURL(v.PublicID, false, false)

		out = append(out, VersionResponse{
			ID:               v.ID,
			DocumentID:       v.DocumentID,
			UserID:           v.UserID,
			Version:          v.Version,
			OriginalFilename: v.OriginalFilename,
			URL:              downloadURL,
			PreviewURL:       previewURL,
			HashedString:     v.HashedString,
			CreatedAt:        v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeJSON(w, http.StatusOK, out)
}

/* ================================================================== */
/*  GET /api/versions                                                  */
/* ================================================================== */

// GetAllVersions lists every version in the system.
//
// @Summary      List all versions (admin only)
// @Tags         document-versions
// @Produce      json
// @Security     BearerAuth
// @Success 200 {array} documentversion.VersionResponse
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden — admin only"
// @Failure      500  {object}  auth.ErrorResponse  "Internal server error"
// @Router       /versions [get]
func (h *Handler) GetAllVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := h.Service.GetAllVersions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, versions)
}

/* ================================================================== */
/*  GET /api/versions/{id}                                             */
/* ================================================================== */

// GetVersion returns one version's metadata.
//
// @Summary      Get a version
// @Tags         document-versions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Version UUID"
// @Success      200  {object}  documentversion.VersionResponse
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      404  {object}  auth.ErrorResponse  "Version not found"
// @Failure      500  {object}  auth.ErrorResponse  "Internal server error"
// @Router       /versions/{id} [get]
func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id")

	v, err := h.Service.GetVersionByID(r.Context(), versionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Version not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	downloadURL, _ := storage.GetFileURL(v.PublicID, true, false)
	previewURL, _ := storage.GetFileURL(v.PublicID, false, false)

	writeJSON(w, http.StatusOK, VersionResponse{
		ID:               v.ID,
		DocumentID:       v.DocumentID,
		UserID:           v.UserID,
		Version:          v.Version,
		OriginalFilename: v.OriginalFilename,
		URL:              downloadURL,
		PreviewURL:       previewURL,
		HashedString:     v.HashedString,
		CreatedAt:        v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

/* ================================================================== */
/*  GET /api/versions/{id}/download                                    */
/* ================================================================== */

// DownloadVersion returns a download URL for a version.
//
// @Summary      Download a version
// @Description  Returns JSON with url + filename. The frontend triggers the download.
// @Tags         document-versions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Version UUID"
// @Success      200  {object}  documentversion.VersionDownloadResponse
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      404  {object}  auth.ErrorResponse  "Version not found"
// @Failure      500  {object}  auth.ErrorResponse  "Internal server error"
// @Router       /versions/{id}/download [get]
func (h *Handler) DownloadVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id")

	v, err := h.Service.GetVersionByID(r.Context(), versionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Version not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	url, err := storage.GetFileURL(v.PublicID, true, false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to build URL")
		return
	}

	// JSON only — the previous version called http.Redirect and then wrote JSON,
	// which produced a "superfluous WriteHeader" warning and dropped the body.
	writeJSON(w, http.StatusOK, VersionDownloadResponse{
		URL:      url,
		Filename: v.OriginalFilename,
	})
}

/* ================================================================== */
/*  POST /api/versions/{id}/restore                                    */
/* ================================================================== */

// RestoreVersion re-points the master document to an old version.
//
// @Summary      Restore an old version
// @Description  Makes the given version the current one.
// @Tags         document-versions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Version UUID"
// @Success      200  {object}  documentversion.VersionRestoreResponse
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      404  {object}  auth.ErrorResponse  "Version not found"
// @Failure      500  {object}  auth.ErrorResponse  "Internal server error"
// @Router       /versions/{id}/restore [post]
func (h *Handler) RestoreVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id")

	v, err := h.Service.GetVersionByID(r.Context(), versionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Version not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.Service.RestoreVersion(r.Context(), v.DocumentID, v.Version, v.PublicID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, VersionRestoreResponse{
		Message: "Version restored",
		Version: v.Version,
	})
}

/* ================================================================== */
/*  DELETE /api/versions/{id}                                          */
/* ================================================================== */

// DeleteVersion removes a single version.
//
// @Summary      Delete a version (admin only)
// @Tags         document-versions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Version UUID"
// @Success      200  {object}  documentversion.VersionMessageResponse
// @Failure      401  {object}  auth.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  auth.ErrorResponse  "Forbidden — admin only"
// @Failure      500  {object}  auth.ErrorResponse  "Internal server error"
// @Router       /versions/{id} [delete]
func (h *Handler) DeleteVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id")

	if err := h.Service.DeleteVersion(r.Context(), versionID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, VersionMessageResponse{
		Message: "Version deleted",
	})
}

/* ================================================================== */
/*  Local helpers                                                      */
/* ================================================================== */

func parsePathInt64(r *http.Request, name string) (int64, error) {
	s := r.PathValue(name)
	return strconv.ParseInt(s, 10, 64)
}

func currentUser(r *http.Request, h *Handler) (int64, error) {
	return auth.GetUserIDFromContext(r.Context())
}

// keep fmt import used if you wire logging later
var _ = fmt.Sprintf
