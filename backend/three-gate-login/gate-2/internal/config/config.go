package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"time"
)

// Config holds all runtime configuration for the Gate 2 service.
type Config struct {
	// Addr is the TCP address the HTTP server binds to (Envoy talks to this).
	Addr string

	// Gate 2 signs G2-JWTs with its own private key (RS256).
	JWTPrivateKey *rsa.PrivateKey
	JWTTTL        time.Duration
	JWTIssuer     string
	JWTAudience   string

	// Gate 2 verifies incoming G1-JWTs from Gate 1 with Gate 1's public key.
	Gate1PublicKey *rsa.PublicKey

	// ClientIDHeader is the header Envoy sets after successful mTLS, containing
	// the client certificate's Distinguished Name (DN).
	ClientIDHeader string

	// MockGate1 skips G1-JWT verification (for local dev without Gate 1 running).
	MockGate1 bool

	// Kafka Broker URL
	KafkaBrokerUrl string
}

// Load reads configuration from environment variables and returns a validated Config.
// Missing key files cause an immediate fatal log (fail-fast).
func Load() Config {
	gate2KeyPath := getEnv("GATE2_JWT_PRIVATE_KEY_PATH", "certs/gate2/private.pem")
	gate1PubPath := getEnv("GATE2_GATE1_PUBLIC_KEY_PATH", "certs/gate1/public.pem")

	gate2PrivKey := loadRSAPrivateKey(gate2KeyPath)

	var gate1PubKey *rsa.PublicKey
	mockGate1 := getBool("GATE2_MOCK_GATE1")
	if !mockGate1 {
		gate1PubKey = loadRSAPublicKey(gate1PubPath)
	}

	return Config{
		Addr:           getEnv("GATE2_ADDR", ":8080"),
		JWTPrivateKey:  gate2PrivKey,
		JWTTTL:         getDuration("GATE2_JWT_TTL", 10*time.Minute),
		JWTIssuer:      getEnv("GATE2_JWT_ISSUER", "shield-gate2"),
		JWTAudience:    getEnv("GATE2_JWT_AUDIENCE", "shield-gate3"),
		Gate1PublicKey: gate1PubKey,
		ClientIDHeader: getEnv("GATE2_CLIENT_ID_HEADER", "x-client-dn"),
		MockGate1:      mockGate1,
		KafkaBrokerUrl: getEnv("KAFKA_BROKER_URL", "localhost:9092"),
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBool(key string) bool {
	v := os.Getenv(key)
	return v == "true" || v == "1"
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		log.Printf("[WARN] invalid duration for %s=%q, using default %s", key, v, fallback)
	}
	return fallback
}

// loadRSAPrivateKey parses an RSA private key from a PEM file.
// Supports PKCS#1 and PKCS#8 formats. Fatals if the file cannot be read or parsed.
func loadRSAPrivateKey(path string) *rsa.PrivateKey {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("[FATAL] cannot read Gate 2 private key %s: %v\n"+
			"  → Generate keys with: ./scripts/gen-gate2-keys.sh", path, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		log.Fatalf("[FATAL] no PEM block in %s", path)
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Fatalf("[FATAL] parsing PKCS1 key from %s: %v", path, err)
		}
		return key
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			log.Fatalf("[FATAL] parsing PKCS8 key from %s: %v", path, err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			log.Fatalf("[FATAL] key in %s is not RSA (got %T)", path, key)
		}
		return rsaKey
	default:
		log.Fatalf("[FATAL] unexpected PEM type %q in %s", block.Type, path)
	}
	return nil
}

// loadRSAPublicKey parses an RSA public key from a PEM file (PKIX format).
// This is Gate 1's public key used to verify incoming G1-JWTs.
func loadRSAPublicKey(path string) *rsa.PublicKey {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("[FATAL] cannot read Gate 1 public key %s: %v\n"+
			"  → Copy from gate-1: cp ../gate-1/certs/gate1/public.pem %s", path, err, path)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		log.Fatalf("[FATAL] no PEM block in %s", path)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatalf("[FATAL] parsing public key from %s: %v", path, err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		log.Fatalf("[FATAL] public key in %s is not RSA (got %T)", path, pub)
	}
	return rsaPub
}
