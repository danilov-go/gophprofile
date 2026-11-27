package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/danilov-go/gophprofile/internal/models"
)

// GetAvatar возвращает метаданные и файл аватарки.
func (s *Service) GetAvatar(ctx context.Context, id, size, format string) (models.Avatar, []byte, error) {
	meta, err := s.getAvatar(ctx, id)
	if err != nil {
		return models.Avatar{}, nil, err
	}
	return s.loadAvatar(ctx, meta, size, format)
}

// GetAvatarByUserID возвращает метаданные и файл аватарки по ID пользователя.
func (s *Service) GetAvatarByUserID(ctx context.Context, userID, size, format string) (models.Avatar, []byte, error) {
	meta, err := s.storage.GetByUserID(ctx, userID)
	if err != nil {
		return models.Avatar{}, nil, err
	}
	return s.loadAvatar(ctx, meta, size, format)
}

func (s *Service) loadAvatar(ctx context.Context, meta models.Avatar, size, format string) (models.Avatar, []byte, error) {
	var key string
	if size == "" || size == string(models.SizeOriginal) {
		key = meta.S3Key
	} else {
		avatarSize := models.SizeAvatar(size)
		if _, ok := models.Targets[avatarSize]; !ok {
			return models.Avatar{}, nil, models.ErrInvalidSize
		}
		thumbKey, ok := meta.ThumbnailS3Keys[avatarSize]
		if !ok || thumbKey == "" {
			return models.Avatar{}, nil, models.ErrAvatarNotFound
		}
		key = thumbKey
		meta.MimeType = "image/jpeg"
	}
	imageData, err := s.s3.Download(ctx, key)
	if err != nil {
		return models.Avatar{}, nil, err
	}
	if format != "" {
		switch format {
		case "jpeg", "png", "webp":
			if meta.MimeType != "image/"+format {
				imageData, meta.MimeType, err = s.converter.Convert(imageData, format)
				if err != nil {
					return models.Avatar{}, nil, err
				}
			}
		default:
			return models.Avatar{}, nil, models.ErrInvalidFormat
		}
	}
	return meta, imageData, nil
}

func (s *Service) getAvatar(ctx context.Context, id string) (models.Avatar, error) {
	if _, err := uuid.Parse(id); err != nil {
		return models.Avatar{}, models.ErrAvatarNotFound
	}
	return s.storage.Get(ctx, id)
}
