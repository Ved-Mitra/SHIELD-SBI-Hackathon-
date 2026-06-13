package model

// AttestRequest is the unified request body for POST /gate1/attest.
// The client sends either Android or iOS-specific fields depending on the platform.
type AttestRequest struct {
	// Platform must be "android" or "ios".
	Platform string `json:"platform"`

	// --- Android fields ---
	// IntegrityToken is the base64-encoded Play Integrity token returned by the Android SDK.
	IntegrityToken string `json:"integrity_token,omitempty"`

	// --- iOS fields ---
	// AttestObject is the base64url-encoded CBOR attestation object from Apple's DCAppAttestService.
	AttestObject string `json:"attest_object,omitempty"`
	// ClientDataHash is base64(SHA-256(nonce)) — the client must hash our nonce before passing
	// it to attestKey(key:clientDataHash:).
	ClientDataHash string `json:"client_data_hash,omitempty"`
	// KeyID is the base64url key identifier returned by DCAppAttestService.generateKey().
	KeyID string `json:"key_id,omitempty"`

	// Nonce is the server-issued 32-byte hex nonce (both platforms).
	// Must be consumed exactly once; replayed nonces are rejected.
	Nonce string `json:"nonce"`
}

// AttestResponse is returned on successful attestation.
type AttestResponse struct {
	// Token is the signed G1-JWT (RS256, aud="shield-gate2", TTL=2m).
	Token string `json:"token"`
	// ExpiresIn is the number of seconds until the token expires.
	ExpiresIn int `json:"expires_in"`
	// TokenType is always "Bearer".
	TokenType string `json:"token_type"`
}

// ErrorResponse is returned for all error conditions.
type ErrorResponse struct {
	Error string `json:"error"`
}
