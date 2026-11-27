package broker

import (
	"context"
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/require"

	"github.com/danilov-go/gophprofile/internal/models"
)

var errTest = errors.New("error test")

const avatarID = "1f0f950c-fabd-47cc-b72e-4a268ef169e1"

func TestProducer_Send(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name      string
		ctx       context.Context
		setupMock func(sp *mocks.SyncProducer)
		expErr    error
	}{
		{
			name:      "положительный тест",
			ctx:       context.Background(),
			setupMock: func(sp *mocks.SyncProducer) { sp.ExpectSendMessageAndSucceed() },
		},
		{
			name:      "ошибка брокера при отправке",
			ctx:       context.Background(),
			setupMock: func(sp *mocks.SyncProducer) { sp.ExpectSendMessageAndFail(errTest) },
			expErr:    errTest,
		},
		{
			name:   "ошибка отмены контекста",
			ctx:    canceledCtx,
			expErr: context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncProducer := mocks.NewSyncProducer(t, nil)
			t.Cleanup(func() { require.NoError(t, syncProducer.Close()) })
			if tt.setupMock != nil {
				tt.setupMock(syncProducer)
			}
			producer := &Producer{syncProducer: syncProducer, topic: "avatar-events"}
			err := producer.Send(tt.ctx, avatarID, models.EventTypeUpload, []byte("event"))
			require.ErrorIs(t, err, tt.expErr)
		})
	}
}

func TestProducer_SendToDLQ(t *testing.T) {
	msg := &sarama.ConsumerMessage{
		Topic:  "avatar-events",
		Offset: 15,
	}
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name      string
		ctx       context.Context
		setupMock func(sp *mocks.SyncProducer)
		expErr    error
	}{
		{
			name:      "успешная отправка",
			ctx:       context.Background(),
			setupMock: func(sp *mocks.SyncProducer) { sp.ExpectSendMessageAndSucceed() },
		},
		{
			name:      "ошибка брокера при отправке",
			ctx:       context.Background(),
			setupMock: func(sp *mocks.SyncProducer) { sp.ExpectSendMessageAndFail(errTest) },
			expErr:    errTest,
		},
		{
			name:   "отмененный контекст",
			ctx:    canceledCtx,
			expErr: context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncProducer := mocks.NewSyncProducer(t, nil)
			t.Cleanup(func() { require.NoError(t, syncProducer.Close()) })
			if tt.setupMock != nil {
				tt.setupMock(syncProducer)
			}
			producer := &Producer{syncProducer: syncProducer, topic: "avatar-events-dlq"}
			err := producer.SendToDLQ(tt.ctx, msg, errTest)
			require.ErrorIs(t, err, tt.expErr)
		})
	}
}
