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

func TestService_GetAvatar(t *testing.T) {
	avatar := models.Avatar{
		ID:       avatarID,
		UserID:   userID,
		FileName: fileName,
		MimeType: "image/png",
		S3Key:    originalKey,
		ThumbnailS3Keys: map[models.SizeAvatar]string{
			models.SizeSmall: smallKey,
		},
	}
	convertedAvatar := avatar
	convertedAvatar.MimeType = "image/webp"
	tests := []struct {
		name       string
		id         string
		size       string
		format     string
		setupMocks func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter)
		expMeta    models.Avatar
		expData    []byte
		expErr     error
	}{
		{
			name:    "положительный тест (без параметров)",
			id:      avatarID,
			expMeta: avatar,
			expData: imageData,
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
				s3.On("Download", mock.Anything, originalKey).Return(imageData, nil).Once()
			},
		},
		{
			name:    "положительный тест",
			id:      avatarID,
			size:    string(models.SizeSmall),
			format:  "webp",
			expMeta: convertedAvatar,
			expData: convertedImg,
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
				s3.On("Download", mock.Anything, smallKey).Return(smallImg, nil).Once()
				converter.On("Convert", smallImg, "webp").Return(convertedImg, "image/webp", nil).Once()
			},
		},
		{
			name: "аватарка не найдена в БД",
			id:   avatarID,
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				storage.On("Get", mock.Anything, avatarID).Return(models.Avatar{}, models.ErrAvatarNotFound).Once()
			},
			expErr: models.ErrAvatarNotFound,
		},
		{
			name: "недопустимый размер",
			id:   avatarID,
			size: "unknown",
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
			},
			expErr: models.ErrInvalidSize,
		},
		{
			name: "миниатюра еще не сгенерирована",
			id:   avatarID,
			size: string(models.SizeMedium),
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
			},
			expErr: models.ErrAvatarNotFound,
		},
		{
			name: "ошибка S3 при скачивании",
			id:   avatarID,
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
				s3.On("Download", mock.Anything, originalKey).Return(nil, errTest).Once()
			},
			expErr: errTest,
		},
		{
			name:   "ошибка конвертации изображения",
			id:     avatarID,
			format: "webp",
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				storage.On("Get", mock.Anything, avatarID).Return(avatar, nil).Once()
				s3.On("Download", mock.Anything, originalKey).Return(imageData, nil).Once()
				converter.On("Convert", imageData, "webp").Return(nil, "", errTest).Once()
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
			tt.setupMocks(storage, s3, converter)
			svc := service.NewService(storage, s3, producer, converter)
			meta, data, err := svc.GetAvatar(context.Background(), tt.id, tt.size, tt.format)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Empty(t, meta)
				assert.Nil(t, data)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expMeta, meta)
			assert.Equal(t, tt.expData, data)
		})
	}
}

func TestService_GetAvatarByUserID(t *testing.T) {
	avatar := models.Avatar{
		ID:       avatarID,
		UserID:   userID,
		FileName: fileName,
		MimeType: "image/jpeg",
		S3Key:    originalKey,
	}
	tests := []struct {
		name       string
		userID     string
		setupMocks func(storage *mocks.MockStorage, s3 *mocks.MockS3)
		expMeta    models.Avatar
		expData    []byte
		expErr     error
	}{
		{
			name:   "положительный тест",
			userID: userID,
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3) {
				storage.On("GetByUserID", mock.Anything, userID).Return(avatar, nil).Once()
				s3.On("Download", mock.Anything, originalKey).Return(imageData, nil).Once()
			},
			expMeta: avatar,
			expData: imageData,
		},
		{
			name:   "ошибка БД",
			userID: userID,
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3) {
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
			tt.setupMocks(storage, s3)
			svc := service.NewService(storage, s3, producer, converter)
			meta, data, err := svc.GetAvatarByUserID(context.Background(), tt.userID, "", "")
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Empty(t, meta)
				assert.Nil(t, data)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expMeta, meta)
			assert.Equal(t, tt.expData, data)
		})
	}
}
