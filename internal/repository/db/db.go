// Package db реализует хранилище данных в базе PostgreSQL.
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type storageDB struct {
	db *sql.DB
}

// NewStorageDB создает новый экземпляр storageDB.
func NewStorageDB(sql *sql.DB) *storageDB {
	return &storageDB{
		db: sql,
	}
}

// InitDB инициализирует подключение к базе данных PostgreSQL по переданной строке соединения.
func InitDB(ps string) (*storageDB, error) {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	return &storageDB{
		db: db,
	}, nil
}

// Save сохраняет метаданные аватарки в базе данных.
func (s *storageDB) Save(ctx context.Context, meta models.Avatar) error {
	thumbs, err := json.Marshal(meta.ThumbnailS3Keys)
	if err != nil {
		return err
	}
	if meta.ThumbnailS3Keys == nil {
		thumbs = []byte("{}")
	}
	query := `
		INSERT INTO avatars (id, user_id, file_name, mime_type, size_bytes,
		width, height, s3_key, thumbnail_s3_keys,upload_status, processing_status,
		created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = s.db.ExecContext(ctx, query,
		meta.ID, meta.UserID, meta.FileName, meta.MimeType, meta.SizeBytes,
		meta.Dimensions.Width, meta.Dimensions.Height, meta.S3Key, thumbs,
		meta.UploadStatus, meta.ProcessingStatus, meta.CreatedAt, meta.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// Get возвращает метаданные аватарки из базы данных.
func (s *storageDB) Get(ctx context.Context, id string) (models.Avatar, error) {
	query := `
		SELECT id, user_id, file_name, mime_type, size_bytes, width, height, 
		s3_key, thumbnail_s3_keys, upload_status, processing_status, created_at, 
		updated_at, deleted_at FROM avatars WHERE id = $1 AND deleted_at IS NULL
	`
	var meta models.Avatar
	var thumbs []byte
	var deletedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&meta.ID, &meta.UserID, &meta.FileName, &meta.MimeType, &meta.SizeBytes,
		&meta.Dimensions.Width, &meta.Dimensions.Height, &meta.S3Key, &thumbs,
		&meta.UploadStatus, &meta.ProcessingStatus, &meta.CreatedAt, &meta.UpdatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Avatar{}, models.ErraAvatarNotFound
		}
		return models.Avatar{}, err
	}
	if len(thumbs) > 0 {
		if err := json.Unmarshal(thumbs, &meta.ThumbnailS3Keys); err != nil {
			return models.Avatar{}, err
		}
	}
	if deletedAt.Valid {
		meta.DeletedAt = &deletedAt.Time
	}
	return meta, nil
}

// Delete помечает метаданные аватарки как удаленные в базе данных.
func (s *storageDB) Delete(ctx context.Context, id string) error {
	query := `UPDATE avatars SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return models.ErraAvatarNotFound
	}
	return nil
}

// UpdateProcessingStatus обновляет статус обработки аватарки и ключи миниатюр в базе данных.
func (s *storageDB) UpdateProcessingStatus(ctx context.Context, id, status string, thumbnailKeys map[models.SizeAvatar]string) error {
	var thumbs []byte
	var err error
	if thumbnailKeys != nil {
		thumbs, err = json.Marshal(thumbnailKeys)
		if err != nil {
			return err
		}
	}
	query := `
		UPDATE avatars 
		SET processing_status = $1, 
		    thumbnail_s3_keys = COALESCE($2::jsonb, thumbnail_s3_keys), 
		    updated_at = $3 
		WHERE id = $4 AND deleted_at IS NULL
	`
	res, err := s.db.ExecContext(ctx, query, status, thumbs, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return models.ErraAvatarNotFound
	}
	return nil
}

// CLose закрывает соединение с БД.
func (s *storageDB) Close() error {
	return s.db.Close()
}

// Ping проверяет доступность хранилища.
func (s *storageDB) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// GetByUserID возвращает актуальную аватарку пользователя из базы данных.
func (s *storageDB) GetByUserID(ctx context.Context, userID string) (models.Avatar, error) {
	query := `
		SELECT id, user_id, file_name, mime_type, size_bytes, width, height, 
		s3_key, thumbnail_s3_keys, upload_status, processing_status, created_at, 
		updated_at, deleted_at FROM avatars WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT 1
	`
	var meta models.Avatar
	var thumbs []byte
	var deletedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&meta.ID, &meta.UserID, &meta.FileName, &meta.MimeType, &meta.SizeBytes,
		&meta.Dimensions.Width, &meta.Dimensions.Height, &meta.S3Key, &thumbs,
		&meta.UploadStatus, &meta.ProcessingStatus, &meta.CreatedAt, &meta.UpdatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Avatar{}, models.ErraAvatarNotFound
		}
		return models.Avatar{}, err
	}
	if len(thumbs) > 0 {
		if err := json.Unmarshal(thumbs, &meta.ThumbnailS3Keys); err != nil {
			return models.Avatar{}, err
		}
	}
	if deletedAt.Valid {
		meta.DeletedAt = &deletedAt.Time
	}
	return meta, nil
}

// GetAllByUserID возвращает список всех аватарок пользователя из базы данных.
func (s *storageDB) GetAllByUserID(ctx context.Context, userID string) ([]models.Avatar, error) {
	query := `
		SELECT id, user_id, file_name, mime_type, size_bytes, width, height, 
		s3_key, thumbnail_s3_keys, upload_status, processing_status, created_at, 
		updated_at, deleted_at FROM avatars WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var avatars []models.Avatar
	for rows.Next() {
		var meta models.Avatar
		var thumbs []byte
		var deletedAt sql.NullTime
		err := rows.Scan(
			&meta.ID, &meta.UserID, &meta.FileName, &meta.MimeType, &meta.SizeBytes,
			&meta.Dimensions.Width, &meta.Dimensions.Height, &meta.S3Key, &thumbs,
			&meta.UploadStatus, &meta.ProcessingStatus, &meta.CreatedAt, &meta.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, err
		}
		if len(thumbs) > 0 {
			if err := json.Unmarshal(thumbs, &meta.ThumbnailS3Keys); err != nil {
				return nil, err
			}
		}
		if deletedAt.Valid {
			meta.DeletedAt = &deletedAt.Time
		}
		avatars = append(avatars, meta)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return avatars, nil
}
