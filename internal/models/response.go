package models

import "time"

type UploadResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	MaxSize int64  `json:"max_size,omitempty"`
}

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
