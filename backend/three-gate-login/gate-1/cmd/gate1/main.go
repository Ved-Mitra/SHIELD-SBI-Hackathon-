package main

import (
	"log"

	"shield/gate1/internal/config"
	"shield/gate1/internal/kafka"
	"shield/gate1/internal/server"
)

func main() {
	cfg := config.Load()

	if cfg.MockAttestation {
		log.Println("╔══════════════════════════════════════════════════════════╗")
		log.Println("║  ⚠️  WARNING: MOCK_ATTESTATION=true                      ║")
		log.Println("║  Play Integrity / App Attest checks are DISABLED.        ║")
		log.Println("║  Any client can obtain a G1-JWT without real attestation.║")
		log.Println("║  NEVER deploy this configuration to production.          ║")
		log.Println("╚══════════════════════════════════════════════════════════╝")
	}

	h, err := server.New(cfg)
	if err != nil {
		log.Fatalf("[FATAL] server init: %v", err)
	}

	defer kafka.CloseProducer()

	srv := server.DefaultServer(cfg.Addr, h)

	log.Printf("[INFO] gate-1 listening on %s (mock=%v)", cfg.Addr, cfg.MockAttestation)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("[FATAL] server stopped: %v", err)
	}
}
