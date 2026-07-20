// Package appattest implements server-side verification of Apple App Attest attestation objects.
//
// Verification steps (per Apple's documentation):
//  1. Decode the base64url attestation object (CBOR format).
//  2. Verify fmt == "apple-appattest".
//  3. Build and verify the x5c certificate chain against the Apple App Attest Root CA.
//  4. Compute expectedNonce = SHA-256(authData || clientDataHash) and verify it matches
//     the nonce embedded in the credential certificate's OID 1.2.840.113635.100.8.2 extension.
//  5. Verify that authData[0:32] (rpIdHash) matches SHA-256(AppID).
//  6. Verify the sign counter in authData is 0 (new key, never used before).
//  7. Verify the keyID matches SHA-256(credCert.PublicKey DER).
//
// Reference: https://developer.apple.com/documentation/devicecheck/validating_apps_that_connect_to_your_server
package appattest

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/fxamacker/cbor/v2"
)

// ── CBOR structures ───────────────────────────────────────────────────────────

type attestationObject struct {
	Fmt      string          `cbor:"fmt"`
	AttStmt  attestStatement `cbor:"attStmt"`
	AuthData []byte          `cbor:"authData"`
}

type attestStatement struct {
	X5C     [][]byte `cbor:"x5c"`
	Receipt []byte   `cbor:"receipt"`
}

// ── Verifier ──────────────────────────────────────────────────────────────────

// Config holds parameters for the App Attest verifier.
type Config struct {
	// AppID is the fully-qualified iOS App ID: "<TEAM_ID>.<bundle_identifier>".
	// Example: "ABCD1234EF.com.sbi.yono"
	AppID string
}

// Verifier validates Apple App Attest attestation objects.
type Verifier struct {
	cfg     Config
	rootCAs *x509.CertPool
}

// New creates a Verifier and loads the Apple App Attest Root CA into a cert pool.
func New(cfg Config) (*Verifier, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM([]byte(appleAppAttestRootCA)) {
		return nil, fmt.Errorf("failed to load Apple App Attest Root CA")
	}
	return &Verifier{cfg: cfg, rootCAs: pool}, nil
}

// Verify validates an App Attest attestation object.
//
// Parameters:
//   - attestObjectB64: base64 (or base64url) CBOR attestation object from the iOS SDK.
//   - clientDataHashB64: base64(SHA-256(nonce)) — the client-side hash of our challenge nonce.
//   - keyID: base64url key identifier from DCAppAttestService.generateKey().
//
// Returns a subject string ("ios:<AppID>") on success.
func (v *Verifier) Verify(attestObjectB64, clientDataHashB64, keyID string) (string, error) {
	// ── 1. Decode the attestation object ─────────────────────────────────────
	attestBytes, err := flexDecode(attestObjectB64)
	if err != nil {
		return "", fmt.Errorf("decoding attest object: %w", err)
	}

	var obj attestationObject
	if err := cbor.Unmarshal(attestBytes, &obj); err != nil {
		return "", fmt.Errorf("parsing attestation CBOR: %w", err)
	}

	// ── 2. Validate format ────────────────────────────────────────────────────
	if obj.Fmt != "apple-appattest" {
		return "", fmt.Errorf("unexpected attestation format: %q (want apple-appattest)", obj.Fmt)
	}
	if len(obj.AttStmt.X5C) < 2 {
		return "", fmt.Errorf("x5c certificate chain too short (got %d, need ≥2)", len(obj.AttStmt.X5C))
	}

	// ── 3. Parse and verify certificate chain ────────────────────────────────
	credCert, err := x509.ParseCertificate(obj.AttStmt.X5C[0])
	if err != nil {
		return "", fmt.Errorf("parsing credential cert: %w", err)
	}
	interCert, err := x509.ParseCertificate(obj.AttStmt.X5C[1])
	if err != nil {
		return "", fmt.Errorf("parsing intermediate cert: %w", err)
	}

	intermediates := x509.NewCertPool()
	intermediates.AddCert(interCert)

	if _, err := credCert.Verify(x509.VerifyOptions{
		Roots:         v.rootCAs,
		Intermediates: intermediates,
	}); err != nil {
		return "", fmt.Errorf("certificate chain verification failed: %w", err)
	}

	// ── 4. Verify nonce in credCert extension ─────────────────────────────────
	clientDataHash, err := flexDecode(clientDataHashB64)
	if err != nil {
		return "", fmt.Errorf("decoding clientDataHash: %w", err)
	}

	// expectedNonce = SHA-256(authData || clientDataHash)
	composite := append(obj.AuthData, clientDataHash...) //nolint:gocritic
	expectedNonce := sha256.Sum256(composite)

	if err := verifyNonceExtension(credCert, expectedNonce[:]); err != nil {
		return "", fmt.Errorf("nonce extension: %w", err)
	}

	// ── 5. Verify rpIdHash ────────────────────────────────────────────────────
	if len(obj.AuthData) < 37 {
		return "", fmt.Errorf("authData too short (%d bytes)", len(obj.AuthData))
	}
	rpIdHash := sha256.Sum256([]byte(v.cfg.AppID))
	if !bytes.Equal(obj.AuthData[:32], rpIdHash[:]) {
		return "", fmt.Errorf("rpIdHash mismatch — wrong App ID or tampered authData")
	}

	// ── 6. Verify counter == 0 ────────────────────────────────────────────────
	counter := binary.BigEndian.Uint32(obj.AuthData[33:37])
	if counter != 0 {
		return "", fmt.Errorf("expected counter=0 for new attestation, got %d", counter)
	}

	// ── 7. Verify keyID matches public key hash ───────────────────────────────
	if err := verifyKeyID(credCert, keyID); err != nil {
		return "", fmt.Errorf("keyID mismatch: %w", err)
	}

	log.Printf("[INFO] App Attest pass: appID=%s keyID=%s", v.cfg.AppID, keyID)
	return "ios:" + v.cfg.AppID, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

// nonceExtensionOID is Apple's OID for the nonce embedded in the credential certificate.
// OID: 1.2.840.113635.100.8.2
var nonceExtensionOID = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 2}

// verifyNonceExtension extracts and compares the nonce from the credential certificate.
func verifyNonceExtension(cert *x509.Certificate, expectedNonce []byte) error {
	for _, ext := range cert.Extensions {
		if !ext.Id.Equal(nonceExtensionOID) {
			continue
		}
		// The extension value is a DER SEQUENCE containing an OCTET STRING of the nonce.
		var outer asn1.RawValue
		if _, err := asn1.Unmarshal(ext.Value, &outer); err != nil {
			return fmt.Errorf("parsing outer ASN.1: %w", err)
		}
		var nonce []byte
		if _, err := asn1.Unmarshal(outer.Bytes, &nonce); err != nil {
			return fmt.Errorf("parsing inner nonce: %w", err)
		}
		if !bytes.Equal(nonce, expectedNonce) {
			return fmt.Errorf("nonce value mismatch")
		}
		return nil
	}
	return fmt.Errorf("nonce extension (OID 1.2.840.113635.100.8.2) not found in credCert")
}

// verifyKeyID checks that the keyID matches SHA-256 of the credential public key DER bytes.
func verifyKeyID(cert *x509.Certificate, keyIDB64 string) error {
	ecKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("credential cert public key is not ECDSA")
	}
	pubKeyDER, err := x509.MarshalPKIXPublicKey(ecKey)
	if err != nil {
		return fmt.Errorf("marshalling public key: %w", err)
	}
	expectedKeyID := sha256.Sum256(pubKeyDER)
	expectedB64 := base64.RawURLEncoding.EncodeToString(expectedKeyID[:])

	if keyIDB64 != expectedB64 {
		return fmt.Errorf("keyID mismatch: got %q, want %q", keyIDB64, expectedB64)
	}
	return nil
}

// flexDecode decodes a string from standard base64, base64url, or their raw (no-padding) forms.
func flexDecode(s string) ([]byte, error) {
	// Try each encoding in order of likelihood
	encodings := []struct {
		name string
		enc  *base64.Encoding
	}{
		{"std", base64.StdEncoding},
		{"rawurl", base64.RawURLEncoding},
		{"url", base64.URLEncoding},
		{"rawstd", base64.RawStdEncoding},
	}
	for _, e := range encodings {
		if b, err := e.enc.DecodeString(s); err == nil {
			return b, nil
		}
	}
	// Last resort: try reading as PEM
	if block, _ := pem.Decode([]byte(s)); block != nil {
		return block.Bytes, nil
	}
	return nil, fmt.Errorf("cannot decode base64 string (tried std, url, raw variants)")
}
