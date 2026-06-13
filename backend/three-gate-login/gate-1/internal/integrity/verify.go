// Package integrity implements server-side verification of Google Play Integrity tokens.
//
// Flow:
//  1. The Android client calls PlayIntegrityManager.requestIntegrityToken(nonce) using
//     our server-issued nonce as the request hash.
//  2. The client forwards the opaque integrity token to POST /gate1/attest.
//  3. This package sends the token to the Play Integrity API, which decodes and returns
//     a structured verdict containing app, device, and account details.
//  4. We validate the verdict fields and return a subject string on success.
//
// Reference: https://developer.android.com/google/play/integrity/overview
package integrity

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	playintegrity "google.golang.org/api/playintegrity/v1"
)

// Config holds the parameters the Verifier needs to call the Play Integrity API.
type Config struct {
	// PackageName is the expected Android application package name (e.g. "com.sbi.yono").
	PackageName string
	// ServiceAccountPath is the path to a Google service-account JSON key file.
	// Leave empty to use Application Default Credentials (ADC).
	ServiceAccountPath string
}

// Verifier calls the Play Integrity API to validate an integrity token.
type Verifier struct {
	cfg Config
}

// New creates a new Verifier.
func New(cfg Config) *Verifier {
	return &Verifier{cfg: cfg}
}

// Verify decodes and validates an Android Play Integrity token.
// Returns a subject string ("android:<packageName>") on success.
//
// Validation steps:
//  1. requestDetails.nonce must match the nonce we issued.
//  2. requestDetails.requestPackageName must match our expected package.
//  3. appIntegrity.appRecognitionVerdict must be "PLAY_RECOGNIZED".
//  4. deviceIntegrity must contain "MEETS_DEVICE_INTEGRITY".
func (v *Verifier) Verify(ctx context.Context, integrityToken, nonce string) (string, error) {
	svc, err := v.newService(ctx)
	if err != nil {
		return "", fmt.Errorf("play integrity client: %w", err)
	}

	resp, err := svc.V1.DecodeIntegrityToken(
		v.cfg.PackageName,
		&playintegrity.DecodeIntegrityTokenRequest{IntegrityToken: integrityToken},
	).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("decodeIntegrityToken RPC: %w", err)
	}

	payload := resp.TokenPayloadExternal
	if payload == nil {
		return "", fmt.Errorf("empty token payload")
	}

	if err := v.validatePayload(payload, nonce); err != nil {
		return "", err
	}

	return "android:" + v.cfg.PackageName, nil
}

// validatePayload checks all required verdict fields.
func (v *Verifier) validatePayload(p *playintegrity.TokenPayloadExternal, expectedNonce string) error {
	// ── 1. Request details ────────────────────────────────────────────────────
	if p.RequestDetails == nil {
		return fmt.Errorf("verdict missing requestDetails")
	}
	if p.RequestDetails.Nonce != expectedNonce {
		return fmt.Errorf("nonce mismatch: got %q, want %q",
			p.RequestDetails.Nonce, expectedNonce)
	}
	if p.RequestDetails.RequestPackageName != v.cfg.PackageName {
		return fmt.Errorf("package mismatch: got %q, want %q",
			p.RequestDetails.RequestPackageName, v.cfg.PackageName)
	}

	// ── 2. App integrity ──────────────────────────────────────────────────────
	if p.AppIntegrity == nil {
		return fmt.Errorf("verdict missing appIntegrity")
	}
	switch p.AppIntegrity.AppRecognitionVerdict {
	case "PLAY_RECOGNIZED":
		// ✅ pass
	case "UNRECOGNIZED_VERSION":
		return fmt.Errorf("app verdict UNRECOGNIZED_VERSION — sideloaded or unofficial build")
	default:
		return fmt.Errorf("app verdict not recognized: %q", p.AppIntegrity.AppRecognitionVerdict)
	}

	// ── 3. Device integrity ───────────────────────────────────────────────────
	if p.DeviceIntegrity == nil || len(p.DeviceIntegrity.DeviceRecognitionVerdict) == 0 {
		return fmt.Errorf("verdict missing deviceIntegrity")
	}
	if !containsVerdict(p.DeviceIntegrity.DeviceRecognitionVerdict, "MEETS_DEVICE_INTEGRITY") {
		return fmt.Errorf("device integrity not met: verdicts=%v",
			p.DeviceIntegrity.DeviceRecognitionVerdict)
	}

	log.Printf("[INFO] Play Integrity pass: pkg=%s device_verdicts=%v",
		v.cfg.PackageName, p.DeviceIntegrity.DeviceRecognitionVerdict)
	return nil
}

// newService creates a Play Integrity API service, optionally with a service-account key.
func (v *Verifier) newService(ctx context.Context) (*playintegrity.Service, error) {
	var opts []option.ClientOption

	if v.cfg.ServiceAccountPath != "" {
		data, err := os.ReadFile(v.cfg.ServiceAccountPath)
		if err != nil {
			return nil, fmt.Errorf("reading service account %s: %w", v.cfg.ServiceAccountPath, err)
		}
		opts = append(opts, option.WithCredentialsJSON(data))
	}
	// If no service account path, the Google client uses ADC automatically.

	return playintegrity.NewService(ctx, opts...)
}

func containsVerdict(verdicts []string, target string) bool {
	for _, v := range verdicts {
		if v == target {
			return true
		}
	}
	return false
}
