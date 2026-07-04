# Gate-3: User Authenticity

## Overview
Gate-3 is the final layer of the SHIELD backend defense. It authenticates the actual human user using FIDO2/WebAuthn biometrics (e.g., fingerprint, FaceID). This provides 100% resistance against password-based phishing, as cryptographic keys never leave the device enclave.

## Workflow
1. Client presents the `G2-JWT`.
2. **Registration**: The user registers a biometric passkey with the server.
3. **Authentication**: The server issues a challenge (stored in Redis). The device signs it with the private enclave key, and Gate-3 verifies the signature using the stored public key.
4. On success, Gate-3 issues a final 256-bit `session_token` with a 30-minute TTL, stored in Redis.

## Current Implementation Details
- **Language**: Go
- **Storage**: User FIDO2 credentials and session tokens are stored in PostgreSQL (`intel_db`) and Redis.
- **Rate Limiting**: Token-bucket rate limiter per IP (20 req / 3 sec, burst 20).
- **Mock Mode**: Supports `GATE3_MOCK_FIDO2=true` for local development to bypass actual hardware key ceremonies.
- **Kafka Auditing**: Publishes success/failure/rate-limit events to the `auth-events` Kafka topic.
