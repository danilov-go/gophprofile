package handler

import (
	"net/http"

	"github.com/danilov-go/gophprofile/internal/models"
)

// GetAvatarMetadata возвращает метаданные аватарки.
func (h *Handler) GetAvatarMetadata(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("avatar_id")
	meta, err := h.service.GetMetadata(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, meta)
}

// GetUserAvatars возвращает список метаданных всех аватарок пользователя.
func (h *Handler) GetUserAvatars(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
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
