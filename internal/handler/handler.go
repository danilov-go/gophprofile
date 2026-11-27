// Package handler реализует HTTP-интерфейс приложения.
package handler

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/danilov-go/gophprofile/internal/models"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

// Service определяет методы взаимодействия с бизнес логикой.
type Service interface {
	SaveAvatar(ctx context.Context, imageData []byte, userID, fileName, memTipe string) (models.Avatar, error)
	GetAvatar(ctx context.Context, id, size, format string) (models.Avatar, []byte, error)
	GetAvatarByUserID(ctx context.Context, userID, size, format string) (models.Avatar, []byte, error)
	GetAvatarsByUserID(ctx context.Context, userID string) ([]models.MetaDataResponse, error)
	GetMetadata(ctx context.Context, id string) (models.MetaDataResponse, error)
	Delete(ctx context.Context, id, userID string) error
	DeleteByUserID(ctx context.Context, userID string) error
	GetUserID(tokenString, key string) (string, error)
	Health(ctx context.Context) models.HealthResponse
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
	http.ServeFile(w, r, filepath.Join(h.webPath, "index.html"))
}

// GetGallery отдает HTML-форму галереии пользователя.
func (h *Handler) GetGallery(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(h.webPath, "gallery.html"))
}

// HealthHandler проверяет состояние сервиса.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	resp := h.service.Health(r.Context())
	if resp.DB != "ok" || resp.S3 != "ok" || resp.Broker != "ok" {
		h.writeJSON(w, http.StatusServiceUnavailable, resp)
		return
	}
	h.writeJSON(w, http.StatusOK, resp)
}
