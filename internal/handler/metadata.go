package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/danilov-go/gophprofile/internal/models"
)

// GetAvatarMetadata возвращает метаданные аватарки.
func (h *Handler) GetAvatarMetadata(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "avatar_id")
	if id == "" {
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Empty avatarID",
			Details: "avatar_id path parameter is required",
		})
		return
	}
	meta, err := h.service.GetMetadata(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, meta)
}

// GetUserAvatars возвращает список метаданных всех аватарок пользователя.
func (h *Handler) GetUserAvatars(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Empty userID",
			Details: "user_id path parameter is required",
		})
		return
	}
	avatars, err := h.service.GetAvatarsByUserID(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, avatars)
}
