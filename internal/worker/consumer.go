package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/danilov-go/gophprofile/internal/models"
)

func (w *Worker) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (w *Worker) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (w *Worker) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			var err error
			switch getAction(message.Headers) {
			case models.EventTypeUpload:
				var event models.AvatarUploadEvent
				if err = json.Unmarshal(message.Value, &event); err == nil {
					err = w.retry(session.Context(), func() error {
						return w.HandleUploadEvent(session.Context(), event)
					})
				}
			case models.EventTypeDelete:
				var event models.AvatarDeleteEvent
				if err = json.Unmarshal(message.Value, &event); err == nil {
					err = w.retry(session.Context(), func() error {
						return w.HandleDeleteEvent(session.Context(), event)
					})
				}
			default:
				err = fmt.Errorf("неизвестный тип события: %s", getAction(message.Headers))
			}
			if err != nil {
				if errors.Is(err, context.Canceled) || session.Context().Err() != nil {
					return nil
				}
				if dlq := w.producer.SendToDLQ(session.Context(), message, err); dlq != nil {
					w.logger.Errorw("ошибка отправки в DLQ", "error", dlq)
					return dlq
				}
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func getAction(headers []*sarama.RecordHeader) models.EventType {
	for _, h := range headers {
		if string(h.Key) == models.EventTypeHeader {
			return models.EventType(h.Value)
		}
	}
	return ""
}
