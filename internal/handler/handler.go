package handler

import (
	"net/http"
)

type Logger interface {
	Errorw(msg string, keysAndValues ...any)
}

type Service interface{}

type Handler struct {
	webPath string
	service Service
	logger  Logger
}

func NewHandler(service Service, logger Logger, webPath string) *Handler {
	return &Handler{
		webPath: webPath,
		service: service,
		logger:  logger,
	}
}

func (h *Handler) GetForm(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.webPath)
}
