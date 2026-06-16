package main

import (
	"log"

	"shield/three-gate-login/internal/config"
	"shield/three-gate-login/internal/server"
)

func main() {
	cfg := config.Load()

	if cfg.MockGate1 {
		log.Println("╔══════════════════════════════════════════════════════════╗")
		log.Println("║  ⚠️  WARNING: GATE2_MOCK_GATE1=true                      ║")
		log.Println("║  G1-JWT verification is DISABLED.                        ║")
		log.Println("║  Any client can call /gate2/token without a G1-JWT.      ║")
		log.Println("║  NEVER deploy this configuration to production.          ║")
		log.Println("╚══════════════════════════════════════════════════════════╝")
	}

	h := server.New(cfg)
	srv := server.DefaultServer(cfg.Addr, h)

	log.Printf("[INFO] gate-2 listening on %s (mock_gate1=%v)", cfg.Addr, cfg.MockGate1)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("[FATAL] server stopped: %v", err)
	}
}
