package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/danilov-go/gophprofile/internal/service"
	"github.com/danilov-go/gophprofile/internal/service/mocks"
)

func TestService_Delete(t *testing.T) {
	avatar := models.Avatar{
		ID:     avatarID,
		UserID: userID,
		S3Key:  originalKey,
		ThumbnailS3Keys: map[models.SizeAvatar]string{
			models.SizeSmall:  smallKey,
			models.SizeMedium: mediumKey,
		},
	}
	expKeys := []string{originalKey, smallKey, mediumKey}
	tests := []struct {
		name       string
		avatarID   string
		userID     string
		setupMocks func(storage *mocks.MockStorage)
		expErr     error
	}{
		{
			name:     "положительный тест",
			avatarID: avatarID,
			userID:   userID,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
				storage.On("Delete", mock.Anything, avatarID, mock.MatchedBy(func(keys []string) bool {
					return assert.ElementsMatch(t, expKeys, keys)
				})).Return(nil).Once()
			},
			expErr: nil,
		},
		{
			name:       "невалидный формат UUID",
			avatarID:   "invalid",
			userID:     userID,
			setupMocks: func(storage *mocks.MockStorage) {},
			expErr:     models.ErrAvatarNotFound,
		},
		{
			name:     "аватарка не найдена в БД",
			avatarID: avatarID,
			userID:   userID,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("Get", mock.Anything, avatarID).Return(models.Avatar{}, models.ErrAvatarNotFound).Once()
			},
			expErr: models.ErrAvatarNotFound,
		},
		{
			name:     "удаление чужой аватарки",
			avatarID: avatarID,
			userID:   unknowID,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
			},
			expErr: models.ErrForbidden,
		},
		{
			name:     "ошибка БД",
			avatarID: avatarID,
			userID:   userID,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
				storage.On("Delete", mock.Anything, avatarID, mock.Anything).Return(errTest).Once()
			},
			expErr: errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			s3 := mocks.NewMockS3(t)
			producer := mocks.NewMockProducer(t)
			converter := mocks.NewMockConverter(t)
			tt.setupMocks(storage)
			svc := service.NewService(storage, s3, producer, converter)
			err := svc.Delete(context.Background(), tt.avatarID, tt.userID)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_DeleteByUserID(t *testing.T) {
	avatar := models.Avatar{
		ID:     avatarID,
		UserID: userID,
		S3Key:  originalKey,
	}
	tests := []struct {
		name       string
		userID     string
		setupMocks func(storage *mocks.MockStorage)
		expErr     error
	}{
		{
			name:   "положительный тест",
			userID: userID,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("GetByUserID", mock.Anything, userID).Return(avatar, nil).Once()
				storage.On("Delete", mock.Anything, avatarID, []string{originalKey}).Return(nil).Once()
			},
			expErr: nil,
		},
		{
			name:   "текущая аватарка пользователя не найдена",
			userID: userID,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("GetByUserID", mock.Anything, userID).Return(models.Avatar{}, errTest).Once()
			},
			expErr: errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			s3 := mocks.NewMockS3(t)
			producer := mocks.NewMockProducer(t)
			converter := mocks.NewMockConverter(t)
			tt.setupMocks(storage)
			svc := service.NewService(storage, s3, producer, converter)
			err := svc.DeleteByUserID(context.Background(), tt.userID)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
