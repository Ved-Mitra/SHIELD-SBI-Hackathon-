# Threat Intel Database

## Overview
This module consumes threat intelligence signals (phishing URLs) from the Kafka `url-events` topic and persists them into a PostgreSQL database for the Grafana dashboard and historical auditing.

## Workflow
1. Connects to Kafka as a consumer group (`threat-intel-db-group`).
2. Listens to `url-events` (published by `risk-url-engine`).
3. Upserts records into the `threat_intel` table.

## Current Implementation Details
- **Language**: Go 1.26.3
- **Database**: PostgreSQL (`intel_db`)
- **Resilience**: 
  - Connection pooling enabled (max 10 open, 5 idle).
  - Graceful shutdown on SIGTERM.
  - Automatic upsert (`ON CONFLICT (indicator_type, indicator_value) DO UPDATE`) to prevent duplicate rows for the same URL, instead updating the `last_seen` timestamp and `evidence` JSON.
  - Kafka consumer ignores malformed JSON without crashing.
- **Docker**: Deploys as `thread-intel-consumer` in `docker-compose.yml`.

## Database Schema
- Managed via `schema/001_init.sql` using idempotent `CREATE TYPE` and `CREATE TABLE IF NOT EXISTS`.
- Key columns: `indicator_type`, `indicator_value`, `confidence`, `severity`, `status`.
