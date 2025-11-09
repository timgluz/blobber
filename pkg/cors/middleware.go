package cors

import (
	"net/http"
	"strings"
)

func CORSAllowedHeaders() string {
	allowedHeaders := []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization", "X-API-Token"}

	var b strings.Builder
	for i, header := range allowedHeaders {
		b.WriteString(header)

		if i < len(allowedHeaders)-1 {
			b.WriteString(", ")
		}
	}

	return b.String()
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", CORSAllowedHeaders())
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
