package publisher

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/gophprofile/internal/models"
	"github.com/danilov-go/gophprofile/internal/publisher/mocks"
)

var errTest = errors.New("error test")

const (
	eventID1  = "eventID1"
	eventID2  = "eventID2"
	avatarID1 = "avatarID1"
	avatarID2 = "avatarID2"
)

func TestEventPublisher_Publish(t *testing.T) {
	events := []models.Event{
		{ID: eventID1, AvatarID: avatarID1, EventType: models.EventTypeUpload, Payload: []byte("event")},
		{ID: eventID2, AvatarID: avatarID2, EventType: models.EventTypeDelete, Payload: []byte("event")},
	}
	tests := []struct {
		name      string
		setupMock func(st *mocks.MockStorage, pr *mocks.MockProducer)
	}{
		{
			name: "положительный тест",
			setupMock: func(st *mocks.MockStorage, pr *mocks.MockProducer) {
				st.EXPECT().GetEvents(mock.Anything, 50).Return(events, nil).Once()
				pr.EXPECT().Send(mock.Anything, avatarID1, models.EventTypeUpload, []byte("event")).Return(nil).Once()
				st.EXPECT().UpdateEventStatus(mock.Anything, eventID1).Return(nil).Once()
				pr.EXPECT().Send(mock.Anything, avatarID2, models.EventTypeDelete, []byte("event")).Return(nil).Once()
				st.EXPECT().UpdateEventStatus(mock.Anything, eventID2).Return(nil).Once()
			},
		},
		{
			name: "ошибка чтения событий из БД",
			setupMock: func(st *mocks.MockStorage, _ *mocks.MockProducer) {
				st.EXPECT().GetEvents(mock.Anything, 50).Return(nil, errTest).Once()
			},
		},
		{
			name: "ошибка отправки в Kafka",
			setupMock: func(st *mocks.MockStorage, pr *mocks.MockProducer) {
				st.EXPECT().GetEvents(mock.Anything, 50).Return(events, nil).Once()
				pr.EXPECT().Send(mock.Anything, avatarID1, models.EventTypeUpload, []byte("event")).Return(errTest).Once()
			},
		},
		{
			name: "ошибка обновления статуса в БД",
			setupMock: func(st *mocks.MockStorage, pr *mocks.MockProducer) {
				st.EXPECT().GetEvents(mock.Anything, 50).Return(events, nil).Once()
				pr.EXPECT().Send(mock.Anything, avatarID1, models.EventTypeUpload, []byte("event")).Return(nil).Once()
				st.EXPECT().UpdateEventStatus(mock.Anything, eventID1).Return(errTest).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			producer := mocks.NewMockProducer(t)
			logger := zaptest.NewLogger(t).Sugar()
			tt.setupMock(storage, producer)
			publisher := NewEventPublisher(storage, producer, logger, 10*time.Millisecond, 50)
			publisher.Publish(context.Background())
		})
	}
}

func TestEventPublisher_Run(t *testing.T) {
	storage := mocks.NewMockStorage(t)
	producer := mocks.NewMockProducer(t)
	logger := zaptest.NewLogger(t).Sugar()
	storage.EXPECT().GetEvents(mock.Anything, 50).Return(nil, nil).Maybe()
	publisher := NewEventPublisher(storage, producer, logger, 10*time.Millisecond, 50)
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()
	assert.NoError(t, publisher.Run(ctx))
}
