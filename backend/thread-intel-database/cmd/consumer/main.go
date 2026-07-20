package main

import (
	"log"
	"time"

	"shield/thread-intel/internal/config"
	"shield/thread-intel/internal/kafka"
	"shield/thread-intel/internal/store"
)

func main() {
	log.Println("Starting Threat-Intel Kafka Consumer...")

	cfg := config.Load()

	// Wait for DB to be ready
	var ts *store.ThreatStore
	var err error
	for i := 0; i < 10; i++ {
		ts, err = store.NewThreatStore(cfg.DatabaseDSN)
		if err == nil {
			break
		}
		log.Printf("Waiting for database... %v", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to database after retries: %v", err)
	}
	defer ts.Close()

	// Start Kafka Consumer
	consumer, err := kafka.NewConsumer(cfg, ts)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Start(); err != nil {
		log.Fatalf("Consumer error: %v", err)
	}
}
