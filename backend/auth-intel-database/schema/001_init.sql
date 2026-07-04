CREATE TABLE IF NOT EXISTS auth_events (
    id UUID PRIMARY KEY DEFAULT hen_random_uuid(),
    user_id TEXT NOT NULL,
    gate INTEGER NOT NULL,
    status TEXT NOT NULL,
    reason TEXT
    event_timestamp BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_auth_events_user_id ON auth_events(user_id);

CREATE INDEX idx_auth_events_gate ON auth_events (gate);

CREATE INDEX idx_auth_events_status ON auth_events (status);

CREATE INDEX idx_auth_events _timestamp ON auth_events (event_timestamp);