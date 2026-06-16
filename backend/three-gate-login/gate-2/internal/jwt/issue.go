package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// IssueInput carries all parameters needed to create a G2-JWT.
type IssueInput struct {
	PrivateKey *rsa.PrivateKey // Gate 2's RSA private key
	Issuer     string          // "shield-gate2"
	Audience   string          // "shield-gate3"
	Subject    string          // mTLS client DN from Envoy header
	TTL        time.Duration   // 10 minutes
	Now        time.Time
}

// IssueToken signs a new G2-JWT with RS256 using Gate 2's private key.
// A unique JTI is generated per token to allow downstream replay detection.
func IssueToken(input IssueInput) (string, error) {
	jti, err := newJTI()
	if err != nil {
		return "", fmt.Errorf("generating JTI: %w", err)
	}

	claims := jwtlib.RegisteredClaims{
		ID:        jti,
		Issuer:    input.Issuer,
		Subject:   input.Subject,
		Audience:  jwtlib.ClaimStrings{input.Audience},
		IssuedAt:  jwtlib.NewNumericDate(input.Now),
		ExpiresAt: jwtlib.NewNumericDate(input.Now.Add(input.TTL)),
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	return token.SignedString(input.PrivateKey)
}

func newJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
