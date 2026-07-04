package main

import (
	"auth-intel/internal/config"
	"auth-intel/internal/kafka"
	"auth-intel/internal/store"
	"log"
	"time"
)

func main() {
	log.Printf("Starting Auth-intel kafka Consumer");

	cfg:=config.Load()

	var as *store.AuthStore
	var err error
	for i:=0; i<10; i++{
		as, err = store.NewAuthStore(cfg.DatabaseDSN)
		if err==nil{
			break
		}
		log.Printf("Waiting for database: %v", err)
		time.Sleep(2*time.Second)
	}
	if err!=nil{
		log.Fatalf("[FATAL] Could not connect to database after retries in Auth-intel Database: %v",err)
	}

	defer as.Close()

	consumer, err :=kafka.NewConsumer(cfg,as)
	if err!=nil{
		log.Fatalf("[FATAL] Failed to initialize Kafka consumer in Auth-intel database")
	}
	defer consumer.Close()

	if err:=consumer.Start(); err!=nil{
		log.Fatalf("[FATAL] Consumer error: %v",err)
	}
	
}