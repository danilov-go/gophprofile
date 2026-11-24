package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/danilov-go/gophprofile/internal/repository/memory"
)

func TestMemStorage_Save(t *testing.T) {
	now := time.Now().UTC()
	avatar := models.Avatar{
		ID:        "avatar-1",
		UserID:    "user-1",
		CreatedAt: now,
	}
	tests := []struct {
		name      string
		expMeta   models.Avatar
		cancelCtx bool
		expErr    error
	}{
		{
			name:      "положительный тест",
			expMeta:   avatar,
			cancelCtx: false,
			expErr:    nil,
		},
		{
			name:      "ошибка отмены контекста",
			expMeta:   avatar,
			cancelCtx: true,
			expErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorage()
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}
			err := storage.Save(ctx, tt.expMeta)
			if tt.expErr != nil {
				assert.ErrorIs(t, err, tt.expErr)
				return
			}
			require.NoError(t, err)
			meta, err := storage.Get(context.Background(), tt.expMeta.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expMeta, meta)
		})
	}
}

func TestMemStorage_Get(t *testing.T) {
	now := time.Now().UTC()
	deletedTime := now.Add(time.Minute)
	avatar := models.Avatar{ID: "1", UserID: "UserID", CreatedAt: now}
	deletedAvatar := models.Avatar{ID: "2", UserID: "UserID", CreatedAt: now, DeletedAt: &deletedTime}
	tests := []struct {
		name      string
		cancelCtx bool
		metaData  []models.Avatar
		expID     string
		expMeta   models.Avatar
		expErr    error
	}{
		{
			name:      "положительный тест",
			cancelCtx: false,
			metaData:  []models.Avatar{avatar},
			expID:     "1",
			expMeta:   avatar,
			expErr:    nil,
		},
		{
			name:      "запись не найдена",
			cancelCtx: false,
			metaData:  []models.Avatar{avatar},
			expID:     "unknown",
			expMeta:   models.Avatar{},
			expErr:    models.ErrAvatarNotFound,
		},
		{
			name:      "получение удаленной записи",
			cancelCtx: false,
			metaData:  []models.Avatar{deletedAvatar},
			expID:     "2",
			expMeta:   models.Avatar{},
			expErr:    models.ErrAvatarNotFound,
		},
		{
			name:      "ошибка отмены контекста",
			cancelCtx: true,
			metaData:  []models.Avatar{avatar},
			expID:     "1",
			expMeta:   models.Avatar{},
			expErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorage()
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}
			for _, m := range tt.metaData {
				require.NoError(t, storage.Save(context.Background(), m))
			}
			meta, err := storage.Get(ctx, tt.expID)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Equal(t, tt.expMeta, meta)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expMeta, meta)
		})
	}
}

func TestMemStorage_Delete(t *testing.T) {
	now := time.Now().UTC()
	deletedTime := now.Add(time.Minute)
	avatar := models.Avatar{ID: "1", UserID: "UserID", CreatedAt: now}
	deletedAvatar := models.Avatar{ID: "2", UserID: "UserID", CreatedAt: now, DeletedAt: &deletedTime}
	tests := []struct {
		name      string
		cancelCtx bool
		metaData  []models.Avatar
		deleteID  string
		expErr    error
	}{
		{
			name:      "положительный тест",
			cancelCtx: false,
			metaData:  []models.Avatar{avatar},
			deleteID:  "1",
			expErr:    nil,
		},
		{
			name:      "удаление несуществующей записи",
			cancelCtx: false,
			metaData:  nil,
			deleteID:  "unknown",
			expErr:    models.ErrAvatarNotFound,
		},
		{
			name:      "повторное удаление записи",
			cancelCtx: false,
			metaData:  []models.Avatar{deletedAvatar},
			deleteID:  "2",
			expErr:    models.ErrAvatarNotFound,
		},
		{
			name:      "ошибка отмены контекста",
			cancelCtx: true,
			metaData:  []models.Avatar{avatar},
			deleteID:  "1",
			expErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorage()
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}
			for _, m := range tt.metaData {
				require.NoError(t, storage.Save(context.Background(), m))
			}
			err := storage.Delete(ctx, tt.deleteID)
			if tt.expErr != nil {
				assert.ErrorIs(t, err, tt.expErr)
				return
			}
			require.NoError(t, err)
			_, err = storage.Get(context.Background(), tt.deleteID)
			assert.ErrorIs(t, err, models.ErrAvatarNotFound)
		})
	}
}

func TestMemStorage_GetAllByUserID(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	deletedTime := baseTime.Add(time.Hour)
	meta1 := models.Avatar{ID: "1", UserID: "UserID", CreatedAt: baseTime}
	meta2 := models.Avatar{ID: "2", UserID: "UserID", CreatedAt: baseTime.Add(10 * time.Minute)}
	meta3 := models.Avatar{ID: "3", UserID: "UserID", CreatedAt: baseTime.Add(20 * time.Minute)}
	meta4 := models.Avatar{ID: "4", UserID: "UserID", CreatedAt: baseTime.Add(30 * time.Minute), DeletedAt: &deletedTime}
	meta5 := models.Avatar{ID: "5", UserID: "UserID2", CreatedAt: baseTime}
	metaData := []models.Avatar{meta1, meta2, meta3, meta4, meta5}
	tests := []struct {
		name      string
		cancelCtx bool
		userID    string
		want      []models.Avatar
		expErr    error
	}{
		{
			name:      "положительный тест",
			cancelCtx: false,
			userID:    "UserID",
			want:      []models.Avatar{meta3, meta2, meta1},
			expErr:    nil,
		},
		{
			name:      "пользователь без аватарок",
			cancelCtx: false,
			userID:    "unknown",
			want:      nil,
			expErr:    nil,
		},
		{
			name:      "ошибка отмены контекста",
			cancelCtx: true,
			userID:    "UserID",
			want:      nil,
			expErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorage()
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}

			for _, m := range metaData {
				require.NoError(t, storage.Save(context.Background(), m))
			}

			meta, err := storage.GetAllByUserID(ctx, tt.userID)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Equal(t, tt.want, meta)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, meta)
		})
	}
}

func TestMemStorage_GetByUserID(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	deletedTime := baseTime.Add(time.Hour)
	meta1 := models.Avatar{ID: "1", UserID: "UserID", CreatedAt: baseTime}
	meta2 := models.Avatar{ID: "2", UserID: "UserID", CreatedAt: baseTime.Add(10 * time.Minute)}
	meta3 := models.Avatar{ID: "3", UserID: "UserID", CreatedAt: baseTime.Add(20 * time.Minute)}
	meta4 := models.Avatar{ID: "4", UserID: "UserID", CreatedAt: baseTime.Add(30 * time.Minute), DeletedAt: &deletedTime}
	meta5 := models.Avatar{ID: "5", UserID: "UserIDDeleted", CreatedAt: baseTime, DeletedAt: &deletedTime}
	metaData := []models.Avatar{meta1, meta2, meta3, meta4, meta5}
	tests := []struct {
		name      string
		cancelCtx bool
		userID    string
		expMeta   models.Avatar
		expErr    error
	}{
		{
			name:      "положительный тест",
			cancelCtx: false,
			userID:    "UserID",
			expMeta:   meta3,
			expErr:    nil,
		},
		{
			name:      "у пользователя только удаленные аватарки",
			cancelCtx: false,
			userID:    "UserIDDeleted",
			expMeta:   models.Avatar{},
			expErr:    models.ErrAvatarNotFound,
		},
		{
			name:      "пользователь не найден",
			cancelCtx: false,
			userID:    "unknown",
			expMeta:   models.Avatar{},
			expErr:    models.ErrAvatarNotFound,
		},
		{
			name:      "ошибка отмены контекста",
			cancelCtx: true,
			userID:    "UserID",
			expMeta:   models.Avatar{},
			expErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorage()
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}
			for _, m := range metaData {
				require.NoError(t, storage.Save(context.Background(), m))
			}
			meta, err := storage.GetByUserID(ctx, tt.userID)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Equal(t, tt.expMeta, meta)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expMeta, meta)
		})
	}
}

func TestMemStorage_Ping(t *testing.T) {
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
			storage := memory.NewMemStorage()
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
