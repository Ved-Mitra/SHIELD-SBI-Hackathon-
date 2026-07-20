package main

import (
	"log"
	"net/http"

	"shield/risk-url-engine/internal/config"
	"shield/risk-url-engine/internal/kafka"
	"shield/risk-url-engine/internal/server"
)


func main(){
	//load configuration
	cfg := config.Load()

	h, err :=server.New(cfg)
	if err!=nil{
		log.Fatalf("[FATAL] server init: %v", err)
	}

	defer kafka.CloseProducer()

	srv := server.DefaultServer(cfg.Addr,h)

	log.Printf("[INFO] risk-url-engine listening on %s", cfg.Addr)

	if err := srv.ListenAndServe(); err!=nil && err !=http.ErrServerClosed {
		log.Fatalf("[FATAL] server failed: %v", err)
	}
}
