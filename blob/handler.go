package blob

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/timgluz/blobber/pkg/blobstore"
)

type Handler struct {
	store  blobstore.BlobStore
	logger *slog.Logger
}

func NewHandler(store blobstore.BlobStore, logger *slog.Logger) *Handler {
	return &Handler{store: store, logger: logger}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getBlob(w, r)
	case http.MethodPut:
		h.putBlob(w, r)
	case http.MethodDelete:
		h.deleteBlob(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listBlobs(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) listBlobs(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	prefix = strings.TrimSpace(prefix)

	h.logger.Debug("Listing blobs", slog.String("prefix", prefix))
	blobs, err := h.store.List(r.Context(), prefix)
	if err != nil {
		h.logger.Error("failed to list blobs", slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(blobs)
}

func (h *Handler) getBlob(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Fetching blob", slog.String("key", key))
	data, err := h.store.Get(r.Context(), key)
	if err != nil {
		http.Error(w, "Blob not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *Handler) putBlob(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = h.store.Put(r.Context(), key, data)
	if err != nil {
		http.Error(w, "Failed to store blob", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Blob stored successfully"))
}

func (h *Handler) deleteBlob(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Deleting blob", slog.String("key", key))
	err := h.store.Delete(r.Context(), key)
	if err != nil {
		http.Error(w, "Failed to delete blob", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("Blob deleted successfully"))
}
