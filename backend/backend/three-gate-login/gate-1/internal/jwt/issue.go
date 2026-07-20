package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// IssueInput carries all parameters needed to create a G1-JWT.
type IssueInput struct {
	PrivateKey *rsa.PrivateKey
	Issuer     string
	Audience   string
	Subject    string // e.g. "android:com.sbi.yono" or "ios:TEAMID.com.sbi.yono"
	TTL        time.Duration
	Now        time.Time
}

// IssueToken signs a new JWT with RS256 using the provided private key.
// A unique JTI (JWT ID) is generated per token to enable replay detection downstream.
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

// newJTI generates a cryptographically-random 16-byte hex JWT ID.
func newJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
