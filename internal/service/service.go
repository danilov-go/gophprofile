// Package service реализует бизнес логику приложения.
package service

import (
	"context"

	"github.com/danilov-go/gophprofile/internal/models"
)

// Storage определяет методы взаимодействия с хранилищем метаданных аватарок.
type Storage interface {
	Save(ctx context.Context, meta models.Avatar) error
	Get(ctx context.Context, id string) (models.Avatar, error)
	GetByUserID(ctx context.Context, id string) (models.Avatar, error)
	GetAllByUserID(ctx context.Context, userID string) ([]models.Avatar, error)
	Delete(ctx context.Context, id string, keys []string) error
	Ping(ctx context.Context) error
}

// S3 определяет методы взаимодействия с объектным хранилищем файлов.
type S3 interface {
	Upload(ctx context.Context, s3Key string, imageData []byte) error
	Download(ctx context.Context, s3Key string) ([]byte, error)
	Delete(ctx context.Context, s3Key string) error
	Ping(ctx context.Context) error
}

// Producer определяет методы проверки работы брокера.
type Producer interface {
	Ping(ctx context.Context) error
}

// Converter определяет методы конвертации изображений и получения их размеров.
type Converter interface {
	Convert(imageData []byte, format string) ([]byte, string, error)
	GetDimensions(imageData []byte) (models.Dimensions, error)
}

// Service реализует бизнес логику работы с аватарками.
type Service struct {
	storage   Storage
	s3        S3
	producer  Producer
	converter Converter
}

// NewService создает новый экземпляр Service.
func NewService(storage Storage, s3 S3, producer Producer, converter Converter) *Service {
	return &Service{
		storage:   storage,
		s3:        s3,
		producer:  producer,
		converter: converter,
	}
}

// Health проверяет доступность базы данных, S3-хранилища и брокера.
func (s *Service) Health(ctx context.Context) models.HealthResponse {
	resp := models.HealthResponse{
		DB:     "ok",
		S3:     "ok",
		Broker: "ok",
	}
	if err := s.storage.Ping(ctx); err != nil {
		resp.DB = err.Error()
	}
	if err := s.s3.Ping(ctx); err != nil {
		resp.S3 = err.Error()
	}
	if err := s.producer.Ping(ctx); err != nil {
		resp.Broker = err.Error()
	}
	return resp
}
