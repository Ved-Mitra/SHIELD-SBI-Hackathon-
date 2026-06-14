package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const requestIDHeader = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(requestIDHeader) == "" {
			id := make([]byte, 16)
			_, _ = rand.Read(id)
			r.Header.Set(requestIDHeader, hex.EncodeToString(id))
		}
		w.Header().Set(requestIDHeader, r.Header.Get(requestIDHeader))
		next.ServeHTTP(w, r)
	})
}

