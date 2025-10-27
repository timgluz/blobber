package secret

import (
	"log/slog"
	"net/http"
)

type APITokenMiddleware struct {
	store  SecretStore
	logger *slog.Logger
}

func NewAPITokenMiddleware(store SecretStore, logger *slog.Logger) *APITokenMiddleware {
	return &APITokenMiddleware{store: store, logger: logger}
}

func (m *APITokenMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-API-Token")
		if token == "" {
			http.Error(w, "Missing API token", http.StatusUnauthorized)
			return
		}

		valid, err := m.store.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Error validating API token", http.StatusInternalServerError)
			return
		}

		if !valid {
			http.Error(w, "Invalid API token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
