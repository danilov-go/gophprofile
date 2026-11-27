package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/danilov-go/gophprofile/internal/service"
	"github.com/danilov-go/gophprofile/internal/service/mocks"
)

const (
	avatarID    = "1f0f950c-fabd-47cc-b72e-4a268ef169e1"
	avatarID2   = "1f0f950c-fabd-47cc-b72e-4a268ef169e2"
	userID      = "userID"
	unknowID    = "unknowID"
	originalKey = "avatars/userID/original.jpg"
	fileName    = "avatar.jpg"
)

var (
	errTest      = errors.New("test error")
	smallKey     = models.GenerateThumbnailS3Key(avatarID, models.SizeSmall)
	mediumKey    = models.GenerateThumbnailS3Key(avatarID, models.SizeMedium)
	imageData    = []byte("imageData")
	smallImg     = []byte("smallImg")
	convertedImg = []byte("convertedImg")
)

func TestService_Health(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(st *mocks.MockStorage, s3 *mocks.MockS3, pr *mocks.MockProducer)
		expResp models.HealthResponse
	}{
		{
			name: "положительный тест",
			setup: func(st *mocks.MockStorage, s3 *mocks.MockS3, pr *mocks.MockProducer) {
				st.EXPECT().Ping(mock.Anything).Return(nil)
				s3.EXPECT().Ping(mock.Anything).Return(nil)
				pr.EXPECT().Ping(mock.Anything).Return(nil)
			},
			expResp: models.HealthResponse{
				DB:     "ok",
				S3:     "ok",
				Broker: "ok",
			},
		},
		{
			name: "ошибка БД",
			setup: func(st *mocks.MockStorage, s3 *mocks.MockS3, pr *mocks.MockProducer) {
				st.EXPECT().Ping(mock.Anything).Return(errTest)
				s3.EXPECT().Ping(mock.Anything).Return(nil)
				pr.EXPECT().Ping(mock.Anything).Return(nil)
			},
			expResp: models.HealthResponse{
				DB:     errTest.Error(),
				S3:     "ok",
				Broker: "ok",
			},
		},
		{
			name: "ошибка S3",
			setup: func(st *mocks.MockStorage, s3 *mocks.MockS3, pr *mocks.MockProducer) {
				st.EXPECT().Ping(mock.Anything).Return(nil)
				s3.EXPECT().Ping(mock.Anything).Return(errTest)
				pr.EXPECT().Ping(mock.Anything).Return(nil)
			},
			expResp: models.HealthResponse{
				DB:     "ok",
				S3:     errTest.Error(),
				Broker: "ok",
			},
		},
		{
			name: "ошибка брокера",
			setup: func(st *mocks.MockStorage, s3 *mocks.MockS3, pr *mocks.MockProducer) {
				st.EXPECT().Ping(mock.Anything).Return(nil)
				s3.EXPECT().Ping(mock.Anything).Return(nil)
				pr.EXPECT().Ping(mock.Anything).Return(errTest)
			},
			expResp: models.HealthResponse{
				DB:     "ok",
				S3:     "ok",
				Broker: errTest.Error(),
			},
		},
		{
			name: "ошибка сервиса",
			setup: func(st *mocks.MockStorage, s3 *mocks.MockS3, pr *mocks.MockProducer) {
				st.EXPECT().Ping(mock.Anything).Return(errTest)
				s3.EXPECT().Ping(mock.Anything).Return(errTest)
				pr.EXPECT().Ping(mock.Anything).Return(errTest)
			},
			expResp: models.HealthResponse{
				DB:     errTest.Error(),
				S3:     errTest.Error(),
				Broker: errTest.Error(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mocks.NewMockStorage(t)
			s3 := mocks.NewMockS3(t)
			broker := mocks.NewMockProducer(t)
			conv := mocks.NewMockConverter(t)
			tt.setup(db, s3, broker)
			svc := service.NewService(db, s3, broker, conv)
			resp := svc.Health(context.Background())
			assert.Equal(t, tt.expResp, resp)
		})
	}
}
