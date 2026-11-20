package models

import "time"

// MaxSize определяет максимальный размер аватарки.
const (
	MaxSize = 10 << 20
)

// HealthResponse определяет ответ сервера о состоянии компонентов.
type HealthResponse struct {
	DB     string `json:"db"`
	S3     string `json:"s3"`
	Broker string `json:"broker"`
}

// UploadResponse определяет ответ сервера при успешной загрузке аватарки.
type UploadResponse struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	URL       string           `json:"url"`
	Status    ProcessingStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
}

// ErrorResponse определяет ответ сервера при возникновении ошибки.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	MaxSize int64  `json:"max_size,omitempty"`
}

// MetaDataResponse определяет ответ сервера с метаданными аватарки.
type MetaDataResponse struct {
	ID         string      `json:"id"`
	UserID     string      `json:"user_id"`
	FileName   string      `json:"file_name"`
	MimeType   string      `json:"mime_type"`
	Size       int64       `json:"size"`
	Dimensions Dimensions  `json:"dimensions"`
	Thumbnails []Thumbnail `json:"thumbnails"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
