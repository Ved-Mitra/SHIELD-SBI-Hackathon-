package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"shield/gate3/internal/store"
)

type WebAuthnHandler struct {
	WebAuthn  *webauthn.WebAuthn
	Sessions  *store.MemorySessionStore
	UserStore *store.InMemoryUserStore
}

func (h *WebAuthnHandler) RegisterBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.UserStore.GetOrCreate(r.Context(), req.Username, req.DisplayName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	creation, sessionData, err := h.WebAuthn.BeginRegistration(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Sessions.Save(r.Context(), "reg:"+req.Username, sessionData)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creation)
}

func (h *WebAuthnHandler) RegisterFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	user, err := h.UserStore.Get(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sessionData, err := h.Sessions.Load(r.Context(), "reg:"+username)
	if err != nil {
		http.Error(w, "session expired", http.StatusBadRequest)
		return
	}

	credential, err := h.WebAuthn.FinishRegistration(user, *sessionData, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.UserStore.AddCredential(r.Context(), username, credential)
	h.Sessions.Delete(r.Context(), "reg:"+username)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"credential_id": credential.ID,
		"registered":    true,
	})
}

func (h *WebAuthnHandler) AuthBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct{ Username string `json:"username"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.UserStore.Get(r.Context(), req.Username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	assertion, sessionData, err := h.WebAuthn.BeginLogin(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Sessions.Save(r.Context(), "auth:"+req.Username, sessionData)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assertion)
}

func (h *WebAuthnHandler) AuthFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	user, err := h.UserStore.Get(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sessionData, err := h.Sessions.Load(r.Context(), "auth:"+username)
	if err != nil {
		http.Error(w, "session expired", http.StatusBadRequest)
		return
	}

	credential, err := h.WebAuthn.FinishLogin(user, *sessionData, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.UserStore.UpdateCounter(r.Context(), username, credential.ID, credential.Authenticator.SignCount)
	h.Sessions.Delete(r.Context(), "auth:"+username)

	// Issue opaque session token
	b := make([]byte, 32)
	rand.Read(b)
	sessionToken := hex.EncodeToString(b)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"session_token": sessionToken,
		"expires_in":    1800,
		"user_id":       string(user.WebAuthnID()),
	})
}
