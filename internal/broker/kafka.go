// Package broker реализует взаимодействие с брокером сообщений Kafka.
package broker

import (
	"context"
	"errors"
	"strconv"

	"github.com/IBM/sarama"

	"github.com/danilov-go/gophprofile/internal/models"
)

// Producer отправляет события в Kafka.
type Producer struct {
	client       sarama.Client
	syncProducer sarama.SyncProducer
	topic        string
}

// NewProducer создает новый экземпляр Producer.
func NewProducer(addrs []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1
	client, err := sarama.NewClient(addrs, config)
	if err != nil {
		return nil, err
	}
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return &Producer{
		client:       client,
		syncProducer: p,
		topic:        topic,
	}, nil
}

// Send публикует сообщение с указанным ключом и типом события в Kafka.
func (p *Producer) Send(ctx context.Context, key string, eventType models.EventType, eventBytes []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte(models.EventTypeHeader),
				Value: []byte(eventType),
			},
		},
		Value: sarama.ByteEncoder(eventBytes),
	}
	_, _, err := p.syncProducer.SendMessage(msg)
	return err
}

// SendToDLQ отправляет необработанное сообщение в топик ошибок (DLQ).
func (p *Producer) SendToDLQ(ctx context.Context, msg *sarama.ConsumerMessage, dlq error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if msg == nil {
		return errors.New("consumer message is nil")
	}
	dlqErr := "unknown"
	if dlq != nil {
		dlqErr = dlq.Error()
	}
	headers := make([]sarama.RecordHeader, 0, len(msg.Headers)+4)
	for _, h := range msg.Headers {
		if h != nil {
			headers = append(headers, *h)
		}
	}
	headers = append(headers,
		sarama.RecordHeader{Key: []byte("x-dlq-reason"), Value: []byte(dlqErr)},
		sarama.RecordHeader{Key: []byte("x-original-topic"), Value: []byte(msg.Topic)},
		sarama.RecordHeader{Key: []byte("x-original-partition"), Value: []byte(strconv.Itoa(int(msg.Partition)))},
		sarama.RecordHeader{Key: []byte("x-original-offset"), Value: []byte(strconv.FormatInt(msg.Offset, 10))},
	)
	dlqMsg := &sarama.ProducerMessage{
		Topic:   p.topic,
		Key:     sarama.ByteEncoder(msg.Key),
		Value:   sarama.ByteEncoder(msg.Value),
		Headers: headers,
	}
	_, _, err := p.syncProducer.SendMessage(dlqMsg)
	return err
}

// Ping проверяет доступность брокера.
func (p *Producer) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.client.RefreshMetadata(p.topic)
}

// Close закрывает продюсер и клиентское соединение с Kafka.
func (p *Producer) Close() error {
	errProd := p.syncProducer.Close()
	errClient := p.client.Close()
	if errProd != nil {
		return errProd
	}
	return errClient
}
