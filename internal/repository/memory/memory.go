// Package memory реализует хранилище данных в оперативной памяти.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/danilov-go/gophprofile/internal/models"
)

// MemStorage определяет хранилище метаданных в оперативной памяти.
type MemStorage struct {
	mu       sync.RWMutex
	metaData map[string]models.Avatar
}

// NewMemStorage создает новый экземпляр MemStorage.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		metaData: make(map[string]models.Avatar),
	}
}

// Save сохраняет метаданные аватарки в оперативной памяти.
func (m *MemStorage) Save(ctx context.Context, meta models.Avatar) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metaData[meta.ID] = meta
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
		return models.Avatar{}, models.ErraAvatarNotFound
	}
	return meta, nil
}

// Delete удаляет метаданные аватарки из оперативной памяти.
func (m *MemStorage) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, ok := m.metaData[id]
	if !ok || meta.DeletedAt != nil {
		return models.ErraAvatarNotFound
	}
	now := time.Now().UTC()
	meta.DeletedAt = &now
	meta.UpdatedAt = now
	m.metaData[id] = meta
	return nil
}

// Ping проверяет доступность хранилища.
func (m *MemStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
