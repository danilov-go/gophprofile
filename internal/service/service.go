// Package service реализует бизнес логику приложения.
package service

import (
	"context"
	"errors"

	"github.com/danilov-go/gophprofile/internal/models"
)

// Storage определяет методы взаимодействия с хранилищем метаданных аватарок.
type Storage interface {
	Save(ctx context.Context, meta models.Avatar) error
	Get(ctx context.Context, id string) (models.Avatar, error)
	GetByUserID(ctx context.Context, id string) (models.Avatar, error)
	GetAllByUserID(ctx context.Context, userID string) ([]models.Avatar, error)
	Delete(ctx context.Context, id string) error
	Ping(ctx context.Context) error
}

// S3 определяет методы взаимодействия с объектным хранилищем файлов.
type S3 interface {
	Upload(ctx context.Context, s3Key string, imageData []byte) error
	Download(ctx context.Context, s3Key string) ([]byte, error)
	Delete(ctx context.Context, s3Key string) error
	Ping(ctx context.Context) error
}

type Producer interface {
	Send(ctx context.Context, key, eventType string, eventBytes []byte) error
	Ping(ctx context.Context) error
}

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

// Health проверяет доступность базы данных и S3-хранилища.
func (s *Service) Health(ctx context.Context) error {
	errS3 := s.s3.Ping(ctx)
	errDB := s.storage.Ping(ctx)
	errProducer := s.producer.Ping(ctx)
	return errors.Join(errS3, errDB, errProducer)
}
