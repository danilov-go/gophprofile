// Package models содержит структуры данных и константы для работы с аватарками.
package models

import (
	"errors"
	"fmt"
	"time"
)

// MaxSize определяет максимальный размер аватарки.
const (
	MaxSize = 10 << 20
)

// SizeAvatar представляет доступные размеры аватарок.
type SizeAvatar string

// Поддерживаемых размеры аватарок.
const (
	SizeSmall    SizeAvatar = "100x100"
	SizeMedium   SizeAvatar = "300x300"
	SizeOriginal SizeAvatar = "original"
)

// Dimensions определяет ширину и высоту аватари в пикселях.
type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Thumbnail определяет размер и URL-адрес миниатюры аватарки.
type Thumbnail struct {
	Size string `json:"size"`
	URL  string `json:"url"`
}

type ThumbnailData struct {
	Size SizeAvatar
	Data []byte
}

var Targets = map[SizeAvatar]Dimensions{
	SizeSmall:  {Width: 100, Height: 100},
	SizeMedium: {Width: 300, Height: 300},
}

// Ошибки бизнес логики при работе с аватарками.
var (
	ErrInvalidFormat   = errors.New("invalid file format")
	ErrInvalidSize     = errors.New("invalid size")
	ErrFileTooLarge    = errors.New("file too large")
	ErraAvatarNotFound = errors.New("avatar not found")
	ErrForbidden       = errors.New("forbidden")
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
	UploadStatus     string                `json:"upload_status"`
	ProcessingStatus string                `json:"processing_status"`
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

var mimeToExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// GetExt возвращает расширение аватарки по ее MIME-типу.
func GetExt(mimeType string) (string, bool) {
	ext, ok := mimeToExt[mimeType]
	return ext, ok
}
