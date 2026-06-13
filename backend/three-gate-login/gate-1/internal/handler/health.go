package handler

import (
	"net/http"
)

// Health handles GET /healthz.
// Returns 200 OK with body "ok" when the service is running.
// Envoy and Docker health checks hit this endpoint.
func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
