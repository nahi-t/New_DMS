// package documentversion

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"strconv"
// )

// type Handler struct {
// 	Service *Service
// }

// func NewDocumentVersionHandler(service *Service) *Handler {
// 	return &Handler{Service: service}
// }

// func (h *Handler) GetDocumentVersionHandler(w http.ResponseWriter, r *http.Request) {
// 	// Read from query parameter ?document_id=123
// 	idStr := r.URL.Query().Get("document_id")
// 	id, err := strconv.ParseInt(idStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid document ID", http.StatusBadRequest)
// 		return
// 	}

// 	versions, err := h.Service.GetDocumentVersionsByDocumentID(r.Context(), h.Service.Repo.DB, id)
// 	if err != nil {
// 		http.Error(w, "Failed to retrieve document versions", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(versions)
// }
// func (h *Handler) DownloadDocumentVersionHandler(w http.ResponseWriter, r *http.Request) {
// 	idStr := r.PathValue("id") // This is the ID of the row in document_versions
// 	id, err := strconv.ParseInt(idStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid version ID", http.StatusBadRequest)
// 		return
// 	}

// 	// Fetch file path and version number from document_versions
// 	var filePath string
// 	var versionNum int
// 	query := `SELECT file_path, version FROM document_versions WHERE id = $1`
// 	err = h.Service.Repo.DB.QueryRow(r.Context(), query, id).Scan(&filePath, &versionNum)
// 	if err != nil {
// 		http.Error(w, "Version file not found", http.StatusNotFound)
// 		return
// 	}

//		// Send the file to the client
//		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"v%d_file\"", versionNum))
//		http.ServeFile(w, r, filePath)
//	}
package documentversion

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/storage"
	"github.com/jackc/pgx/v5"
)

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
/*  List every version of a document                                   */
/* ================================================================== */

func (h *Handler) ListVersionsForDocument(w http.ResponseWriter, r *http.Request) {
	docID, err := parsePathInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	// Permission: only owner / manager / admin (checked in document service).
	requester, err := currentUser(r, h)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	_ = requester // hook permission check here if you want

	versions, err := h.Service.GetDocumentVersionsByDocumentID(
		r.Context(), h.Service.Repo.DB, docID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Enrich each version with a fresh Cloudinary URL.
	type versionOut struct {
		ID               string `json:"id"`
		DocumentID       int64  `json:"document_id"`
		UserID           int64  `json:"user_id"`
		Version          int    `json:"version"`
		OriginalFilename string `json:"original_filename"`
		URL              string `json:"url"`         // forces download
		PreviewURL       string `json:"preview_url"` // inline
		HashedString     string `json:"hashed_string"`
		CreatedAt        string `json:"created_at"`
	}

	out := make([]versionOut, 0, len(versions))
	for _, v := range versions {
		downloadURL, _ := storage.GetFileURL(v.PublicID, true, false)
		previewURL, _ := storage.GetFileURL(v.PublicID, false, false)

		out = append(out, versionOut{
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
/*  List all versions (admin only)                                     */
/* ================================================================== */

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
/*  Get one version's metadata                                         */
/* ================================================================== */

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id") // UUID string

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

	writeJSON(w, http.StatusOK, map[string]any{
		"id":                v.ID,
		"document_id":       v.DocumentID,
		"user_id":           v.UserID,
		"version":           v.Version,
		"original_filename": v.OriginalFilename,
		"url":               downloadURL,
		"preview_url":       previewURL,
		"hashed_string":     v.HashedString,
		"created_at":        v.CreatedAt,
	})
}

/* ================================================================== */
/*  GET /api/versions/{id}/download                                    */
/*  Redirect to Cloudinary's download URL                              */
/* ================================================================== */

func (h *Handler) DownloadVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id") // UUID string

	v, err := h.Service.GetVersionByID(r.Context(), versionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Version not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build the Cloudinary download URL.
	url, err := storage.GetFileURL(v.PublicID, true, false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to build URL")
		return
	}

	// 302 redirect — the browser downloads from Cloudinary directly.
	http.Redirect(w, r, url, http.StatusFound)
	writeJSON(w, http.StatusOK, map[string]string{
		"url":      url,
		"filename": v.OriginalFilename,
	})
}

/* ================================================================== */
/*  POST /api/versions/{id}/restore                                    */
/*  Re-point the master document to an old version                     */
/* ================================================================== */

func (h *Handler) RestoreVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id")

	// 1. Fetch the version to restore
	v, err := h.Service.GetVersionByID(r.Context(), versionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Version not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2. Re-point the master document row to the old version's public_id
	err = h.Service.RestoreVersion(r.Context(), v.DocumentID, v.Version, v.PublicID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Version restored",
		"version": string(rune(v.Version)),
	})
}

/* ================================================================== */
/*  DELETE /api/versions/{id}  (admin only)                            */
/* ================================================================== */

func (h *Handler) DeleteVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("id")

	if err := h.Service.DeleteVersion(r.Context(), versionID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Version deleted",
	})
}

/* ================================================================== */
/*  Local helpers                                                      */
/* ================================================================== */

func parsePathInt64(r *http.Request, name string) (int64, error) {
	s := r.PathValue(name)
	return strconvParseInt(s)
}

func strconvParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// currentUser is optional — you can drop it if you don't need permission checks here.
func currentUser(r *http.Request, h *Handler) (int64, error) {
	return auth.GetUserIDFromContext(r.Context())
}
