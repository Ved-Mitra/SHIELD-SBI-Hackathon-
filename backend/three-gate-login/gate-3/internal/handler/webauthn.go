package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"shield/gate3/internal/kafka"
	"shield/gate3/internal/store"

	"github.com/go-webauthn/webauthn/webauthn"
)

type SessionStore interface {
	Save(ctx context.Context, key string, data *webauthn.SessionData) error
	Load(ctx context.Context, key string) (*webauthn.SessionData, error)
	Delete(ctx context.Context, key string)
}

// TokenStore is the interface for storing and validating opaque session tokens.
type TokenStore interface {
	StoreToken(ctx context.Context, token, userID string) error
	ValidateToken(ctx context.Context, token string) (string, error)
}

type WebAuthnHandler struct {
	WebAuthn  *webauthn.WebAuthn
	Sessions  SessionStore
	UserStore store.UserStore
	Tokens    TokenStore
	MockFido2 bool
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
		go kafka.PublishEvent(kafka.AuthEvent{UserID: req.Username, Gate: 3, Status: "FAILED", Reason: fmt.Sprintf("Server error: %d", http.StatusInternalServerError), TimeStamp: time.Now().UnixMilli()})
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
		go kafka.PublishEvent(kafka.AuthEvent{UserID: username, Gate: 3, Status: "FAILED", Reason: "User not found in DB during RegisterFinish", TimeStamp: time.Now().UnixMilli()})
		return
	}

	sessionData, err := h.Sessions.Load(r.Context(), "reg:"+username)
	if err != nil {
		http.Error(w, "session expired", http.StatusBadRequest)
		go kafka.PublishEvent(kafka.AuthEvent{UserID: username, Gate: 3, Status: "FAILED", Reason: "session expired", TimeStamp: time.Now().UnixMilli()})
		return
	}

	var credential *webauthn.Credential
	if h.MockFido2 {
		credential = &webauthn.Credential{
			ID:              []byte(fmt.Sprintf("mock_cred_%s", username)),
			PublicKey:       []byte("mock_public_key"),
			AttestationType: "mock",
		}
	} else {
		var err error
		credential, err = h.WebAuthn.FinishRegistration(user, *sessionData, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			go kafka.PublishEvent(kafka.AuthEvent{UserID: username, Gate: 3, Status: "FAILED", Reason: fmt.Sprintf("Error: %d", http.StatusUnauthorized), TimeStamp: time.Now().UnixMilli()})
			return
		}
	}

	go kafka.PublishEvent(kafka.AuthEvent{UserID: username, Gate: 3, Status: "PASSED", Reason: "Gate-3 Registration finished", TimeStamp: time.Now().UnixMilli()})

	if err := h.UserStore.AddCredential(r.Context(), username, credential); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.UserStore.Get(r.Context(), req.Username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		go kafka.PublishEvent(kafka.AuthEvent{UserID: req.Username, Gate: 3, Status: "FAILED", Reason: "User not found in DB during AuthBegin", TimeStamp: time.Now().UnixMilli()})
		return
	}

	assertion, sessionData, err := h.WebAuthn.BeginLogin(user)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"BeginLogin failed", "details": "%v"}`, err), http.StatusInternalServerError)
		go kafka.PublishEvent(kafka.AuthEvent{UserID: req.Username, Gate: 3, Status: "FAILED", Reason: fmt.Sprintf("Server error: %v", err), TimeStamp: time.Now().UnixMilli()})
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
		go kafka.PublishEvent(kafka.AuthEvent{UserID: username, Gate: 3, Status: "FAILED", Reason: "session expired", TimeStamp: time.Now().UnixMilli()})
		return
	}

	var credential *webauthn.Credential
	if h.MockFido2 {
		credential = &webauthn.Credential{
			ID: []byte("new_credential_id"),
			Authenticator: webauthn.Authenticator{
				SignCount: 1,
			},
		}
	} else {
		var err error
		credential, err = h.WebAuthn.FinishLogin(user, *sessionData, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			go kafka.PublishEvent(kafka.AuthEvent{UserID: username, Gate: 3, Status: "FAILED", Reason: fmt.Sprintf("Error: %d", http.StatusUnauthorized), TimeStamp: time.Now().UnixMilli()})
			return
		}
	}

	h.UserStore.UpdateCounter(r.Context(), username, credential.ID, credential.Authenticator.SignCount)
	h.Sessions.Delete(r.Context(), "auth:"+username)

	go kafka.PublishEvent(kafka.AuthEvent{UserID: username, Gate: 3, Status: "PASSED", Reason: "Gate-3 Authentication finished", TimeStamp: time.Now().UnixMilli()})

	// Issue opaque session token and store it in Redis (C-5 fix)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, `{"error":"failed to generate session token"}`, http.StatusInternalServerError)
		return
	}
	sessionToken := hex.EncodeToString(b)

	if h.Tokens != nil {
		if err := h.Tokens.StoreToken(r.Context(), sessionToken, string(user.WebAuthnID())); err != nil {
			http.Error(w, `{"error":"failed to store session token"}`, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"session_token": sessionToken,
		"expires_in":    1800,
		"user_id":       string(user.WebAuthnID()),
	})
}
