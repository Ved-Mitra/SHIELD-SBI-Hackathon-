package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"shield/three-gate-login/internal/config"
	"shield/three-gate-login/internal/jwt"
	"shield/three-gate-login/internal/model"
	"shield/three-gate-login/internal/mtls"
)

type TokenHandler struct {
	Config config.Config
}

func (h TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	clientID, ok := mtls.ClientIdentity(r, h.Config.ClientIDHeader)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := jwt.IssueToken(jwt.IssueInput{
		Secret:   h.Config.JWTSecret,
		Issuer:   h.Config.JWTIssuer,
		Audience: h.Config.JWTAudience,
		Subject:  clientID,
		TTL:      h.Config.JWTTTL,
		Now:      time.Now(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := model.TokenResponse{
		Token:     token,
		ExpiresIn: int(h.Config.JWTTTL.Seconds()),
		TokenType: "Bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

