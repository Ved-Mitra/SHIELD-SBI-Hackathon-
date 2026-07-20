package middleware

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps ResponseWriter to capture the status code for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging wraps a handler with structured access logging.
// Each request is logged with method, path, status code, latency, and request ID.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		log.Printf("gate1 %s %s → %d (%s) req_id=%s",
			r.Method, r.URL.Path, rec.status, time.Since(start),
			r.Header.Get("X-Request-ID"),
		)
	})
}
