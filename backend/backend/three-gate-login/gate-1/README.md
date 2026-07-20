# Gate-1: Device Authenticity

## Overview
Gate-1 is the first layer of the SHIELD backend defense. It verifies that the incoming request originates from a genuine, unrooted, unmodified mobile device running the official YONO/SBI app.

## Workflow
1. The mobile app requests a cryptographic `nonce` from Gate-1.
2. The nonce is stored in Redis (with a 5-minute TTL) and returned to the app.
3. The app generates an attestation payload (Apple App Attest or Google Play Integrity) incorporating the nonce, and sends it to Gate-1.
4. Gate-1 validates the attestation (verifying the nonce using Redis `SETNX` to prevent replay attacks).
5. If valid, Gate-1 issues a short-lived RS256 JWT (`G1-JWT`) required for Gate-2.

## Current Implementation Details
- **Language**: Go
- **Rate Limiting**: Token-bucket rate limiter per IP (10 sustained req/min, burst 20).
- **Timeouts**: Enforces strict read, write, and idle timeouts on the HTTP server.
- **Mock Mode**: Supports `GATE1_MOCK_ATTESTATION=true` for local development/hackathons to bypass actual Apple/Google API checks.
- **Kafka Auditing**: Publishes success/failure/rate-limit events to the `auth-events` Kafka topic.
