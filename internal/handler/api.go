package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/danilov-go/gophprofile/internal/models"
)

// APIUploadHandler загружает новую аватарку.
func (h *Handler) APIUploadHandler(w http.ResponseWriter, r *http.Request) {
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
		Status:    meta.ProcessingStatus,
		CreatedAt: meta.CreatedAt,
	})
}

// APIGetAvatarHandler возвращает аватарку.
func (h *Handler) APIGetAvatarHandler(w http.ResponseWriter, r *http.Request) {
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

// APIGetMetaDataHandler возвращает метаданные аватарки.
func (h *Handler) APIGetMetaDataHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("avatar_id")
	meta, err := h.service.GetMetadata(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, meta)
}

// APIDeleteAvatarHandler удаляет аватарку.
func (h *Handler) APIDeleteAvatarHandler(w http.ResponseWriter, r *http.Request) {
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

// HealthHandler проверяет состояние сервиса.
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Health(r.Context()); err != nil {
		h.writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
