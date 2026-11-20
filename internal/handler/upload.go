package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/danilov-go/gophprofile/internal/models"
)

// UploadAvatar загружает новую аватарку.
func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request, userID string) {
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
	meta, err := h.service.SaveAvatar(r.Context(), imageData, userID, header.Filename, mimeType)
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
