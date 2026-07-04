# Auth Intel Database

## Overview
This module consumes authentication audit logs from the Kafka `auth-events` topic and persists them into a PostgreSQL database. It tracks all passes, failures, and brute-force attempts across all three SHIELD authentication gates.

## Workflow
1. Connects to Kafka as a consumer group (`auth-intel-db-group`).
2. Listens to `auth-events` (published by Gate-1, Gate-2, and Gate-3).
3. Inserts records into the `auth_events` table.

## Current Implementation Details
- **Language**: Go 1.26.3
- **Database**: PostgreSQL (`intel_db`)
- **Resilience**: 
  - Connection pooling enabled (max 10 open, 5 idle) using `github.com/lib/pq`.
  - Graceful shutdown via `os.Signal` handling and Kafka reader context cancellation.
  - 10-retry startup loop for database connections.
- **Docker**: Deploys as `auth-intel-consumer` in `docker-compose.yml`.

## Database Schema
- Schema is managed in `schema/001_init.sql`.
- Key columns: `user_id`, `gate`, `status`, `reason`, `event_timestamp`.
