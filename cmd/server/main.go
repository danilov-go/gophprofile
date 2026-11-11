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
	"github.com/danilov-go/gophprofile/internal/repository/memory"
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
	var storage service.Storage
	if cfg.DatabaseDSN != "" {
		db, err := db.InitDB(cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Sugar().Fatalw("ошибка инициализации базы данных", "error", err)
		}
		storage = db
	} else {
		storage = memory.NewMemStorage()
	}
	srvс := service.NewService(storage)
	h := handler.NewHandler(srvс, logger.Log.Sugar(), cfg.WebPath)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	g, gCtx := errgroup.WithContext(ctx)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /web/upload", h.GetForm)
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
