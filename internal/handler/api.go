package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/danilov-go/gophprofile/internal/models"
)

func (h *Handler) ApiUploadHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Empty userID",
			Details: "X-User-ID header is required",
		})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, models.MaxSize)
	file, header, err := r.FormFile("image")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			h.handleError(w, models.ErrFileTooLarge)
			return
		}
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()
	imageData, err := io.ReadAll(file)
	if err != nil {
		h.handleError(w, err)
		return
	}
	if len(imageData) == 0 {
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Empty file",
			Details: "Uploaded file cannot be empty",
		})
		return
	}
	mimeType := http.DetectContentType(imageData)
	meta, err := h.service.Save(r.Context(), imageData, userID, header.Filename, mimeType)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, models.UploadResponse{
		ID:        meta.ID,
		UserID:    meta.UserID,
		URL:       "/api/v1/avatars/" + meta.ID,
		Status:    "processing",
		CreatedAt: meta.CreatedAt,
	})
}

func (h *Handler) ApiGetAvatarHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("avatar_id")
	meta, imageData, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Cache-Control", "max-age=86400")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(imageData); err != nil {
		h.logger.Errorw("ошибка записи тела ответа", "error", err)
	}
}

func (h *Handler) ApiGetMetaDataHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("avatar_id")
	meta, err := h.service.GetMetadata(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, meta)
}

func (h *Handler) ApiDeleteAvatarHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Empty userID",
			Details: "X-User-ID header is required",
		})
		return
	}
	id := r.PathValue("avatar_id")
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
