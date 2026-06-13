// Package handler implements the POST /gate1/attest HTTP endpoint.
//
// The handler orchestrates the full Gate 1 verification pipeline:
//  1. Decode and validate the request body.
//  2. Consume the nonce (replay prevention).
//  3. Dispatch to the platform-specific verifier (Android or iOS).
//  4. Issue a G1-JWT (RS256, aud="shield-gate2") on success.
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"shield/gate1/internal/appattest"
	"shield/gate1/internal/config"
	"shield/gate1/internal/integrity"
	"shield/gate1/internal/jwt"
	"shield/gate1/internal/model"
	"shield/gate1/internal/nonce"
)

// AttestHandler wires together all Gate 1 dependencies needed to handle an
// attestation request.
type AttestHandler struct {
	cfg         config.Config
	nonceStore  nonce.Store
	androidVer  *integrity.Verifier
	iosVer      *appattest.Verifier
}

// New creates an AttestHandler with fully initialised verifiers.
func New(cfg config.Config, ns nonce.Store) (*AttestHandler, error) {
	// Android verifier
	androidVer := integrity.New(integrity.Config{
		PackageName:        cfg.PackageName,
		ServiceAccountPath: cfg.ServiceAccountPath,
	})

	// iOS verifier
	iosVer, err := appattest.New(appattest.Config{AppID: cfg.AppID})
	if err != nil {
		return nil, err
	}

	return &AttestHandler{
		cfg:        cfg,
		nonceStore: ns,
		androidVer: androidVer,
		iosVer:     iosVer,
	}, nil
}

// ServeHTTP handles POST /gate1/attest.
func (h *AttestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ── Only POST is accepted ─────────────────────────────────────────────────
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// ── Decode request body ───────────────────────────────────────────────────
	var req model.AttestRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// ── Validate required fields ──────────────────────────────────────────────
	if req.Nonce == "" {
		writeError(w, http.StatusBadRequest, "nonce is required")
		return
	}
	if req.Platform != "android" && req.Platform != "ios" {
		writeError(w, http.StatusBadRequest, `platform must be "android" or "ios"`)
		return
	}

	// ── Consume nonce (replay prevention) ────────────────────────────────────
	// SetNX on Redis / in-memory: returns true only on first call.
	if !h.nonceStore.Consume(r.Context(), req.Nonce) {
		log.Printf("[WARN] replayed or invalid nonce %q from %s", req.Nonce, r.RemoteAddr)
		writeError(w, http.StatusUnauthorized, "invalid or replayed nonce")
		return
	}

	// ── Mock mode (demo / local dev only) ────────────────────────────────────
	var subject string
	var verifyErr error

	if h.cfg.MockAttestation {
		subject = req.Platform + ":mock"
		log.Printf("[WARN] MOCK attestation accepted for platform=%s nonce=%s", req.Platform, req.Nonce)
	} else {
		// ── Real verification ─────────────────────────────────────────────────
		switch req.Platform {
		case "android":
			subject, verifyErr = h.verifyAndroid(r, req)
		case "ios":
			subject, verifyErr = h.verifyIOS(r, req)
		}

		if verifyErr != nil {
			log.Printf("[ERROR] attestation failed platform=%s nonce=%s: %v",
				req.Platform, req.Nonce, verifyErr)
			writeError(w, http.StatusUnauthorized, "attestation verification failed")
			return
		}
	}

	// ── Issue G1-JWT ──────────────────────────────────────────────────────────
	now := time.Now()
	token, err := jwt.IssueToken(jwt.IssueInput{
		PrivateKey: h.cfg.JWTPrivateKey,
		Issuer:     h.cfg.JWTIssuer,
		Audience:   h.cfg.JWTAudience,
		Subject:    subject,
		TTL:        h.cfg.JWTTTL,
		Now:        now,
	})
	if err != nil {
		log.Printf("[ERROR] JWT issuance failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Printf("[INFO] G1-JWT issued: sub=%s exp=%s", subject, now.Add(h.cfg.JWTTTL).Format(time.RFC3339))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.AttestResponse{
		Token:     token,
		ExpiresIn: int(h.cfg.JWTTTL.Seconds()),
		TokenType: "Bearer",
	})
}

// ── platform dispatch ─────────────────────────────────────────────────────────

func (h *AttestHandler) verifyAndroid(r *http.Request, req model.AttestRequest) (string, error) {
	if strings.TrimSpace(req.IntegrityToken) == "" {
		return "", writeClientError("integrity_token is required for platform=android")
	}
	return h.androidVer.Verify(r.Context(), req.IntegrityToken, req.Nonce)
}

func (h *AttestHandler) verifyIOS(r *http.Request, req model.AttestRequest) (string, error) {
	switch {
	case strings.TrimSpace(req.AttestObject) == "":
		return "", writeClientError("attest_object is required for platform=ios")
	case strings.TrimSpace(req.ClientDataHash) == "":
		return "", writeClientError("client_data_hash is required for platform=ios")
	case strings.TrimSpace(req.KeyID) == "":
		return "", writeClientError("key_id is required for platform=ios")
	}
	return h.iosVer.Verify(req.AttestObject, req.ClientDataHash, req.KeyID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// writeClientError creates a plain error value from a validation message.
// The message is logged internally but a generic error is returned to the client.
type clientError struct{ msg string }

func (e clientError) Error() string { return e.msg }

func writeClientError(msg string) error { return clientError{msg: msg} }

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: msg})
}
