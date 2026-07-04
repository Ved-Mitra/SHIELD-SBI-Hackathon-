package store

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type AuthEvent struct{
	UserID string `json:"user_id"`
	Gate int `json:"gate"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	Timestamp int64 `json:"timestamp"`
}

type AuthStore struct {
	db *sql.DB
}

func NewAuthStore(dsn string) (*AuthStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// A-4 fix: configure connection pool to prevent exhausting Postgres
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &AuthStore{db: db}, nil
}

func (s *AuthStore) InsertAuth(event AuthEvent) error {
	query := `
		INSERT INTO auth_events (user_id, gate, status, reason, event_timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.Exec(query,
		event.UserID,
		event.Gate,
		event.Status,
		event.Reason,
		event.Timestamp,
	)
	return err
}


func (s* AuthStore) Close() error {
	return s.db.Close()
}