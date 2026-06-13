// Package server wires together all Gate 1 routes and middleware into a single
// http.Handler ready to be passed to http.Server.
package server

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"shield/gate1/internal/config"
	"shield/gate1/internal/handler"
	"shield/gate1/internal/middleware"
	"shield/gate1/internal/nonce"
)

// New builds and returns the fully configured Gate 1 HTTP handler.
//
// Middleware stack (outermost → innermost):
//   Logging → RequestID → RateLimiter → Routes
//
// Routes:
//   GET  /healthz        — liveness probe (no rate limit)
//   POST /gate1/attest   — platform attestation endpoint
func New(cfg config.Config) (http.Handler, error) {
	// ── Nonce store ───────────────────────────────────────────────────────────
	ns := nonce.NewRedisStore(cfg.NonceStoreDSN, cfg.NonceTTL)

	// ── Attestation handler ───────────────────────────────────────────────────
	attestH, err := handler.New(cfg, ns)
	if err != nil {
		return nil, fmt.Errorf("initialising attest handler: %w", err)
	}

	// Rate limiter: 10 req/min per IP (burst of 20 for demo) ───────────────
	// rate.Every(6*time.Second) ≈ 10 requests per minute sustained.
	// Burst of 20 allows rapid test sequences without self-rate-limiting.
	rl := middleware.NewRateLimiter(rate.Every(6*time.Second), 20)

	// ── Mux ───────────────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// Health check — no rate limiting needed
	mux.HandleFunc("/healthz", handler.Health)

	// Attestation endpoint — apply rate limiter
	mux.Handle("/gate1/attest", rl.Middleware(attestH))

	// ── Middleware chain ──────────────────────────────────────────────────────
	var h http.Handler = mux
	h = middleware.RequestID(h)
	h = middleware.Logging(h)

	return h, nil
}

// DefaultServer returns a production-hardened *http.Server wrapping the handler.
// Call this instead of bare http.ListenAndServe to get proper timeouts.
func DefaultServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
