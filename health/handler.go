package health

import (
	"log/slog"
	"net/http"

	"github.com/timgluz/blobber/pkg/blobstore"
	"github.com/timgluz/blobber/pkg/response"
)

type Handler struct {
	store  blobstore.BlobStore
	logger *slog.Logger
}

func NewHandler(store blobstore.BlobStore, logger *slog.Logger) *Handler {
	return &Handler{store, logger}
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(r.Context()); err != nil {
		h.logger.Error("Blob store ping failed", slog.String("error", err.Error()))
		response.RenderErrorJSON(w, "Blob store is unreachable", http.StatusInternalServerError)
		return
	}

	response.RenderSuccessJSON(w, "Health endpoint is working", http.StatusOK)
}

func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	response.RenderSuccessJSON(w, "Ready endpoint is working", http.StatusOK)
}
