package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/danilov-go/gophprofile/internal/worker/mocks"
)

func TestWorker_ConsumeClaim(t *testing.T) {
	avatarID := "1f0f950c-fabd-47cc-b72e-4a268ef169e1"
	userID := "userID"
	originalKey := "avatars/userID/original.jpg"
	smallKey := models.GenerateThumbnailS3Key(avatarID, models.SizeSmall)
	uploadEvent, err := json.Marshal(models.AvatarUploadEvent{AvatarID: avatarID, UserID: userID, S3Key: originalKey})
	require.NoError(t, err)
	deleteEvent, err := json.Marshal(models.AvatarDeleteEvent{AvatarID: avatarID, S3Keys: []string{originalKey}})
	require.NoError(t, err)
	tests := []struct {
		name       string
		eventType  models.EventType
		event      []byte
		setupMocks func(repo *mocks.MockStorage, s3 *mocks.MockS3, resizer *mocks.MockResizer, producer *mocks.MockProducer)
		wantMarked bool
	}{
		{
			name:       "положительный тест (EventTypeUpload)",
			eventType:  models.EventTypeUpload,
			event:      uploadEvent,
			wantMarked: true,
			setupMocks: func(repo *mocks.MockStorage, s3 *mocks.MockS3, resizer *mocks.MockResizer, _ *mocks.MockProducer) {
				imgBytes, thumbBytes := []byte("raw"), []byte("thumb")
				repo.On("Get", mock.Anything, avatarID).Return(models.Avatar{ID: avatarID, UserID: userID, S3Key: originalKey}, nil).Once()
				s3.On("Download", mock.Anything, originalKey).Return(imgBytes, nil).Once()
				resizer.On("Resize", imgBytes).Return([]models.ThumbnailData{{Size: models.SizeSmall, Data: thumbBytes}}, nil).Once()
				s3.On("Upload", mock.Anything, smallKey, thumbBytes).Return(nil).Once()
				repo.On("UpdateProcessingStatus", mock.Anything, avatarID, models.StatusCompleted, map[models.SizeAvatar]string{models.SizeSmall: smallKey}).Return(nil).Once()
			},
		},
		{
			name:       "положительный тест (EventTypeDelete)",
			eventType:  models.EventTypeDelete,
			event:      deleteEvent,
			wantMarked: true,
			setupMocks: func(_ *mocks.MockStorage, s3 *mocks.MockS3, _ *mocks.MockResizer, _ *mocks.MockProducer) {
				s3.On("Delete", mock.Anything, originalKey).Return(nil).Once()
			},
		},
		{
			name:       "неизвестный тип события",
			eventType:  "unknow",
			event:      []byte(`{}`),
			wantMarked: true,
			setupMocks: func(_ *mocks.MockStorage, _ *mocks.MockS3, _ *mocks.MockResizer, producer *mocks.MockProducer) {
				producer.On("SendToDLQ", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:       "отсутствует заголовок события",
			event:      []byte(`{}`),
			wantMarked: true,
			setupMocks: func(_ *mocks.MockStorage, _ *mocks.MockS3, _ *mocks.MockResizer, producer *mocks.MockProducer) {
				producer.On("SendToDLQ", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:       "битый JSON при загрузке",
			eventType:  models.EventTypeUpload,
			event:      []byte(`{invalid`),
			wantMarked: true,
			setupMocks: func(_ *mocks.MockStorage, _ *mocks.MockS3, _ *mocks.MockResizer, producer *mocks.MockProducer) {
				producer.On("SendToDLQ", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:       "битый JSON при удалении",
			eventType:  models.EventTypeDelete,
			event:      []byte(`{invalid`),
			wantMarked: true,
			setupMocks: func(_ *mocks.MockStorage, _ *mocks.MockS3, _ *mocks.MockResizer, producer *mocks.MockProducer) {
				producer.On("SendToDLQ", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockStorage(t)
			s3 := mocks.NewMockS3(t)
			resizer := mocks.NewMockResizer(t)
			producer := mocks.NewMockProducer(t)
			if tt.setupMocks != nil {
				tt.setupMocks(repo, s3, resizer, producer)
			}
			msg := &sarama.ConsumerMessage{Value: tt.event}
			if tt.eventType != "" {
				msg.Headers = []*sarama.RecordHeader{{Key: []byte(models.EventTypeHeader), Value: []byte(tt.eventType)}}
			}
			chanMsg := make(chan *sarama.ConsumerMessage, 1)
			chanMsg <- msg
			close(chanMsg)
			session := mocks.NewMockConsumerGroupSession(t)
			session.On("Context").Return(context.Background())
			if tt.wantMarked {
				session.On("MarkMessage", msg, "").Once()
			}
			claim := mocks.NewMockConsumerGroupClaim(t)
			claim.On("Messages").Return((<-chan *sarama.ConsumerMessage)(chanMsg))
			w := &Worker{
				repo:       repo,
				s3:         s3,
				producer:   producer,
				resizer:    resizer,
				logger:     zaptest.NewLogger(t).Sugar(),
				maxRetries: 1,
				retryDelay: time.Millisecond,
			}
			require.NoError(t, w.ConsumeClaim(session, claim))
		})
	}
}
