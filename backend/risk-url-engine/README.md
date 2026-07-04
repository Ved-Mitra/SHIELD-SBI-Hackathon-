# Risk URL Engine

## Overview
The Risk URL Engine acts as the backend receiver for the on-device SHIELD phishing detector. When the mobile app detects a phishing URL with high confidence, it reports it to this endpoint.

## Workflow
1. Exposes an HTTP API (`POST /report`).
2. Receives a JSON payload containing the phishing `url`, `device_id`, and `timestamp` (Unix Milli).
3. Publishes the event to the Kafka `url-events` topic for downstream processing.

## Current Implementation Details
- **Language**: Go 1.26.3
- **Framework**: Standard `net/http` with robust timeouts (Read/Write/Idle).
- **Kafka Integration**: Uses `segmentio/kafka-go`. Handles `nil` writer errors gracefully to prevent panics.
- **Docker**: Deploys as `risk-url-engine` on port `8083`.

## API Endpoints
- `GET /healthz`: Health check endpoint.
- `POST /report`: Accepts phishing reports.
  - Example payload: `{"url": "http://evil.com", "device_id": "abc-123", "timestamp": 1720000000000}`
