package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/danilov-go/gophprofile/internal/service"
	"github.com/danilov-go/gophprofile/internal/service/mocks"
)

func TestService_GetMetadata(t *testing.T) {
	now := time.Now().UTC()
	avatar := models.Avatar{
		ID:         avatarID,
		UserID:     userID,
		FileName:   fileName,
		MimeType:   "image/jpeg",
		SizeBytes:  1024,
		Dimensions: models.Dimensions{Width: 700, Height: 700},
		ThumbnailS3Keys: map[models.SizeAvatar]string{
			models.SizeSmall:  smallKey,
			models.SizeMedium: mediumKey,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	expResp := models.MetaDataResponse{
		ID:         avatarID,
		UserID:     userID,
		FileName:   fileName,
		MimeType:   "image/jpeg",
		Size:       1024,
		Dimensions: models.Dimensions{Width: 700, Height: 700},
		Thumbnails: []models.Thumbnail{
			{Size: string(models.SizeSmall), URL: "/api/v1/avatars/" + avatarID + "?size=" + string(models.SizeSmall)},
			{Size: string(models.SizeMedium), URL: "/api/v1/avatars/" + avatarID + "?size=" + string(models.SizeMedium)},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	tests := []struct {
		name       string
		id         string
		setupMocks func(storage *mocks.MockStorage)
		exp        models.MetaDataResponse
		expErr     error
	}{
		{
			name:   "положительный тест",
			id:     avatarID,
			exp:    expResp,
			expErr: nil,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
			},
		},
		{
			name:       "невалидный формат UUID",
			id:         "invalid",
			setupMocks: func(storage *mocks.MockStorage) {},
			expErr:     models.ErrAvatarNotFound,
		},
		{
			name: "ошибка БД",
			id:   avatarID,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("Get", mock.Anything, avatarID).Return(models.Avatar{}, errTest).Once()
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
			res, err := svc.GetMetadata(context.Background(), tt.id)

			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Empty(t, res)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.exp.Thumbnails, res.Thumbnails)
			tt.exp.Thumbnails = nil
			res.Thumbnails = nil
			assert.Equal(t, tt.exp, res)
		})
	}
}

func TestService_GetAvatarsByUserID(t *testing.T) {
	now := time.Now().UTC()
	avatars := []models.Avatar{
		{
			ID:         avatarID,
			UserID:     userID,
			FileName:   "1.jpg",
			MimeType:   "image/jpeg",
			SizeBytes:  512,
			Dimensions: models.Dimensions{Width: 400, Height: 400},
			ThumbnailS3Keys: map[models.SizeAvatar]string{
				models.SizeSmall: smallKey,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         avatarID2,
			UserID:     userID,
			FileName:   "2.png",
			MimeType:   "image/png",
			SizeBytes:  2048,
			Dimensions: models.Dimensions{Width: 1000, Height: 1000},
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
	expResp := []models.MetaDataResponse{
		{
			ID:         avatarID,
			UserID:     userID,
			FileName:   "1.jpg",
			MimeType:   "image/jpeg",
			Size:       512,
			Dimensions: models.Dimensions{Width: 400, Height: 400},
			Thumbnails: []models.Thumbnail{
				{Size: string(models.SizeSmall), URL: "/api/v1/avatars/" + avatarID + "?size=" + string(models.SizeSmall)},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         avatarID2,
			UserID:     userID,
			FileName:   "2.png",
			MimeType:   "image/png",
			Size:       2048,
			Dimensions: models.Dimensions{Width: 1000, Height: 1000},
			Thumbnails: []models.Thumbnail{},
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
	tests := []struct {
		name       string
		userID     string
		setupMocks func(storage *mocks.MockStorage)
		exp        []models.MetaDataResponse
		expErr     error
	}{
		{
			name:   "положительный тест",
			userID: userID,
			exp:    expResp,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("GetAllByUserID", mock.Anything, userID).Return(avatars, nil).Once()
			},
		},
		{
			name:   "пустой список аватарок",
			userID: userID,
			exp:    []models.MetaDataResponse{},
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("GetAllByUserID", mock.Anything, userID).Return([]models.Avatar{}, nil).Once()
			},
		},
		{
			name:   "ошибка БД",
			userID: userID,
			setupMocks: func(storage *mocks.MockStorage) {
				storage.On("GetAllByUserID", mock.Anything, userID).Return(nil, errTest).Once()
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
			res, err := svc.GetAvatarsByUserID(context.Background(), tt.userID)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.exp, res)
		})
	}
}
