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

func TestService_SaveAvatar(t *testing.T) {
	dims := models.Dimensions{Width: 100, Height: 100}
	tests := []struct {
		name       string
		data       []byte
		mimeType   string
		setupMocks func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter)
		expErr     error
		checkRes   func(t *testing.T, res models.Avatar)
	}{
		{
			name:     "положительный тест",
			data:     imageData,
			mimeType: "image/jpeg",
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				converter.On("GetDimensions", imageData).Return(dims, nil).Once()
				s3.On("Upload", mock.Anything, mock.AnythingOfType("string"), imageData).Return(nil).Once()
				storage.On("Save", mock.Anything, mock.MatchedBy(func(m models.Avatar) bool {
					return m.UserID == userID && m.FileName == fileName && m.S3Key == "avatars/"+userID+"/"+m.ID+".jpg"
				})).Return(nil).Once()
			},
			checkRes: func(t *testing.T, res models.Avatar) {
				t.Helper()
				assert.NotEmpty(t, res.ID)
				assert.Equal(t, "avatars/"+userID+"/"+res.ID+".jpg", res.S3Key)
				assert.Equal(t, userID, res.UserID)
				assert.Equal(t, fileName, res.FileName)
				assert.Equal(t, "image/jpeg", res.MimeType)
				assert.Equal(t, int64(len(imageData)), res.SizeBytes)
				assert.Equal(t, dims, res.Dimensions)
				assert.Equal(t, models.StatusProcessing, res.ProcessingStatus)
				assert.Equal(t, models.UploadStatusUploaded, res.UploadStatus)
				assert.False(t, res.CreatedAt.IsZero())
				assert.Equal(t, res.CreatedAt, res.UpdatedAt)
			},
		},
		{
			name:     "превышен максимальный размер файла",
			data:     make([]byte, models.MaxSize+1),
			mimeType: "image/jpeg",
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
			},
			expErr: models.ErrFileTooLarge,
		},
		{
			name:     "неподдерживаемый mimeType",
			data:     imageData,
			mimeType: "application/pdf",
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
			},
			expErr: models.ErrInvalidFormat,
		},
		{
			name:     "ошибка получения размеров",
			data:     imageData,
			mimeType: "image/jpeg",
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				converter.On("GetDimensions", imageData).Return(models.Dimensions{}, errTest).Once()
			},
			expErr: errTest,
		},
		{
			name:     "ошибка S3 при загрузке",
			data:     imageData,
			mimeType: "image/jpeg",
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				converter.On("GetDimensions", imageData).Return(dims, nil).Once()
				s3.On("Upload", mock.Anything, mock.AnythingOfType("string"), imageData).Return(errTest).Once()
			},
			expErr: errTest,
		},
		{
			name:     "ошибка БД при сохранении — удаление файла из S3",
			data:     imageData,
			mimeType: "image/jpeg",
			setupMocks: func(storage *mocks.MockStorage, s3 *mocks.MockS3, converter *mocks.MockConverter) {
				converter.On("GetDimensions", imageData).Return(dims, nil).Once()
				s3.On("Upload", mock.Anything, mock.AnythingOfType("string"), imageData).Return(nil).Once()
				storage.On("Save", mock.Anything, mock.AnythingOfType("models.Avatar")).Return(errTest).Once()
				s3.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
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
			res, err := svc.SaveAvatar(context.Background(), tt.data, userID, fileName, tt.mimeType)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Empty(t, res)
				return
			}
			require.NoError(t, err)
			tt.checkRes(t, res)
		})
	}
}
