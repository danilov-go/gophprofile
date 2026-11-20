package service

import (
	"context"
	"time"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/google/uuid"
)

// SaveAvatar загружает аватарку в S3 и сохраняет метаданные в хранилище.
func (s *Service) SaveAvatar(ctx context.Context, imageData []byte, userID, fileName, mimeType string) (models.Avatar, error) {
	if int64(len(imageData)) > models.MaxSize {
		return models.Avatar{}, models.ErrFileTooLarge
	}
	ext, ok := models.GetExt(mimeType)
	if !ok {
		return models.Avatar{}, models.ErrInvalidFormat
	}
	dimensions, err := s.converter.GetDimensions(imageData)
	if err != nil {
		return models.Avatar{}, err
	}
	now := time.Now().UTC()
	meta := models.Avatar{
		ID:               uuid.NewString(),
		UserID:           userID,
		FileName:         fileName,
		MimeType:         mimeType,
		SizeBytes:        int64(len(imageData)),
		Dimensions:       dimensions,
		ProcessingStatus: models.StatusProcessing,
		UploadStatus:     models.UploadStatusUploaded,
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
