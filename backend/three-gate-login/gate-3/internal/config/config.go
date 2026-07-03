package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type Config struct {
	Addr           string
	Gate2PublicKey *rsa.PublicKey
	RedisAddr      string
	DatabaseDSN    string
	MockGate2      bool
	MockFido2      bool

	WebAuthn *webauthn.Config
	KafkaBrokerUrl string
}

func Load() Config {
	mockGate2 := getBool("GATE3_MOCK_GATE2")
	gate2PubPath := getEnv("GATE3_GATE2_PUBLIC_KEY_PATH", "certs/gate2/public.pem")

	var gate2PubKey *rsa.PublicKey
	if !mockGate2 {
		gate2PubKey = loadRSAPublicKey(gate2PubPath)
	}

	originsRaw := getEnv("GATE3_RP_ORIGIN", "http://localhost:8080,http://localhost:8082")
	origins := strings.Split(originsRaw, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	wconfig := &webauthn.Config{
		RPDisplayName: "SBI YONO",
		RPID:          getEnv("GATE3_RP_ID", "localhost"),
		RPOrigins:     origins,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			UserVerification:        protocol.VerificationRequired,
			ResidentKey:             protocol.ResidentKeyRequirementRequired,
		},
		Debug: true,
	}

	return Config{
		Addr:           getEnv("GATE3_ADDR", ":8082"),
		Gate2PublicKey: gate2PubKey,
		RedisAddr:      getEnv("GATE3_REDIS_ADDR", "localhost:6379"),
		DatabaseDSN:    getEnv("GATE3_DB_DSN", ""),
		MockGate2:      mockGate2,
		MockFido2:      getBool("GATE3_MOCK_FIDO2"),
		WebAuthn:       wconfig,
		KafkaBrokerUrl: getEnv("KAFKA_BROKER_URL","localhost:9092"),
	}
}

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

func loadRSAPublicKey(path string) *rsa.PublicKey {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("[FATAL] cannot read Gate 2 public key %s: %v\n"+
			"  → Copy from gate-2: cp ../gate-2/certs/gate2/public.pem %s", path, err, path)
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
