package service

import (
	"context"
	"encoding/json"

	"github.com/danilov-go/gophprofile/internal/models"
)

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
	return s.producer.Send(ctx, avatarID, models.EventTypeUpload, eventBytes)
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
	return s.producer.Send(ctx, avatarID, models.EventTypeDelete, eventBytes)
}
