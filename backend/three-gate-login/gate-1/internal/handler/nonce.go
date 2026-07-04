package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

func GenerateNonce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate nonce")
		return
	}
	
	nonce := base64.RawURLEncoding.EncodeToString(b)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"nonce": nonce})
}
