package server

import (
	"encoding/json"
	"net/http"

	"github.com/nomos/nomos/internal/app"
)

// apiKeyAuth wraps next so that requests carrying X-API-Key are validated
// against the cosmos key store. Requests without the header pass through
// unchanged (Traefik / another upstream proxy handles their auth).
func apiKeyAuth(cosmosPath string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !app.ValidateKey(cosmosPath, key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED", "error": "invalid API key"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
