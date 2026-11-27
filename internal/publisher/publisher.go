// Package publisher реализует фоновую отправку событий из хранилища в брокер сообщений.
package publisher

import (
	"context"
	"errors"
	"time"

	"github.com/danilov-go/gophprofile/internal/models"
)

// Storage определяет методы взаимодействия с хранилищем.
type Storage interface {
	GetEvents(ctx context.Context, limit int) ([]models.Event, error)
	UpdateEventStatus(ctx context.Context, id string) error
}

// Producer определяет методы отправки событий в брокер сообщений.
type Producer interface {
	Send(ctx context.Context, key string, eventType models.EventType, eventBytes []byte) error
}

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

// EventPublisher публикует события из базы данных в Kafka.
type EventPublisher struct {
	storage      Storage
	producer     Producer
	logger       log
	pollInterval time.Duration
	eventLimit   int
}

// NewEventPublisher создает новый экземпляр EventPublisher.
func NewEventPublisher(storage Storage, producer Producer, logger log, interval time.Duration, limit int) *EventPublisher {
	return &EventPublisher{
		storage:      storage,
		producer:     producer,
		logger:       logger,
		pollInterval: interval,
		eventLimit:   limit,
	}
}

// Run запускает цикл отправки событий.
func (p *EventPublisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			p.Publish(ctx)
		}
	}
}

// Publish читает неотправленные события из хранилища, публикует их в брокер сообщений и обновляет статус отправки.
func (p *EventPublisher) Publish(ctx context.Context) {
	events, err := p.storage.GetEvents(ctx, p.eventLimit)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			p.logger.Errorw("ошибка чтения событий", "error", err)
		}
		return
	}
	for _, event := range events {
		if err := p.producer.Send(ctx, event.AvatarID, event.EventType, event.Payload); err != nil {
			if !errors.Is(err, context.Canceled) {
				p.logger.Errorw("ошибка отправки события", "error", err)
			}
			return
		}
		if err := p.storage.UpdateEventStatus(ctx, event.ID); err != nil {
			p.logger.Errorw("ошибка обновления статуса события", "error", err)
			return
		}
	}
}
