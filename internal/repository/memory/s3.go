package memory

import (
	"bytes"
	"context"
	"sync"

	"github.com/danilov-go/gophprofile/internal/models"
)

// MemStorageS3 определяет хранение картинок в оперативной памяти.
type MemStorageS3 struct {
	mu    sync.RWMutex
	image map[string][]byte
}

// NewMemStorageS3 создает новый экземпляр MemStorage.
func NewMemStorageS3() *MemStorageS3 {
	return &MemStorageS3{
		image: make(map[string][]byte),
	}
}

// Upload загружает данные в оперативную память.
func (m *MemStorageS3) Upload(ctx context.Context, s3Key string, imageData []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.image[s3Key] = bytes.Clone(imageData)
	return nil
}

// Download скачивает файл из оперативной памяти.
func (m *MemStorageS3) Download(ctx context.Context, s3Key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	imageData, ok := m.image[s3Key]
	if !ok {
		return nil, models.ErraAvatarNotFound
	}
	return bytes.Clone(imageData), nil
}

// Delete удаляет файл из оперативной памяти.
func (m *MemStorageS3) Delete(ctx context.Context, s3Key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.image, s3Key)
	return nil
}

// Ping проверяет доступность хранилища.
func (m *MemStorageS3) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
