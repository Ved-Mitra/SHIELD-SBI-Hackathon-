package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

	query := `
		INSERT INTO threat_intel (indicator_type, indicator_value, source, confidence, severity, evidence, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
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
