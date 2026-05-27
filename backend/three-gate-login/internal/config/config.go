package config

import (
	"os"
	"time"
)

type Config struct {
	Addr           string
	JWTSecret      []byte
	JWTTTL         time.Duration
	JWTIssuer      string
	JWTAudience    string
	ClientIDHeader string
}

func Load() Config {
	return Config{
		Addr:           getEnv("GATE2_ADDR", ":8080"),
		JWTSecret:      []byte(getEnv("GATE2_JWT_SECRET", "dev-secret-change")),
		JWTTTL:         getDuration("GATE2_JWT_TTL", 10*time.Minute),
		JWTIssuer:      getEnv("GATE2_JWT_ISSUER", "shield-gate2"),
		JWTAudience:    getEnv("GATE2_JWT_AUDIENCE", "shield-gate3"),
		ClientIDHeader: getEnv("GATE2_CLIENT_ID_HEADER", "x-client-dn"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

