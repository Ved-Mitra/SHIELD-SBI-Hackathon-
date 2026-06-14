package model

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	TokenType string `json:"token_type"`
}

