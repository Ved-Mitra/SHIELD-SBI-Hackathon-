package main

import (
	"log"
	"net/http"

	"shield/three-gate-login/internal/config"
	"shield/three-gate-login/internal/server"
)

func main() {
	cfg := config.Load()
	h := server.New(cfg)

	log.Printf("gate-2 listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, h); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

