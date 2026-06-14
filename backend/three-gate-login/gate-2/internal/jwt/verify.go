package jwt

import (
	fmt "fmt"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func VerifyToken(token string, secret []byte, issuer, audience string) (*jwtlib.Token, error) {
	claims := &jwtlib.RegisteredClaims{}
	parsed, err := jwtlib.ParseWithClaims(token, claims, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if err := jwtlib.NewValidator(jwtlib.WithIssuer(issuer), jwtlib.WithAudience(audience)).Validate(claims); err != nil {
		return nil, err
	}

	return parsed, nil
}
