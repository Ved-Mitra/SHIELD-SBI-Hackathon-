package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type UrlEvent struct {
	URL       string `json:"url"`
	DeviceID  string `json:"device_id"`
	Timestamp int64  `json:"timestamp"`
}

type ThreatStore struct {
	db *sql.DB
}

func NewThreatStore(dsn string) (*ThreatStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// M-9 fix: configure connection pool to prevent exhausting Postgres
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &ThreatStore{db: db}, nil
}

func (s *ThreatStore) InsertThreat(event UrlEvent) error {
	evidenceJSON, err := json.Marshal(map[string]any{
		"device_id": event.DeviceID,
		"timestamp": event.Timestamp,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}

	// M-8 fix: upsert — if the same URL is seen again, update last_seen and
	// merge new evidence instead of creating a duplicate row.
	query := `
		INSERT INTO threat_intel (indicator_type, indicator_value, source, confidence, severity, evidence, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (indicator_type, indicator_value)
		DO UPDATE SET
			last_seen  = NOW(),
			evidence   = EXCLUDED.evidence,
			status     = CASE
				WHEN threat_intel.status = 'resolved' THEN 'new'
				ELSE threat_intel.status
			END
	`
	_, err = s.db.Exec(query,
		"url",
		event.URL,
		"device_ml",
		80,
		"high",
		evidenceJSON,
		"new",
	)
	return err
}

func (s *ThreatStore) Close() error {
	return s.db.Close()
}
