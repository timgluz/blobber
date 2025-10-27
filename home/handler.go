package home

import (
	"log/slog"
	"net/http"

	"github.com/timgluz/blobber/pkg/response"
)

type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	response.RenderSuccessJSON(w, "Home endpoint is working", http.StatusOK)
}
