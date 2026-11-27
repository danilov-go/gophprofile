package models

import "time"

type EventType string

const (
	EventTypeHeader           = "event_type"
	EventTypeUpload EventType = "upload"
	EventTypeDelete EventType = "delete"
)

type AvatarUploadEvent struct {
	AvatarID string `json:"avatar_id"`
	UserID   string `json:"user_id"`
	S3Key    string `json:"s3_key"`
}

type AvatarDeleteEvent struct {
	AvatarID string   `json:"avatar_id"`
	S3Keys   []string `json:"s3_keys"`
}

type Event struct {
	ID          string
	EventType   EventType
	AvatarID    string
	Payload     []byte
	CreatedAt   time.Time
	ProcessedAt *time.Time
}
