package middleware

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"strings"

	"shield/gate3/internal/jwt"
)

const (
	gate2Issuer   = "shield-gate2"
	gate2Audience = "shield-gate3"
)

func Gate2Auth(gate2PublicKey *rsa.PublicKey, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gate2PublicKey == nil {
			log.Printf("[WARN] Gate2Auth: mock mode — G2-JWT check SKIPPED")
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeAuthError(w, "missing or malformed Authorization header")
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		_, err := jwt.VerifyToken(tokenStr, gate2PublicKey, gate2Issuer, gate2Audience)
		if err != nil {
			log.Printf("[WARN] Gate2Auth: invalid G2-JWT: %v", err)
			writeAuthError(w, "invalid or expired G2-JWT")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="shield-gate3"`)
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}
