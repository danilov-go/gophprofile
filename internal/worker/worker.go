package worker

import (
	"context"
	"errors"

	"github.com/IBM/sarama"
	"github.com/danilov-go/gophprofile/internal/models"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

type Storage interface {
	Get(ctx context.Context, id string) (models.Avatar, error)
	UpdateProcessingStatus(ctx context.Context, id, status string, thumbnailKeys map[models.SizeAvatar]string) error
}

type S3 interface {
	Download(ctx context.Context, s3Key string) ([]byte, error)
	Upload(ctx context.Context, s3Key string, imageData []byte) error
	Delete(ctx context.Context, s3Key string) error
}

type Resizer interface {
	Resize(imageData []byte) ([]models.ThumbnailData, error)
}

type Worker struct {
	repo          Storage
	s3            S3
	resizer       Resizer
	consumerGroup sarama.ConsumerGroup
	topic         string
	logger        log
}

func NewWorker(brokers []string, groupID, topic string, repo Storage, s3 S3, resizer Resizer, l log) (*Worker, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cg, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}
	return &Worker{
		repo:          repo,
		s3:            s3,
		resizer:       resizer,
		consumerGroup: cg,
		topic:         topic,
		logger:        l,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	defer w.consumerGroup.Close()
	for {
		if err := w.consumerGroup.Consume(ctx, []string{w.topic}, w); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}
