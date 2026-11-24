package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"

	"github.com/danilov-go/gophprofile/internal/broker"
	"github.com/danilov-go/gophprofile/internal/config"
	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/imaging"
	"github.com/danilov-go/gophprofile/internal/logger"
	"github.com/danilov-go/gophprofile/internal/repository/db"
	"github.com/danilov-go/gophprofile/internal/repository/s3"
	"github.com/danilov-go/gophprofile/internal/server"
	"github.com/danilov-go/gophprofile/internal/service"
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
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.Net},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-User-ID"},
		ExposedHeaders:   []string{"ETag"},
		AllowCredentials: false,
		MaxAge:           300,
	})
	r.Use(c.Handler)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(httprate.LimitBy(100, time.Minute, func(r *http.Request) (string, error) {
		return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
	}))
	r.Get("/web/upload", h.GetForm)
	r.Get("/web/gallery/{user_id}", h.GetGallery)
	r.Get("/health", h.Health)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/avatars/{avatar_id}", h.GetAvatar)
		r.Get("/avatars/{avatar_id}/metadata", h.GetAvatarMetadata)
		r.Get("/users/{user_id}/avatar", h.GetUserAvatar)
		r.Get("/users/{user_id}/avatars", h.GetUserAvatars)
		r.Post("/avatars", h.ValidateUserID(h.UploadAvatar))
		r.Delete("/avatars/{avatar_id}", h.ValidateUserID(h.DeleteAvatar))
		r.Delete("/users/{user_id}/avatar", h.ValidateUserID(h.DeleteUserAvatar))
	})
	serv := server.New(cfg.Net, logger.Log.Sugar(), r)
	g.Go(func() error {
		logger.Log.Sugar().Info("Сервер запущен на ", cfg.Net)
		return serv.Run()
	})
	g.Go(func() error {
		<-gCtx.Done()
		logger.Log.Sugar().Info("Остановка HTTP сервера...")
		return serv.Stop(ctx)
	})
	if err := g.Wait(); err != nil {
		logger.Log.Sugar().Errorw("сервер аварийно завершил работу", "err", err)
		return err
	}
	return nil
}
