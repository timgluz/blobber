package home

import (
	"log/slog"
	"net/http"
	"text/template"
)

type HandlerData struct {
	Title        string
	Version      string
	BlobProvider string
}

type Handler struct {
	data HandlerData

	logger *slog.Logger
}

func NewHandler(handlerData HandlerData, logger *slog.Logger) *Handler {
	return &Handler{
		data:   handlerData,
		logger: logger,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/home.html"))

	if err := tmpl.Execute(w, h.data); err != nil {
		h.logger.Error("failed to render home page", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
