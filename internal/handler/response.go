package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/danilov-go/gophprofile/internal/models"
)

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	switch {
	case errors.Is(err, models.ErrInvalidFormat):
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid file format",
			Details: "Supported formats: jpeg, png, webp",
		})
	case errors.Is(err, models.ErrInvalidSize):
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid size",
			Details: "Supported size: original, 100x100, 300x300",
		})
	case errors.Is(err, models.ErrFileTooLarge), errors.As(err, &maxBytesErr):
		h.writeJSON(w, http.StatusRequestEntityTooLarge, models.ErrorResponse{
			Error:   "File too large",
			MaxSize: models.MaxSize,
		})
	case errors.Is(err, models.ErrAvatarNotFound):
		h.writeJSON(w, http.StatusNotFound, models.ErrorResponse{
			Error: "Avatar not found",
		})
	case errors.Is(err, models.ErrForbidden):
		h.writeJSON(w, http.StatusForbidden, models.ErrorResponse{
			Error:   "Forbidden",
			Details: "You can only delete your own avatars",
		})
	default:
		h.logger.Errorw("внутренняя ошибка сервера", "error", err)
		h.writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "Internal server error",
		})
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Errorw("ошибка кодирования JSON-ответа", "error", err)
	}
}
