package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/danilov-go/gophprofile/internal/broker"
	"github.com/danilov-go/gophprofile/internal/config"
	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/imaging"
	"github.com/danilov-go/gophprofile/internal/logger"
	"github.com/danilov-go/gophprofile/internal/repository/db"
	"github.com/danilov-go/gophprofile/internal/repository/s3"
	"github.com/danilov-go/gophprofile/internal/server"
	"github.com/danilov-go/gophprofile/internal/service"
	"golang.org/x/sync/errgroup"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	cfg := &config.ConfigServer{}
	if err := cfg.Parse(); err != nil {
		logger.Log.Sugar().Error("ошибка парсинга конфигурации сервера")
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
	producer, err := broker.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		logger.Log.Sugar().Errorw("ошибка инициализации брокера", "error", err)
		return err
	}
	defer func() { _ = producer.Close() }()
	converter := imaging.NewConverter()
	srvc := service.NewService(storageDB, clientS3, producer, converter)
	h := handler.NewHandler(srvc, logger.Log.Sugar(), cfg.WebPath)
	g, gCtx := errgroup.WithContext(ctx)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /web/upload", h.GetForm)
	mux.HandleFunc("GET /web/gallery/{user_id}", h.GetGallery)
	mux.HandleFunc("POST /api/v1/avatars", h.UploadAvatar)
	mux.HandleFunc("GET /api/v1/avatars/{avatar_id}", h.GetAvatar)
	mux.HandleFunc("GET /api/v1/users/{user_id}/avatar", h.GetUserAvatar)
	mux.HandleFunc("GET /api/v1/users/{user_id}/avatars", h.GetUserAvatars)
	mux.HandleFunc("GET /api/v1/avatars/{avatar_id}/metadata", h.GetAvatarMetadata)
	mux.HandleFunc("DELETE /api/v1/avatars/{avatar_id}", h.DeleteAvatar)
	mux.HandleFunc("DELETE /api/v1/users/{user_id}/avatar", h.DeleteUserAvatar)
	mux.HandleFunc("GET /health", h.Health)
	serv := server.New(cfg.Net, logger.Log.Sugar(), mux)
	g.Go(func() error {
		logger.Log.Sugar().Info("Сервер запущен по протоколу HTTP на ", cfg.Net)
		return serv.Run()
	})
	g.Go(func() error {
		<-gCtx.Done()
		logger.Log.Sugar().Info("Остановка HTTP сервера...")
		return serv.Stop()
	})
	if err := g.Wait(); err != nil {
		logger.Log.Sugar().Errorw("сервер аварийно завершил работу", "err", err)
		return err
	}
	return nil
}
