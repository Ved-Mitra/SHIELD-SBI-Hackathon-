package nonce

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store provides an idempotency-safe nonce consumer:
// the first call for a given nonce returns true; all subsequent calls return false.
// This prevents replay attacks on attestation tokens.
type Store interface {
	// Consume marks a nonce as used. Returns true only on first call for this nonce.
	Consume(ctx context.Context, nonce string) bool
}

// ── Redis-backed store ────────────────────────────────────────────────────────

type redisStore struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewRedisStore connects to Redis and returns a Store backed by it.
// Returns an in-memory fallback (with a loud warning) if the connection fails.
func NewRedisStore(addr string, ttl time.Duration) Store {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[WARN] cannot reach Redis at %s: %v — falling back to in-memory nonce store (NOT safe for multi-instance)", addr, err)
		return NewMemStore(ttl)
	}

	log.Printf("[INFO] nonce store: Redis at %s (TTL=%s)", addr, ttl)
	return &redisStore{rdb: rdb, ttl: ttl}
}

// Consume uses Redis SET NX (set if not exists) for atomic, single-use nonce tracking.
func (s *redisStore) Consume(ctx context.Context, nonce string) bool {
	ok, err := s.rdb.SetNX(ctx, "nonce:"+nonce, "1", s.ttl).Result()
	if err != nil {
		log.Printf("[ERROR] Redis nonce store: %v", err)
		return false
	}
	return ok
}

// ── In-memory fallback store ──────────────────────────────────────────────────
// Used for local development / single-node demos when Redis is unavailable.
// NOT suitable for multi-instance deployments.

type entry struct {
	expiry time.Time
}

type memStore struct {
	mu      sync.Mutex
	entries map[string]entry
	ttl     time.Duration
}

// NewMemStore returns an in-memory nonce store (dev/fallback only).
func NewMemStore(ttl time.Duration) Store {
	s := &memStore{
		entries: make(map[string]entry),
		ttl:     ttl,
	}
	go s.gcLoop()
	return s
}

func (s *memStore) Consume(_ context.Context, nonce string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.entries[nonce]; ok && time.Now().Before(e.expiry) {
		return false // already seen
	}
	s.entries[nonce] = entry{expiry: time.Now().Add(s.ttl)}
	return true
}

// gcLoop periodically removes expired entries to bound memory usage.
func (s *memStore) gcLoop() {
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, e := range s.entries {
			if now.After(e.expiry) {
				delete(s.entries, k)
			}
		}
		s.mu.Unlock()
	}
}
