package worker

import (
	"context"
	"time"

	"github.com/danilov-go/gophprofile/internal/models"
)

func (w *Worker) HandleUploadEvent(ctx context.Context, event models.AvatarUploadEvent) error {
	meta, err := w.repo.Get(ctx, event.AvatarID)
	if err != nil {
		return err
	}
	if meta.ProcessingStatus == "completed" || meta.DeletedAt != nil {
		return nil
	}
	imageData, err := w.s3.Download(ctx, event.S3Key)
	if err != nil {
		return err
	}
	thumbnailData, err := w.resizer.Resize(imageData)
	if err != nil {
		_ = w.repo.UpdateProcessingStatus(ctx, event.AvatarID, models.StatusFailed, nil)
		return err
	}
	thumbKeys := make(map[models.SizeAvatar]string, len(thumbnailData))
	for _, thumb := range thumbnailData {
		key := models.GenerateThumbnailS3Key(event.AvatarID, thumb.Size)
		if err := w.s3.Upload(ctx, key, thumb.Data); err != nil {
			_ = w.repo.UpdateProcessingStatus(ctx, event.AvatarID, models.StatusFailed, nil)
			return err
		}
		thumbKeys[thumb.Size] = key
	}
	return w.repo.UpdateProcessingStatus(ctx, event.AvatarID, models.StatusCompleted, thumbKeys)
}

func (w *Worker) HandleDeleteEvent(ctx context.Context, event models.AvatarDeleteEvent) error {
	for _, key := range event.S3Keys {
		if key == "" {
			continue
		}
		if err := w.s3.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) retry(ctx context.Context, fn func() error) error {
	var err error
	retryDelay := w.retryDelay
	for attempt := 1; attempt <= w.maxRetries; attempt++ {
		if err = fn(); err == nil {
			return nil
		}
		if attempt == w.maxRetries {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDelay):
			retryDelay *= 2
		}
	}
	return err
}
