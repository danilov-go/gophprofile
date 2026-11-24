package worker

import (
	"encoding/json"

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
			switch getAction(message.Headers) {
			case models.EventTypeUpload:
				var event models.AvatarUploadEvent
				if err := json.Unmarshal(message.Value, &event); err != nil {
					w.logger.Errorw("ошибка десериализации", "error", err)
					session.MarkMessage(message, "")
					continue
				}
				if err := w.retry(session.Context(), func() error {
					return w.HandleUploadEvent(session.Context(), event)
				}); err != nil {
					w.logger.Errorw("ошибка обновления данных", "error", err)
				}
			case models.EventTypeDelete:
				var event models.AvatarDeleteEvent
				if err := json.Unmarshal(message.Value, &event); err != nil {
					w.logger.Errorw("ошибка десериализации", "error", err)
					continue
				}
				if err := w.retry(session.Context(), func() error {
					return w.HandleDeleteEvent(session.Context(), event)
				}); err != nil {
					w.logger.Errorw("ошибка удаления данных", "error", err)
				}
			default:
				w.logger.Errorw("неизвестный тип события", "event_type", getAction(message.Headers))
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
