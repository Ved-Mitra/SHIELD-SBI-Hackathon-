package middleware

import (
	"net"
	"net/http"
	"shield/three-gate-login/internal/kafka"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides per-IP rate limiting using a token bucket algorithm.
// Each IP gets its own bucket; stale buckets are periodically evicted.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
	limit   rate.Limit
	burst   int
	idleTTL time.Duration
}

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a RateLimiter.
//
//   - r: sustained requests per second per IP (e.g. rate.Every(6s) ≈ 10/min)
//   - burst: maximum burst size
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*ipEntry),
		limit:   r,
		burst:   burst,
		idleTTL: 10 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

// Middleware returns an http.Handler that rejects requests exceeding the rate
// limit with HTTP 429 Too Many Requests.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		if !rl.allow(ip) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			go kafka.PublishEvent(kafka.AuthEvent{UserID: "unknown", Gate: 2, Status: "FAILED", Reason: "Rate limit exceeded (Possible Brute Force)", TimeStamp: time.Now().UnixMilli()})

			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	e, ok := rl.entries[ip]
	if !ok {
		e = &ipEntry{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.entries[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter.Allow()
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.idleTTL)
		for ip, e := range rl.entries {
			if e.lastSeen.Before(cutoff) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// realIP extracts the client IP, honouring X-Forwarded-For set by Envoy.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return strings.TrimSpace(xff[:i])
			}
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
