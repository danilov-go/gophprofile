package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/danilov-go/gophprofile/internal/config"
	"github.com/danilov-go/gophprofile/internal/handler"
	"github.com/danilov-go/gophprofile/internal/logger"
	"github.com/danilov-go/gophprofile/internal/repository/db"
	"github.com/danilov-go/gophprofile/internal/repository/s3"
	"github.com/danilov-go/gophprofile/internal/server"
	"github.com/danilov-go/gophprofile/internal/service"
	"golang.org/x/sync/errgroup"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	config.PrintBuild(os.Stdout, buildVersion, buildDate, buildCommit)
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	cfg := &config.ConfigServer{}
	if err := cfg.Get(); err != nil {
		logger.Log.Sugar().Fatal("ошибка парсинга конфигурации сервера")
	}
	if err := cfg.Validate(); err != nil {
		logger.Log.Sugar().Fatal(err)
	}
	storageDB, err := db.InitDB(cfg.DatabaseDSN)
	if err != nil {
		logger.Log.Sugar().Fatalw("ошибка инициализации базы данных", "error", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	clientS3, err := s3.NewS3Client(ctx, cfg.S3)
	if err != nil {
		logger.Log.Sugar().Fatalw("ошибка инициализации S3 клиента", "error", err)
	}
	srvс := service.NewService(storageDB, clientS3)
	h := handler.NewHandler(srvс, logger.Log.Sugar(), cfg.WebPath)
	g, gCtx := errgroup.WithContext(ctx)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /web/upload", h.GetForm)
	mux.HandleFunc("POST /api/v1/avatars", h.APIUploadHandler)
	mux.HandleFunc("GET /api/v1/avatars/{avatar_id}", h.APIGetAvatarHandler)
	mux.HandleFunc("GET /api/v1/avatars/{avatar_id}/metadata", h.APIGetMetaDataHandler)
	mux.HandleFunc("DELETE /api/v1/avatars/{avatar_id}", h.APIDeleteAvatarHandler)
	mux.HandleFunc("GET /health", h.HealthHandler)
	serv := server.New(cfg.Net.String(), logger.Log.Sugar(), mux)
	g.Go(func() error {
		logger.Log.Sugar().Info("Сервер запущен по протоколу HTTP на ", cfg.Net.String())
		return serv.Run()
	})
	g.Go(func() error {
		<-gCtx.Done()
		logger.Log.Sugar().Info("Остановка HTTP сервера...")
		return serv.Stop()
	})
	if err := g.Wait(); err != nil {
		logger.Log.Sugar().Fatal("сервер аварийно завершил работу", "err", err)
	}
}
