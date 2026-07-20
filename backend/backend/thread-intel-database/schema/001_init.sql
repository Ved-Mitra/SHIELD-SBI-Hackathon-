-- Thread-Intel Database Initial Schema

-- Enum for Indicator Types
DO $$ BEGIN
    CREATE TYPE indicator_type_enum AS ENUM ('domain', 'url', 'ip', 'handle', 'hash');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Enum for Source
DO $$ BEGIN
    CREATE TYPE source_enum AS ENUM ('crawler', 'manual', 'partner', 'device_ml');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Enum for Severity
DO $$ BEGIN
    CREATE TYPE severity_enum AS ENUM ('low', 'medium', 'high', 'critical');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Enum for Status
DO $$ BEGIN
    CREATE TYPE status_enum AS ENUM ('new', 'in_review', 'takedown_requested', 'resolved', 'false_positive');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Table: Threat Intel
CREATE TABLE IF NOT EXISTS threat_intel (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    indicator_type indicator_type_enum NOT NULL,
    indicator_value TEXT NOT NULL,
    source source_enum NOT NULL DEFAULT 'manual',
    confidence INTEGER CHECK (confidence >= 0 AND confidence <= 100) DEFAULT 50,
    severity severity_enum NOT NULL DEFAULT 'medium',
    evidence JSONB,
    status status_enum NOT NULL DEFAULT 'new',
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- M-10 fix: UNIQUE constraint required for ON CONFLICT upsert in threat_store.go
    UNIQUE (indicator_type, indicator_value)
);

-- Indexes for efficient querying
CREATE INDEX idx_threat_intel_indicator ON threat_intel (indicator_type, indicator_value);
CREATE INDEX idx_threat_intel_status ON threat_intel (status);
CREATE INDEX idx_threat_intel_severity ON threat_intel (severity);
CREATE INDEX idx_threat_intel_last_seen ON threat_intel (last_seen);
CREATE INDEX idx_threat_intel_evidence ON threat_intel USING GIN (evidence);

-- Trigger to auto-update 'updated_at' timestamp
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_threat_intel_modtime
    BEFORE UPDATE ON threat_intel
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_column();
