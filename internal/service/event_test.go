package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/danilov-go/gophprofile/internal/service"
	"github.com/danilov-go/gophprofile/internal/service/mocks"
)

func TestService_PublishUploadEvent(t *testing.T) {
	expEvent, err := json.Marshal(models.AvatarUploadEvent{
		AvatarID: avatarID,
		UserID:   userID,
		S3Key:    originalKey,
	})
	require.NoError(t, err)
	tests := []struct {
		name       string
		setupMocks func(producer *mocks.MockProducer)
		expErr     error
	}{
		{
			name: "положительный тест",
			setupMocks: func(producer *mocks.MockProducer) {
				producer.On("Send", mock.Anything, avatarID, models.EventTypeUpload, expEvent).
					Return(nil).Once()
			},
			expErr: nil,
		},
		{
			name: "ошибка брокера",
			setupMocks: func(producer *mocks.MockProducer) {
				producer.On("Send", mock.Anything, avatarID, models.EventTypeUpload, expEvent).
					Return(errTest).Once()
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
			tt.setupMocks(producer)
			svc := service.NewService(storage, s3, producer, converter)
			err := svc.PublishUploadEvent(context.Background(), avatarID, userID, originalKey)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_PublishDeleteEvent(t *testing.T) {
	s3Keys := []string{originalKey, smallKey, mediumKey}
	expEvent, err := json.Marshal(models.AvatarDeleteEvent{
		AvatarID: avatarID,
		S3Keys:   s3Keys,
	})
	require.NoError(t, err)
	tests := []struct {
		name       string
		setupMocks func(producer *mocks.MockProducer)
		expErr     error
	}{
		{
			name: "положительный тест",
			setupMocks: func(producer *mocks.MockProducer) {
				producer.On("Send", mock.Anything, avatarID, models.EventTypeDelete, expEvent).
					Return(nil).Once()
			},
			expErr: nil,
		},
		{
			name: "ошибка брокера",
			setupMocks: func(producer *mocks.MockProducer) {
				producer.On("Send", mock.Anything, avatarID, models.EventTypeDelete, expEvent).
					Return(errTest).Once()
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
			tt.setupMocks(producer)
			svc := service.NewService(storage, s3, producer, converter)
			err := svc.PublishDeleteEvent(context.Background(), avatarID, s3Keys)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
