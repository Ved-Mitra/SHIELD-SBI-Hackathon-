package config

import "os"

type Config struct{
	KafkaBrokerUrl string
	DatabaseDSN    string
	KafkaTopic     string
	KafkaGroupID   string
}

func Load() Config {
	return Config{
		KafkaBrokerUrl: getEnv("KAFKA_BROKER_URL", "localhost:9092"),
		DatabaseDSN:    getEnv("DB_DSN", "postgres://shield:shield-pass@localhost:5432/intel_db?sslmode=disable"),
		KafkaTopic:     getEnv("KAFKA_TOPIC", "auth-events"),
		KafkaGroupID:   getEnv("KAFKA_GROUP_ID", "auth-intel-db-group"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}