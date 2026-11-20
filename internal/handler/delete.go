package handler

import (
	"net/http"

	"github.com/danilov-go/gophprofile/internal/models"
)

// DeleteAvatar удаляет аватарку.
func (h *Handler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
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

// DeleteUserAvatar удаляет текущую аватарку пользователя.
func (h *Handler) DeleteUserAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Empty userID",
			Details: "user_id path parameter is required",
		})
		return
	}
	headerUserID := r.Header.Get("X-User-ID")
	if headerUserID == "" {
		h.writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Details: "X-User-ID header is required",
		})
		return
	}
	if headerUserID != userID {
		h.handleError(w, models.ErrForbidden)
		return
	}
	if err := h.service.DeleteByUserID(r.Context(), userID); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
