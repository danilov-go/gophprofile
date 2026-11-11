package memory

import (
	"sync"
)

// MemStorage определяет хранилище данных в оперативной памяти.
type MemStorage struct {
	mu sync.RWMutex
}

// NewMemStorage создает новый экземпляр MemStorage.
func NewMemStorage() *MemStorage {
	return &MemStorage{}
}
