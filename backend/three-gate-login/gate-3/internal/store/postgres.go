package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(dsn string) (*PostgresUserStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresUserStore{db: db}, nil
}

func (s *PostgresUserStore) GetOrCreate(ctx context.Context, username, displayName string) (*User, error) {
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, "SELECT id FROM webauthn_users WHERE username = $1", username).Scan(&id)
	if err == sql.ErrNoRows {
		// Create
		err = s.db.QueryRowContext(ctx, "INSERT INTO webauthn_users (username, display_name) VALUES ($1, $2) RETURNING id", username, displayName).Scan(&id)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return s.getUserWithCreds(ctx, id, username, displayName)
}

func (s *PostgresUserStore) Get(ctx context.Context, username string) (*User, error) {
	var id uuid.UUID
	var displayName string
	err := s.db.QueryRowContext(ctx, "SELECT id, display_name FROM webauthn_users WHERE username = $1", username).Scan(&id, &displayName)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, err
	}

	return s.getUserWithCreds(ctx, id, username, displayName)
}

func (s *PostgresUserStore) getUserWithCreds(ctx context.Context, id uuid.UUID, username, displayName string) (*User, error) {
	user := &User{
		id:          id[:],
		username:    username,
		displayName: displayName,
		credentials: []webauthn.Credential{},
	}

	rows, err := s.db.QueryContext(ctx, "SELECT public_key FROM webauthn_credentials WHERE user_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var credJSON []byte
		if err := rows.Scan(&credJSON); err != nil {
			return nil, err
		}
		var cred webauthn.Credential
		if err := json.Unmarshal(credJSON, &cred); err != nil {
			continue // skip malformed
		}
		user.credentials = append(user.credentials, cred)
	}

	return user, nil
}

func (s *PostgresUserStore) AddCredential(ctx context.Context, username string, cred *webauthn.Credential) error {
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, "SELECT id FROM webauthn_users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return err
	}

	credJSON, err := json.Marshal(cred)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO webauthn_credentials (id, user_id, public_key, sign_count) VALUES ($1, $2, $3, $4)",
		cred.ID, id, credJSON, cred.Authenticator.SignCount,
	)
	return err
}

func (s *PostgresUserStore) UpdateCounter(ctx context.Context, username string, credentialID []byte, signCount uint32) error {
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, "SELECT id FROM webauthn_users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return err
	}

	// Update the sign_count column, and also we need to update the JSON payload for completeness
	// (or just rely on the JSON payload. We'll update the JSON payload for simplicity).
	
	var credJSON []byte
	err = s.db.QueryRowContext(ctx, "SELECT public_key FROM webauthn_credentials WHERE user_id = $1 AND id = $2", id, credentialID).Scan(&credJSON)
	if err != nil {
		return err
	}

	var cred webauthn.Credential
	if err := json.Unmarshal(credJSON, &cred); err != nil {
		return err
	}
	
	cred.Authenticator.SignCount = signCount
	updatedJSON, err := json.Marshal(cred)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE webauthn_credentials SET sign_count = $1, public_key = $2, last_used_at = $3 WHERE user_id = $4 AND id = $5",
		signCount, updatedJSON, time.Now(), id, credentialID,
	)
	return err
}
