package jwt

import (
	"crypto/rsa"
	"fmt"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// VerifyToken validates a JWT signed with RS256.
// Used by Gate 3 to verify incoming G2-JWTs.
// gate2PublicKey is Gate 2's public key (distributed to Gate 3).
func VerifyToken(token string, gate2PublicKey *rsa.PublicKey, issuer, audience string) (*jwtlib.Token, error) {
	claims := &jwtlib.RegisteredClaims{}

	parsed, err := jwtlib.ParseWithClaims(token, claims,
		func(t *jwtlib.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwtlib.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return gate2PublicKey, nil
		},
		jwtlib.WithIssuer(issuer),
		jwtlib.WithAudience(audience),
		jwtlib.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}
