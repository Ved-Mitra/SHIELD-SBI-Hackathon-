package main

import (
	"log"

	"shield/gate3/internal/config"
	"shield/gate3/internal/server"
)

func main() {
	cfg := config.Load()

	if cfg.MockGate2 {
		log.Println("╔══════════════════════════════════════════════════════════╗")
		log.Println("║  ⚠️  WARNING: GATE3_MOCK_GATE2=true                      ║")
		log.Println("║  G2-JWT verification is DISABLED.                        ║")
		log.Println("║  Any client can call /gate3 endpoints without G2-JWT.    ║")
		log.Println("║  NEVER deploy this configuration to production.          ║")
		log.Println("╚══════════════════════════════════════════════════════════╝")
	}

	h, err := server.New(cfg)
	if err != nil {
		log.Fatalf("[FATAL] failed to initialize server: %v", err)
	}

	srv := server.DefaultServer(cfg.Addr, h)

	log.Printf("[INFO] gate-3 listening on %s (mock_gate2=%v)", cfg.Addr, cfg.MockGate2)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("[FATAL] server stopped: %v", err)
	}
}
