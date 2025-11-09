package blob

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/timgluz/blobber/pkg/blobstore"
	"github.com/timgluz/blobber/pkg/response"
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
	case http.MethodPut, http.MethodPost:
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
		response.RenderErrorJSON(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) listBlobs(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	prefix = strings.TrimSpace(prefix)

	h.logger.Debug("Listing blobs", slog.String("prefix", prefix))
	blobs, err := h.store.List(r.Context(), prefix)
	if err != nil {
		h.logger.Error("failed to list blobs", slog.String("error", err.Error()))
		response.RenderErrorJSON(w, "Failed to list blobs", http.StatusInternalServerError)
		return
	}

	response.RenderPaginatedJSON(w, blobs, response.Pagination{
		Page:       1,
		PageSize:   len(blobs),
		TotalItems: len(blobs),
		TotalPages: 1,
	})
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

	requestedMimeType := r.Header.Get("Accept")
	if requestedMimeType != "" {
		w.Header().Set("Content-Type", requestedMimeType)
	} else {
		w.Header().Set("Content-Type", response.ContentTypeOctetStream)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *Handler) putBlob(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		response.RenderErrorJSON(w, "Key is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		response.RenderErrorJSON(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = h.store.Put(r.Context(), key, data)
	if err != nil {
		response.RenderErrorJSON(w, "Failed to store blob", http.StatusInternalServerError)
		return
	}

	response.RenderSuccessJSON(w, "Blob stored successfully", http.StatusCreated)
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
		if errors.Is(err, blobstore.ErrBlobNotFound) {
			http.Error(w, "Blob not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to delete blob", http.StatusInternalServerError)
		return
	}

	response.RenderSuccessJSON(w, "Blob deleted successfully", http.StatusNoContent)
}
