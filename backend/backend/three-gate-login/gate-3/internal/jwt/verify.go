package jwt

import (
	"crypto/rsa"
	"fmt"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

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
