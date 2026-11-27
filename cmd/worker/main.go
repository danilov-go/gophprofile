package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/danilov-go/gophprofile/internal/broker"
	"github.com/danilov-go/gophprofile/internal/config"
	"github.com/danilov-go/gophprofile/internal/imaging"
	"github.com/danilov-go/gophprofile/internal/logger"
	"github.com/danilov-go/gophprofile/internal/repository/db"
	"github.com/danilov-go/gophprofile/internal/repository/s3"
	"github.com/danilov-go/gophprofile/internal/worker"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	config.PrintBuild(os.Stdout, buildVersion, buildDate, buildCommit)
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := &config.ConfigWorker{}
	if err := cfg.Parse(); err != nil {
		logger.Log.Sugar().Error("ошибка парсинга конфигурации воркера")
		return err
	}
	if err := cfg.Validate(); err != nil {
		logger.Log.Sugar().Errorw("ошибка валидации конфигурации воркера", "error", err)
		return err
	}
	storageDB, err := db.InitDB(cfg.DatabaseDSN)
	if err != nil {
		logger.Log.Sugar().Errorw("ошибка инициализации базы данных", "error", err)
		return err
	}
	defer func() { _ = storageDB.Close() }()
	clientS3, err := s3.NewS3Client(ctx, cfg.S3)
	if err != nil {
		logger.Log.Sugar().Errorw("ошибка инициализации S3 клиента", "error", err)
		return err
	}
	producer, err := broker.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic+"-dlq")
	if err != nil {
		logger.Log.Sugar().Errorw("ошибка инициализации брокера", "error", err)
		return err
	}
	config := worker.Config{
		Brokers:    cfg.Kafka.Brokers,
		GroupID:    cfg.Kafka.GroupID,
		Topic:      cfg.Kafka.Topic,
		MaxRetries: cfg.MaxRetries,
		RetryDelay: cfg.RetryDelay,
	}
	resizer := imaging.NewResizer()
	w, err := worker.NewWorker(config, storageDB, clientS3, producer, resizer, logger.Log.Sugar())
	if err != nil {
		logger.Log.Sugar().Errorw("ошибка инициализации воркера", "error", err)
		return err
	}
	if err := w.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Log.Sugar().Errorw("ошибка работы воркера", "error", err)
		return err
	}
	return nil
}
