// Package service реализует бизнес логику приложения.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"time"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

// Storage определяет методы взаимодействия с хранилищем метаданных аватарок.
type Storage interface {
	Save(ctx context.Context, meta models.Avatar) error
	Get(ctx context.Context, id string) (models.Avatar, error)
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

// Service реализует бизнес логику работы с аватарками.
type Service struct {
	storage  Storage
	s3       S3
	producer Producer
}

type Producer interface {
	Send(ctx context.Context, key, eventType string, eventBytes []byte) error
}

// NewService создает новый экземпляр Service.
func NewService(storage Storage, s3 S3, producer Producer) *Service {
	return &Service{
		storage:  storage,
		s3:       s3,
		producer: producer,
	}
}

// Save загружает аватарку в S3 и сохраняет метаданные в хранилище.
func (s *Service) Save(ctx context.Context, imageData []byte, userID, fileName, mimeType string) (models.Avatar, error) {
	if int64(len(imageData)) > models.MaxSize {
		return models.Avatar{}, models.ErrFileTooLarge
	}
	ext, ok := models.GetExt(mimeType)
	if !ok {
		return models.Avatar{}, models.ErrInvalidFormat
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return models.Avatar{}, models.ErrInvalidFormat
	}
	now := time.Now().UTC()
	meta := models.Avatar{
		ID:        uuid.NewString(),
		UserID:    userID,
		FileName:  fileName,
		MimeType:  mimeType,
		SizeBytes: int64(len(imageData)),
		Dimensions: models.Dimensions{
			Width:  cfg.Width,
			Height: cfg.Height,
		},
		ProcessingStatus: "processing",
		UploadStatus:     "uploaded",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	meta.GenerateS3Key(ext)
	if err := s.s3.Upload(ctx, meta.S3Key, imageData); err != nil {
		return models.Avatar{}, err
	}
	if err := s.storage.Save(ctx, meta); err != nil {
		_ = s.s3.Delete(context.WithoutCancel(ctx), meta.S3Key)
		return models.Avatar{}, err
	}
	if err := s.PublishUploadEvent(ctx, meta.ID, meta.UserID, meta.S3Key); err != nil {
		_ = s.storage.Delete(ctx, meta.ID)
		_ = s.s3.Delete(context.WithoutCancel(ctx), meta.S3Key)
		return models.Avatar{}, err
	}
	return meta, nil
}

func (s *Service) getAvatar(ctx context.Context, id string) (models.Avatar, error) {
	if _, err := uuid.Parse(id); err != nil {
		return models.Avatar{}, models.ErraAvatarNotFound
	}
	return s.storage.Get(ctx, id)
}

// Get возвращает метаданные и свм файл аватарки.
func (s *Service) Get(ctx context.Context, id string) (models.Avatar, []byte, error) {
	meta, err := s.getAvatar(ctx, id)
	if err != nil {
		return models.Avatar{}, nil, err
	}
	imageData, err := s.s3.Download(ctx, meta.S3Key)
	if err != nil {
		return models.Avatar{}, nil, err
	}
	return meta, imageData, nil
}

// GetMetadata возвращает метаданные аватарки.
func (s *Service) GetMetadata(ctx context.Context, id string) (models.MetaDataResponse, error) {
	meta, err := s.getAvatar(ctx, id)
	if err != nil {
		return models.MetaDataResponse{}, err
	}
	thumbnails := make([]models.Thumbnail, 0, len(meta.ThumbnailS3Keys))
	for size := range meta.ThumbnailS3Keys {
		thumbnails = append(thumbnails, models.Thumbnail{
			Size: string(size),
			URL:  fmt.Sprintf("/api/v1/avatars/%s?size=%s", meta.ID, size),
		})
	}
	return models.MetaDataResponse{
		ID:         meta.ID,
		UserID:     meta.UserID,
		FileName:   meta.FileName,
		MimeType:   meta.MimeType,
		Size:       meta.SizeBytes,
		Dimensions: meta.Dimensions,
		Thumbnails: thumbnails,
		CreatedAt:  meta.CreatedAt,
		UpdatedAt:  meta.UpdatedAt,
	}, nil
}

// Delete удаляет аватарку из базы данных и S3.
func (s *Service) Delete(ctx context.Context, id, userID string) error {
	meta, err := s.getAvatar(ctx, id)
	if err != nil {
		return err
	}
	if meta.UserID != userID {
		return models.ErrForbidden
	}
	if err := s.storage.Delete(ctx, id); err != nil {
		return err
	}
	return s.s3.Delete(ctx, meta.S3Key)
}

// Health проверяет доступность базы данных и S3-хранилища.
func (s *Service) Health(ctx context.Context) error {
	errS3 := s.s3.Ping(ctx)
	errDB := s.storage.Ping(ctx)
	return errors.Join(errS3, errDB)
}

func (s *Service) PublishUploadEvent(ctx context.Context, avatarID, userID, s3Key string) error {
	event := models.AvatarUploadEvent{
		AvatarID: avatarID,
		UserID:   userID,
		S3Key:    s3Key,
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.producer.Send(ctx, avatarID, "upload", eventBytes)
}

func (s *Service) PublishDeleteEvent(ctx context.Context, avatarID string, s3Keys []string) error {
	event := models.AvatarDeleteEvent{
		AvatarID: avatarID,
		S3Keys:   s3Keys,
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.producer.Send(ctx, avatarID, "delete", eventBytes)
}
