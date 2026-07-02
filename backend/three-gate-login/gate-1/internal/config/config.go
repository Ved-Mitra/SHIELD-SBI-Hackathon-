package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"time"
)

// Config holds all runtime configuration for the Gate 1 service.
// All required fields cause a fatal log if missing at startup (fail-fast).
type Config struct {
	// Addr is the TCP address the HTTP server binds to.
	Addr string

	// JWTPrivateKey is the RSA private key used to sign G1-JWTs (RS256).
	JWTPrivateKey *rsa.PrivateKey
	// JWTTTL is the lifetime of issued G1-JWTs.
	JWTTTL time.Duration
	// JWTIssuer is the "iss" claim value. Default: "shield-gate1".
	JWTIssuer string
	// JWTAudience is the "aud" claim value. Default: "shield-gate2".
	JWTAudience string

	// PackageName is the expected Android package name. Default: "com.sbi.yono".
	PackageName string
	// AppID is the expected iOS App ID (<TEAM_ID>.<bundle_id>). Default: "TEAMID.com.sbi.yono".
	AppID string

	// ServiceAccountPath is the path to a Google service account JSON for the Play Integrity API.
	// When empty, Application Default Credentials (ADC) are used.
	ServiceAccountPath string

	// NonceStoreDSN is the Redis address used for nonce replay prevention.
	// Default: "localhost:6379". When empty, falls back to in-memory (dev only).
	NonceStoreDSN string
	// NonceTTL is how long a consumed nonce is kept to prevent replay.
	NonceTTL time.Duration

	// MockAttestation skips real Play Integrity / App Attest verification.
	// MUST NOT be set to true in production. Prints a loud warning at startup.
	MockAttestation bool

	// kafka broker url
	KafkaBrokerUrl string
}

// Load reads configuration from environment variables and returns a validated Config.
// Any required variable that is missing causes an immediate fatal log.
func Load() Config {
	keyPath := getEnv("GATE1_JWT_PRIVATE_KEY_PATH", "certs/gate1/private.pem")
	privateKey := loadRSAPrivateKey(keyPath)

	return Config{
		Addr:               getEnv("GATE1_ADDR", ":8081"),
		JWTPrivateKey:      privateKey,
		JWTTTL:             getDuration("GATE1_JWT_TTL", 2*time.Minute),
		JWTIssuer:          getEnv("GATE1_JWT_ISSUER", "shield-gate1"),
		JWTAudience:        getEnv("GATE1_JWT_AUDIENCE", "shield-gate2"),
		PackageName:        getEnv("GATE1_ANDROID_PACKAGE_NAME", "com.sbi.yono"),
		AppID:              getEnv("GATE1_IOS_APP_ID", "TEAMID.com.sbi.yono"),
		ServiceAccountPath: getEnv("GATE1_SERVICE_ACCOUNT_PATH", ""),
		NonceStoreDSN:      getEnv("GATE1_NONCE_STORE_ADDR", "localhost:6379"),
		NonceTTL:           getDuration("GATE1_NONCE_TTL", 5*time.Minute),
		MockAttestation:    getBool("GATE1_MOCK_ATTESTATION"),
		KafkaBrokerUrl: 	getEnv("KAFKA_BROKER_URL","localhost:9092"),
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

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
// Supports both PKCS#1 ("RSA PRIVATE KEY") and PKCS#8 ("PRIVATE KEY") formats.
func loadRSAPrivateKey(path string) *rsa.PrivateKey {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("[FATAL] cannot read private key %s: %v\n"+
			"  → Generate keys with: ./scripts/gen-gate1-keys.sh", path, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		log.Fatalf("[FATAL] no PEM block found in %s", path)
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
		log.Fatalf("[FATAL] unexpected PEM block type %q in %s", block.Type, path)
	}
	return nil // unreachable
}
