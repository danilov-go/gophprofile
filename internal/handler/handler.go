// Package handler реализует HTTP-интерфейс приложения.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/danilov-go/gophprofile/internal/models"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

// Service определяет методы взаимодействия с бизнес логикой.
type Service interface {
	Save(ctx context.Context, imageData []byte, userID, fileName, memTipe string) (models.Avatar, error)
	Get(ctx context.Context, id string) (models.Avatar, []byte, error)
	GetMetadata(ctx context.Context, id string) (models.MetaDataResponse, error)
	Delete(ctx context.Context, id, userID string) error
	Health(ctx context.Context) error
}

// Handler связывает HTTP-запросов с бизнес логикой.
type Handler struct {
	webPath string
	service Service
	logger  log
}

// NewHandler возвращает новый экземпляр Handler.
func NewHandler(service Service, l log, webPath string) *Handler {
	return &Handler{
		webPath: webPath,
		service: service,
		logger:  l,
	}
}

// GetForm отдает HTML-форму загрузки аватарки.
func (h *Handler) GetForm(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.webPath)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	switch {
	case errors.Is(err, models.ErrInvalidFormat):
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid file format",
			Details: "Supported formats: jpeg, png, webp",
		})
	case errors.Is(err, models.ErrFileTooLarge), errors.As(err, &maxBytesErr):
		h.writeJSON(w, http.StatusRequestEntityTooLarge, models.ErrorResponse{
			Error:   "File too large",
			MaxSize: models.MaxSize,
		})
	case errors.Is(err, models.ErraAvatarNotFound):
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Errorw("ошибка кодирования JSON-ответа", "error", err)
	}
}
