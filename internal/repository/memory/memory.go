// Package memory реализует хранилище данных в оперативной памяти.
package memory

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/danilov-go/gophprofile/internal/models"
)

// MemStorage определяет хранилище метаданных в оперативной памяти.
type MemStorage struct {
	mu       sync.RWMutex
	metaData map[string]models.Avatar
	events   map[string]models.Event
}

// NewMemStorage создает новый экземпляр MemStorage.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		metaData: make(map[string]models.Avatar),
		events:   make(map[string]models.Event),
	}
}

// Save сохраняет метаданные аватарки в оперативной памяти.
func (m *MemStorage) Save(ctx context.Context, meta models.Avatar) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	event, err := json.Marshal(models.AvatarUploadEvent{
		AvatarID: meta.ID,
		UserID:   meta.UserID,
		S3Key:    meta.S3Key,
	})
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metaData[meta.ID] = meta
	eventID := uuid.NewString()
	m.events[eventID] = models.Event{
		ID:        eventID,
		EventType: models.EventTypeUpload,
		AvatarID:  meta.ID,
		Payload:   event,
		CreatedAt: time.Now().UTC(),
	}
	return nil
}

// Get возвращает метаданные аватарки из оперативной памяти.
func (m *MemStorage) Get(ctx context.Context, id string) (models.Avatar, error) {
	if err := ctx.Err(); err != nil {
		return models.Avatar{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	meta, ok := m.metaData[id]
	if !ok || meta.DeletedAt != nil {
		return models.Avatar{}, models.ErrAvatarNotFound
	}
	return meta, nil
}

// Delete удаляет метаданные аватарки из оперативной памяти.
func (m *MemStorage) Delete(ctx context.Context, id string, keys []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	event, err := json.Marshal(models.AvatarDeleteEvent{
		AvatarID: id,
		S3Keys:   keys,
	})
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, ok := m.metaData[id]
	if !ok || meta.DeletedAt != nil {
		return models.ErrAvatarNotFound
	}
	now := time.Now().UTC()
	meta.DeletedAt = &now
	meta.UpdatedAt = now
	m.metaData[id] = meta
	eventID := uuid.NewString()
	m.events[eventID] = models.Event{
		ID:        eventID,
		EventType: models.EventTypeDelete,
		AvatarID:  id,
		Payload:   event,
		CreatedAt: now,
	}
	return nil
}

// GetAllByUserID возвращает список всех аватарок пользователя из оперативной памяти.
func (m *MemStorage) GetAllByUserID(ctx context.Context, userID string) ([]models.Avatar, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var avatars []models.Avatar
	for _, meta := range m.metaData {
		if meta.UserID == userID && meta.DeletedAt == nil {
			avatars = append(avatars, meta)
		}
	}
	sort.Slice(avatars, func(i, j int) bool {
		return avatars[i].CreatedAt.After(avatars[j].CreatedAt)
	})
	return avatars, nil
}

// GetByUserID возвращает актуальную аватарку пользователя из оперативной памяти.
func (m *MemStorage) GetByUserID(ctx context.Context, userID string) (models.Avatar, error) {
	if err := ctx.Err(); err != nil {
		return models.Avatar{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var latest models.Avatar
	var found bool
	for _, meta := range m.metaData {
		if meta.UserID == userID && meta.DeletedAt == nil && meta.CreatedAt.After(latest.CreatedAt) {
			latest = meta
			found = true
		}
	}
	if !found {
		return models.Avatar{}, models.ErrAvatarNotFound
	}
	return latest, nil
}

// Ping проверяет доступность хранилища.
func (m *MemStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

// GetEvents возвращает неотправленные события из оперативной памяти.
func (m *MemStorage) GetEvents(ctx context.Context, limit int) ([]models.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var events []models.Event
	for _, event := range m.events {
		if event.ProcessedAt == nil {
			events = append(events, event)
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}
	return events, nil
}

// UpdateEventStatus помечает событие как отправленное.
func (m *MemStorage) UpdateEventStatus(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	event, ok := m.events[id]
	if !ok {
		return nil
	}
	now := time.Now().UTC()
	event.ProcessedAt = &now
	m.events[id] = event
	return nil
}

// UpdateProcessingStatus обновляет статус обработки аватарки и ключи миниатюр.
func (m *MemStorage) UpdateProcessingStatus(ctx context.Context, id string, status models.ProcessingStatus, thumbnailKeys map[models.SizeAvatar]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, ok := m.metaData[id]
	if !ok || meta.DeletedAt != nil {
		return models.ErrAvatarNotFound
	}
	meta.ProcessingStatus = status
	if thumbnailKeys != nil {
		meta.ThumbnailS3Keys = thumbnailKeys
	}
	meta.UpdatedAt = time.Now().UTC()
	m.metaData[id] = meta
	return nil
}
