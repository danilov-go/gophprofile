package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/danilov-go/gophprofile/internal/models"
)

// DeleteAvatar удаляет аватарку.
func (h *Handler) DeleteAvatar(w http.ResponseWriter, r *http.Request, userID string) {
	id := chi.URLParam(r, "avatar_id")
	if id == "" {
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Empty avatarID",
			Details: "avatar_id path parameter is required",
		})
		return
	}
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteUserAvatar удаляет текущую аватарку пользователя.
func (h *Handler) DeleteUserAvatar(w http.ResponseWriter, r *http.Request, userID string) {
	if err := h.service.DeleteByUserID(r.Context(), userID); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
