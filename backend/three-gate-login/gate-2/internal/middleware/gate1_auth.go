// Package middleware provides HTTP middleware for the Gate 2 service.
//
// Gate1Auth validates the G1-JWT that the mobile client must present in the
// Authorization header when calling POST /gate2/token.
//
// This enforces the sequential gate ordering:
//   - A client that never passed Gate 1 has no G1-JWT → rejected here.
//   - A client with an expired G1-JWT → rejected here.
//   - Only a client that holds a valid, unexpired G1-JWT from shield-gate1 → allowed through.
package middleware

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"shield/three-gate-login/internal/kafka"
	"strings"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const (
	gate1Issuer   = "shield-gate1"
	gate1Audience = "shield-gate2"
)

// Gate1Auth returns middleware that validates the Bearer G1-JWT in the
// Authorization header using Gate 1's RSA public key.
//
// On success the request is forwarded to next unchanged.
// On failure: 401 Unauthorized with a JSON error body.
//
// When gate1PublicKey is nil (mock mode) the check is skipped entirely
// and a warning is logged per request.
func Gate1Auth(gate1PublicKey *rsa.PublicKey, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ── Mock mode: gate 1 key not loaded ─────────────────────────────────
		if gate1PublicKey == nil {
			log.Printf("[WARN] Gate1Auth: mock mode — G1-JWT check SKIPPED (req_id=%s)",
				r.Header.Get("X-Request-ID"))
			next.ServeHTTP(w, r)
			return
		}

		// ── Extract Bearer token ──────────────────────────────────────────────
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeAuthError(w, "missing or malformed Authorization header")
			go kafka.PublishEvent(kafka.AuthEvent{UserID: "unknown", Gate: 2, Status: "FAILED", Reason: "Missing or malformed Authorization header", TimeStamp: time.Now().UnixMilli()})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// ── Validate G1-JWT (RS256, iss=shield-gate1, aud=shield-gate2) ──────
		if err := validateG1JWT(tokenStr, gate1PublicKey); err != nil {
			log.Printf("[WARN] Gate1Auth: invalid G1-JWT: %v (req_id=%s)",
				err, r.Header.Get("X-Request-ID"))
			writeAuthError(w, "invalid or expired G1-JWT")
			go kafka.PublishEvent(kafka.AuthEvent{UserID: "unknown", Gate: 2, Status: "FAILED", Reason: "INvalid or Expired G1-JWT", TimeStamp: time.Now().UnixMilli()})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateG1JWT parses and validates the G1-JWT claims.
func validateG1JWT(tokenStr string, pubKey *rsa.PublicKey) error {
	claims := &jwtlib.RegisteredClaims{}

	_, err := jwtlib.ParseWithClaims(tokenStr, claims,
		func(t *jwtlib.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwtlib.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
			}
			return pubKey, nil
		},
		jwtlib.WithIssuer(gate1Issuer),
		jwtlib.WithAudience(gate1Audience),
		jwtlib.WithExpirationRequired(),
	)
	return err
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="shield-gate2"`)
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}
