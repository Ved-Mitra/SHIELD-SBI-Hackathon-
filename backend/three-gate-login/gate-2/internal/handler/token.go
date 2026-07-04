package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"shield/three-gate-login/internal/config"
	"shield/three-gate-login/internal/jwt"
	"shield/three-gate-login/internal/kafka"
	"shield/three-gate-login/internal/model"
	"shield/three-gate-login/internal/mtls"
)

// TokenHandler issues a G2-JWT after the full Gate 2 trust chain is satisfied:
//
//  1. Rate limiter passed (middleware layer)
//  2. G1-JWT validated (Gate1Auth middleware layer)
//  3. mTLS client identity extracted from Envoy header (this handler)
//  4. G2-JWT signed with Gate 2's RSA private key and returned
type TokenHandler struct {
	Config config.Config
}

func (h TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// ── Extract mTLS client identity from Envoy-injected header ──────────────
	// Envoy sets x-client-dn = %DOWNSTREAM_PEER_SUBJECT% after verifying the
	// client certificate. This is the distinguished name of the mobile app's cert.
	clientID, ok := mtls.ClientIdentity(r, h.Config.ClientIDHeader)
	if !ok {
		log.Printf("[WARN] missing mTLS client identity header %q (req_id=%s)",
			h.Config.ClientIDHeader, r.Header.Get("X-Request-ID"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"mTLS client identity not found"}`))
		go kafka.PublishEvent(kafka.AuthEvent{UserID: clientID, Gate: 2, Status: "FAILED", Reason: "mTLS client identity not found", TimeStamp: time.Now().UnixMilli()})
		return
	}

	// ── Issue G2-JWT (RS256) ──────────────────────────────────────────────────
	now := time.Now()
	token, err := jwt.IssueToken(jwt.IssueInput{
		PrivateKey: h.Config.JWTPrivateKey,
		Issuer:     h.Config.JWTIssuer,
		Audience:   h.Config.JWTAudience,
		Subject:    clientID,
		TTL:        h.Config.JWTTTL,
		Now:        now,
	})
	if err != nil {
		log.Printf("[ERROR] G2-JWT issuance failed: %v (req_id=%s)",
			err, r.Header.Get("X-Request-ID"))
		w.WriteHeader(http.StatusInternalServerError)
		go kafka.PublishEvent(kafka.AuthEvent{UserID: clientID, Gate: 2, Status: "FAILED", Reason: "Internal Server Error issuing G2-JWT", TimeStamp: time.Now().UnixMilli()})
		return
	}

	go kafka.PublishEvent(kafka.AuthEvent{UserID: clientID, Gate: 2, Status: "PASSED", Reason: "Gate-2 mTLS token verified", TimeStamp: time.Now().UnixMilli()})

	log.Printf("[INFO] G2-JWT issued: sub=%q exp=%s req_id=%s",
		clientID, now.Add(h.Config.JWTTTL).Format(time.RFC3339),
		r.Header.Get("X-Request-ID"))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.TokenResponse{
		Token:     token,
		ExpiresIn: int(h.Config.JWTTTL.Seconds()),
		TokenType: "Bearer",
	})
}
