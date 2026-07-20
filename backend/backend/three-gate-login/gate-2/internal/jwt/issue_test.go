package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"
)

// generateTestKey creates a fresh RSA-2048 key pair for unit tests.
// This avoids a file dependency and makes tests self-contained.
func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}

// TestIssueAndVerify checks the full issue → verify round-trip for G2-JWTs.
func TestIssueAndVerify(t *testing.T) {
	privateKey := generateTestKey(t)
	publicKey := &privateKey.PublicKey
	now := time.Now()

	token, err := IssueToken(IssueInput{
		PrivateKey: privateKey,
		Issuer:     "shield-gate2",
		Audience:   "shield-gate3",
		Subject:    "CN=shield-mobile,O=SHIELD,C=IN",
		TTL:        2 * time.Minute,
		Now:        now,
	})
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	parsed, err := VerifyToken(token, publicKey, "shield-gate2", "shield-gate3")
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}

	if !parsed.Valid {
		t.Fatal("expected token to be valid")
	}
}

// TestVerifyRejectsWrongIssuer ensures VerifyToken rejects a token with the wrong iss.
func TestVerifyRejectsWrongIssuer(t *testing.T) {
	key := generateTestKey(t)
	token, _ := IssueToken(IssueInput{
		PrivateKey: key,
		Issuer:     "wrong-issuer",
		Audience:   "shield-gate3",
		Subject:    "test",
		TTL:        2 * time.Minute,
		Now:        time.Now(),
	})

	_, err := VerifyToken(token, &key.PublicKey, "shield-gate2", "shield-gate3")
	if err == nil {
		t.Fatal("expected error for wrong issuer, got nil")
	}
}

// TestVerifyRejectsWrongAudience ensures VerifyToken rejects a token with the wrong aud.
func TestVerifyRejectsWrongAudience(t *testing.T) {
	key := generateTestKey(t)
	token, _ := IssueToken(IssueInput{
		PrivateKey: key,
		Issuer:     "shield-gate2",
		Audience:   "wrong-audience",
		Subject:    "test",
		TTL:        2 * time.Minute,
		Now:        time.Now(),
	})

	_, err := VerifyToken(token, &key.PublicKey, "shield-gate2", "shield-gate3")
	if err == nil {
		t.Fatal("expected error for wrong audience, got nil")
	}
}

// TestVerifyRejectsExpiredToken ensures VerifyToken rejects an expired token.
func TestVerifyRejectsExpiredToken(t *testing.T) {
	key := generateTestKey(t)
	token, _ := IssueToken(IssueInput{
		PrivateKey: key,
		Issuer:     "shield-gate2",
		Audience:   "shield-gate3",
		Subject:    "test",
		TTL:        -1 * time.Second, // already expired
		Now:        time.Now(),
	})

	_, err := VerifyToken(token, &key.PublicKey, "shield-gate2", "shield-gate3")
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

// TestVerifyRejectsWrongKey ensures VerifyToken rejects a token signed by a different key.
func TestVerifyRejectsWrongKey(t *testing.T) {
	signingKey := generateTestKey(t)
	verifyKey := generateTestKey(t) // different key

	token, _ := IssueToken(IssueInput{
		PrivateKey: signingKey,
		Issuer:     "shield-gate2",
		Audience:   "shield-gate3",
		Subject:    "test",
		TTL:        2 * time.Minute,
		Now:        time.Now(),
	})

	_, err := VerifyToken(token, &verifyKey.PublicKey, "shield-gate2", "shield-gate3")
	if err == nil {
		t.Fatal("expected error for wrong signing key, got nil")
	}
}
