package broker

import (
	"context"

	"github.com/IBM/sarama"
)

type Producer struct {
	syncProducer sarama.SyncProducer
	topic        string
}

func NewProducer(addrs []string, topic string) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	p, err := sarama.NewSyncProducer(addrs, cfg)
	if err != nil {
		return nil, err
	}
	return &Producer{
		syncProducer: p,
		topic:        topic,
	}, nil
}

func (p *Producer) Send(ctx context.Context, key, eventType string, eventBytes []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(eventType),
			},
		},
		Value: sarama.ByteEncoder(eventBytes),
	}
	_, _, err := p.syncProducer.SendMessage(msg)
	return err
}

func (p *Producer) Close() error {
	return p.syncProducer.Close()
}
