package handler

import (
	"bytes"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/danilov-go/gophprofile/internal/models"
)

// GetUserAvatar возвращает текущую аватарку пользователя.
func (h *Handler) GetUserAvatar(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Empty userID",
			Details: "user_id path parameter is required",
		})
		return
	}
	size := r.URL.Query().Get("size")
	format := r.URL.Query().Get("format")
	meta, imageData, err := h.service.GetAvatarByUserID(r.Context(), userID, size, format)
	if err != nil {
		h.handleError(w, err)
		return
	}
	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Cache-Control", "max-age=86400")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(imageData))
}

// GetAvatar возвращает аватарку.
func (h *Handler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "avatar_id")
	if id == "" {
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Empty avatarID",
			Details: "avatar_id path parameter is required",
		})
		return
	}
	size := r.URL.Query().Get("size")
	format := r.URL.Query().Get("format")
	meta, imageData, err := h.service.GetAvatar(r.Context(), id, size, format)
	if err != nil {
		h.handleError(w, err)
		return
	}
	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Cache-Control", "max-age=86400")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(imageData))
}
