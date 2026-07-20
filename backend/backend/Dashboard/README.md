# Dashboard

## Overview
The SHIELD Dashboard provides real-time visualization of threat intelligence (phishing URLs detected) and authentication audit events (login successes, failures, and brute-force attempts).

## Current Implementation Details
- **Platform**: Grafana
- **Data Source**: PostgreSQL (`intel_db`) containing the `threat_intel` and `auth_events` tables populated by the Kafka consumers.
- **Docker**: Deploys as the `dashboard` service on port `3000`.

## Accessing the Dashboard
- URL: `http://localhost:3000`
- Default Credentials: `admin` / `admin` (Configured in `docker-compose.yml`)

## Key Metrics Visualized
- High/Critical severity phishing URLs intercepted by the on-device engine.
- WebAuthn registration and authentication events across all three gates.
- Rate-limit events (potential brute force attacks).
