-- Thread-Intel Database Initial Schema

-- Enum for Indicator Types
CREATE TYPE indicator_type_enum AS ENUM ('domain', 'url', 'ip', 'handle', 'hash');

-- Enum for Source
CREATE TYPE source_enum AS ENUM ('crawler', 'manual', 'partner', 'device_ml');

-- Enum for Severity
CREATE TYPE severity_enum AS ENUM ('low', 'medium', 'high', 'critical');

-- Enum for Status
CREATE TYPE status_enum AS ENUM ('new', 'in_review', 'takedown_requested', 'resolved', 'false_positive');

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
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
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
