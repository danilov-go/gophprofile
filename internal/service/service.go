package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"time"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

type Storage interface {
	Save(ctx context.Context, meta models.Avatar) error
	Get(ctx context.Context, id string) (models.Avatar, error)
	Delete(ctx context.Context, id string) error
}

type S3 interface {
	Upload(ctx context.Context, s3Key string, imageData []byte) error
	Download(ctx context.Context, s3Key string) ([]byte, error)
	Delete(ctx context.Context, s3Key string) error
}

type Service struct {
	storage Storage
	s3      S3
}

func NewService(s Storage, s3 S3) *Service {
	return &Service{
		storage: s,
		s3:      s3,
	}
}

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
		CreatedAt: now,
		UpdatedAt: now,
	}
	meta.GenerateS3Key(ext)
	if err := s.s3.Upload(ctx, meta.S3Key, imageData); err != nil {
		return models.Avatar{}, err
	}
	if err := s.storage.Save(ctx, meta); err != nil {
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
