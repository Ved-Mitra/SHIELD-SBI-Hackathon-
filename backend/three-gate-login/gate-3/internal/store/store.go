package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// User represents a WebAuthn user
type User struct {
	id          []byte
	username    string
	displayName string
	credentials []webauthn.Credential
}

func (u *User) WebAuthnID() []byte { return u.id }
func (u *User) WebAuthnName() string { return u.username }
func (u *User) WebAuthnDisplayName() string { return u.displayName }
func (u *User) WebAuthnIcon() string { return "" }
func (u *User) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

type UserStore interface {
	GetOrCreate(ctx context.Context, username, displayName string) (*User, error)
	Get(ctx context.Context, username string) (*User, error)
	AddCredential(ctx context.Context, username string, cred *webauthn.Credential) error
	UpdateCounter(ctx context.Context, username string, credentialID []byte, signCount uint32) error
}

type InMemoryUserStore struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

func (s *InMemoryUserStore) GetOrCreate(ctx context.Context, username, displayName string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		user = &User{
			id:          []byte(username), // Simple ID for mock
			username:    username,
			displayName: displayName,
			credentials: []webauthn.Credential{},
		}
		s.users[username] = user
	}
	return user, nil
}

func (s *InMemoryUserStore) Get(ctx context.Context, username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *InMemoryUserStore) AddCredential(ctx context.Context, username string, cred *webauthn.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}
	user.credentials = append(user.credentials, *cred)
	return nil
}

func (s *InMemoryUserStore) UpdateCounter(ctx context.Context, username string, credentialID []byte, signCount uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}

	for i, cred := range user.credentials {
		if string(cred.ID) == string(credentialID) {
			user.credentials[i].Authenticator.SignCount = signCount
			return nil
		}
	}
	return fmt.Errorf("credential not found")
}

// MemorySessionStore implements session.Store locally
type MemorySessionStore struct {
	mu       sync.Mutex
	sessions map[string]sessionEntry
}

type sessionEntry struct {
	data webauthn.SessionData
	exp  time.Time
}

func NewMemorySessionStore() *MemorySessionStore {
	s := &MemorySessionStore{
		sessions: make(map[string]sessionEntry),
	}
	go s.cleanup()
	return s
}

func (s *MemorySessionStore) Save(ctx context.Context, key string, data *webauthn.SessionData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[key] = sessionEntry{data: *data, exp: time.Now().Add(5 * time.Minute)}
	return nil
}

func (s *MemorySessionStore) Load(ctx context.Context, key string) (*webauthn.SessionData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.sessions[key]
	if !ok || time.Now().After(entry.exp) {
		delete(s.sessions, key)
		return nil, fmt.Errorf("session not found or expired")
	}
	return &entry.data, nil
}

func (s *MemorySessionStore) Delete(ctx context.Context, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, key)
}

func (s *MemorySessionStore) cleanup() {
	for {
		time.Sleep(1 * time.Minute)
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.sessions {
			if now.After(v.exp) {
				delete(s.sessions, k)
			}
		}
		s.mu.Unlock()
	}
}
