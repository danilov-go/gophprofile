// Package models содержит структуры данных и константы для работы с аватарками.
package models

import (
	"errors"
	"fmt"
	"time"
)

// ProcessingStatus представляет статус фоновой обработки аватарки.
type ProcessingStatus string

// Поддерживаемые статусы обработки аватарок.
const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
)

// UploadStatus представляет статус загрузки файла в хранилище.
type UploadStatus string

// Поддерживаемые статусы загрузки аватарок.
const (
	UploadStatusUploading UploadStatus = "uploading"
	UploadStatusUploaded  UploadStatus = "uploaded"
)

// Ошибки бизнес логики при работе с аватарками.
var (
	ErrInvalidFormat  = errors.New("invalid file format")
	ErrInvalidSize    = errors.New("invalid size")
	ErrFileTooLarge   = errors.New("file too large")
	ErrAvatarNotFound = errors.New("avatar not found")
	ErrForbidden      = errors.New("forbidden")
)

// Avatar определяет метаданные аватарки.
type Avatar struct {
	ID               string                `json:"id"`
	UserID           string                `json:"user_id"`
	FileName         string                `json:"file_name"`
	MimeType         string                `json:"mime_type"`
	SizeBytes        int64                 `json:"size_bytes"`
	Dimensions       Dimensions            `json:"dimensions"`
	S3Key            string                `json:"-"`
	ThumbnailS3Keys  map[SizeAvatar]string `json:"-"`
	UploadStatus     UploadStatus          `json:"upload_status"`
	ProcessingStatus ProcessingStatus      `json:"processing_status"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	DeletedAt        *time.Time            `json:"deleted_at,omitempty"`
}

// GenerateS3Key формирует ключ аватарки для хранения в S3.
func (a *Avatar) GenerateS3Key(ext string) {
	a.S3Key = fmt.Sprintf("avatars/%s/%s%s", a.UserID, a.ID, ext)
}

// GenerateThumbnailS3Key формирует ключ миниатюры для хранения в S3.
func GenerateThumbnailS3Key(avatarID string, size SizeAvatar) string {
	return fmt.Sprintf("thumbnails/%s/%s.jpg", avatarID, size)
}
