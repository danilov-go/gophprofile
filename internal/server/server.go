// Package server реализует HTTP-сервер приложения.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type log interface {
	Infow(msg string, keysAndValues ...any)
}

// Server определяет конфигурацию HTTP-сервера.
type Server struct {
	Server *http.Server
	Logger log
}

// New создает новый экземпляр Server.
func New(port string, l log, r http.Handler) *Server {
	server := &http.Server{
		Addr:              port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return &Server{
		Server: server,
		Logger: l,
	}
}

// Run запускает HTTP-сервер.
func (serv *Server) Run() error {
	if err := serv.Server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
	return nil
}

// Stop останавливает HTTP-сервер.
func (serv *Server) Stop(ctx context.Context) error {
	if err := serv.Server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
