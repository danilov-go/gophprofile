package service

import (
	"context"

	"github.com/danilov-go/gophprofile/internal/models"
)

// DeleteByUserID удаляет текущую аватарку пользователя по его ID из базы данных и S3.
func (s *Service) DeleteByUserID(ctx context.Context, userID string) error {
	meta, err := s.storage.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.deleteAvatar(ctx, meta, userID)
}

// Delete удаляет аватарку из базы данных и S3.
func (s *Service) Delete(ctx context.Context, id, userID string) error {
	meta, err := s.getAvatar(ctx, id)
	if err != nil {
		return err
	}
	return s.deleteAvatar(ctx, meta, userID)
}

func (s *Service) deleteAvatar(ctx context.Context, meta models.Avatar, actorUserID string) error {
	if meta.UserID != actorUserID {
		return models.ErrForbidden
	}
	s3Keys := make([]string, 0, len(meta.ThumbnailS3Keys)+1)
	s3Keys = append(s3Keys, meta.S3Key)
	for _, thumbKey := range meta.ThumbnailS3Keys {
		if thumbKey != "" {
			s3Keys = append(s3Keys, thumbKey)
		}
	}
	if err := s.storage.Delete(ctx, meta.ID); err != nil {
		return err
	}
	return s.PublishDeleteEvent(ctx, meta.ID, s3Keys)
}
