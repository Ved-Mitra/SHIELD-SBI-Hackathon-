// Package server wires together all Gate 2 routes and middleware.
//
// Middleware chain (outermost → innermost):
//
//	Logging → RequestID → Routes
//	                         └─ /gate2/token: RateLimit → Gate1Auth → TokenHandler
//
// Routes:
//
//	GET  /healthz       — liveness probe (no auth, no rate limit)
//	POST /gate2/token   — issue G2-JWT (requires G1-JWT + mTLS client cert)
package server

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"shield/three-gate-login/internal/config"
	"shield/three-gate-login/internal/handler"
	"shield/three-gate-login/internal/middleware"
)

// New builds and returns the fully configured Gate 2 HTTP handler.
func New(cfg config.Config) http.Handler {
	// ── Rate limiter: ~20 req/min per IP, burst 20 ────────────────────────────
	// /gate2/token is called once per login — low rate is fine for real traffic.
	// Burst of 20 prevents smoke tests from self-rate-limiting.
	rl := middleware.NewRateLimiter(rate.Every(3*time.Second), 20)

	// ── Token endpoint: RateLimit → Gate1Auth → TokenHandler ─────────────────
	tokenH := handler.TokenHandler{Config: cfg}
	var tokenChain http.Handler = tokenH
	tokenChain = middleware.Gate1Auth(cfg.Gate1PublicKey, tokenChain)
	tokenChain = rl.Middleware(tokenChain)

	// ── Mux ───────────────────────────────────────────────────────────────────
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.Health)
	mux.Handle("/gate2/token", tokenChain)

	// ── Outer middleware (applies to all routes) ──────────────────────────────
	var h http.Handler = mux
	h = middleware.RequestID(h)
	h = middleware.Logging(h)
	return h
}

// DefaultServer returns a production-hardened *http.Server with proper timeouts.
func DefaultServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
