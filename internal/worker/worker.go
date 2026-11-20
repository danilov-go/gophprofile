package worker

import (
	"context"
	"errors"
	"time"

	"github.com/IBM/sarama"
	"github.com/danilov-go/gophprofile/internal/models"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

type Storage interface {
	Get(ctx context.Context, id string) (models.Avatar, error)
	UpdateProcessingStatus(ctx context.Context, id string, status models.ProcessingStatus, thumbnailKeys map[models.SizeAvatar]string) error
}

type S3 interface {
	Download(ctx context.Context, s3Key string) ([]byte, error)
	Upload(ctx context.Context, s3Key string, imageData []byte) error
	Delete(ctx context.Context, s3Key string) error
}

type Resizer interface {
	Resize(imageData []byte) ([]models.ThumbnailData, error)
}

type Config struct {
	Brokers    []string
	GroupID    string
	Topic      string
	MaxRetries int
	RetryDelay time.Duration
}

type Worker struct {
	repo          Storage
	s3            S3
	resizer       Resizer
	consumerGroup sarama.ConsumerGroup
	topic         string
	logger        log
	maxRetries    int
	retryDelay    time.Duration
	interval      time.Duration
}

func NewWorker(cfg Config, repo Storage, s3 S3, resizer Resizer, l log) (*Worker, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, err
	}
	return &Worker{
		repo:          repo,
		s3:            s3,
		resizer:       resizer,
		consumerGroup: consumer,
		topic:         cfg.Topic,
		logger:        l,
		maxRetries:    cfg.MaxRetries,
		retryDelay:    cfg.RetryDelay,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	defer func() { _ = w.consumerGroup.Close() }()
	for {
		if err := w.consumerGroup.Consume(ctx, []string{w.topic}, w); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			w.logger.Errorw("ошибка consumer group", "error", err)
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}
