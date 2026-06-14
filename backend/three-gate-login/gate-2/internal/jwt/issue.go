package jwt

import (
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type IssueInput struct {
	Secret   []byte
	Issuer   string
	Audience string
	Subject  string
	TTL      time.Duration
	Now      time.Time
}

func IssueToken(input IssueInput) (string, error) {
	claims := jwtlib.RegisteredClaims{
		Issuer:    input.Issuer,
		Subject:   input.Subject,
		Audience:  jwtlib.ClaimStrings{input.Audience},
		IssuedAt:  jwtlib.NewNumericDate(input.Now),
		ExpiresAt: jwtlib.NewNumericDate(input.Now.Add(input.TTL)),
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(input.Secret)
}

