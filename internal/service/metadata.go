package service

import (
	"context"
	"fmt"

	"github.com/danilov-go/gophprofile/internal/models"
)

// GetMetadata возвращает метаданные аватарки.
func (s *Service) GetMetadata(ctx context.Context, id string) (models.MetaDataResponse, error) {
	meta, err := s.getAvatar(ctx, id)
	if err != nil {
		return models.MetaDataResponse{}, err
	}
	return getMetaDataResponse(meta), nil
}

// GetAvatarsByUserID возвращает список метаданных всех аватарок пользователя.
func (s *Service) GetAvatarsByUserID(ctx context.Context, userID string) ([]models.MetaDataResponse, error) {
	avatars, err := s.storage.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]models.MetaDataResponse, 0, len(avatars))
	for _, meta := range avatars {
		res = append(res, getMetaDataResponse(meta))
	}
	return res, nil
}

func getMetaDataResponse(meta models.Avatar) models.MetaDataResponse {
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
	}
}
