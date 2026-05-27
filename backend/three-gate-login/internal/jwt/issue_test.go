package jwt

import (
	"testing"
	"time"
)

func TestIssueAndVerify(t *testing.T) {
	secret := []byte("unit-test-secret")
	now := time.Now()

	token, err := IssueToken(IssueInput{
		Secret:   secret,
		Issuer:   "shield-gate2",
		Audience: "shield-gate3",
		Subject:  "client-123",
		TTL:      2 * time.Minute,
		Now:      now,
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	parsed, err := VerifyToken(token, secret, "shield-gate2", "shield-gate3")
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}

	if !parsed.Valid {
		t.Fatalf("token should be valid")
	}
}

