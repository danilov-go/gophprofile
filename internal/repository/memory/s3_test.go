package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/danilov-go/gophprofile/internal/repository/memory"
)

func TestMemStorageS3_Upload(t *testing.T) {
	expKey := "avatars/userID/image.png"
	expData := []byte("image.png")
	tests := []struct {
		name      string
		cancelCtx bool
		expErr    error
	}{
		{
			name:      "положительный тест",
			cancelCtx: false,
			expErr:    nil,
		},
		{
			name:      "ошибка отмены контекста",
			cancelCtx: true,
			expErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorageS3()
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}
			err := storage.Upload(ctx, expKey, expData)
			if tt.expErr != nil {
				assert.ErrorIs(t, err, tt.expErr)
				return
			}
			require.NoError(t, err)
			imageData, err := storage.Download(context.Background(), expKey)
			require.NoError(t, err)
			assert.Equal(t, expData, imageData)
		})
	}
}

func TestMemStorageS3_Download(t *testing.T) {
	expKey := "avatars/userID/image.png"
	expData := []byte("image.png")
	tests := []struct {
		name      string
		cancelCtx bool
		upload    bool
		expKey    string
		expData   []byte
		expErr    error
	}{
		{
			name:      "положительный тест",
			cancelCtx: false,
			upload:    true,
			expKey:    expKey,
			expData:   expData,
			expErr:    nil,
		},
		{
			name:      "файл не найден",
			cancelCtx: false,
			upload:    false,
			expKey:    "avatars/userID/unknown.png",
			expData:   nil,
			expErr:    models.ErrAvatarNotFound,
		},
		{
			name:      "ошибка отмены контекста",
			cancelCtx: true,
			upload:    true,
			expKey:    expKey,
			expData:   nil,
			expErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorageS3()
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}
			if tt.upload {
				require.NoError(t, storage.Upload(context.Background(), expKey, expData))
			}
			data, err := storage.Download(ctx, tt.expKey)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Equal(t, tt.expData, data)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expData, data)
		})
	}
}

func TestMemStorageS3_Delete(t *testing.T) {
	expKey := "avatars/userID/image.png"
	expData := []byte("image.png")
	tests := []struct {
		name      string
		cancelCtx bool
		expErr    error
	}{
		{
			name:      "положительный тест",
			cancelCtx: false,
			expErr:    nil,
		},
		{
			name:      "ошибка отмены контекста",
			cancelCtx: true,
			expErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorageS3()
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}
			require.NoError(t, storage.Upload(context.Background(), expKey, expData))
			err := storage.Delete(ctx, expKey)
			if tt.expErr != nil {
				assert.ErrorIs(t, err, tt.expErr)
				return
			}
			require.NoError(t, err)
			_, err = storage.Download(context.Background(), expKey)
			assert.ErrorIs(t, err, models.ErrAvatarNotFound)
		})
	}
}

func TestMemStorageS3_Ping(t *testing.T) {
	tests := []struct {
		name      string
		cancelCtx bool
		expErr    error
	}{
		{
			name:      "положительный тест",
			cancelCtx: false,
			expErr:    nil,
		},
		{
			name:      "ошибка отмены контекста",
			cancelCtx: true,
			expErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorageS3()
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}
			err := storage.Ping(ctx)
			if tt.expErr != nil {
				assert.ErrorIs(t, err, tt.expErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
