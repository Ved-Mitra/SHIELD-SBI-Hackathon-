# Gate-2: Channel Authenticity

## Overview
Gate-2 is the second layer of the SHIELD backend defense. It ensures that the network channel is secure and that the client communicating with the backend is the legitimate application, preventing Man-In-The-Middle (MITM) and traffic interception attacks.

## Workflow
1. The client establishes a Mutual TLS (mTLS) connection with an Envoy proxy.
2. The Envoy proxy validates the client's TLS certificate.
3. The Envoy proxy forwards the request to the Gate-2 Go application, passing the client identity in headers.
4. Gate-2 receives the request, validates the `G1-JWT` (issued by Gate-1), and issues a new RS256 JWT (`G2-JWT`).

## Current Implementation Details
- **Proxy**: Envoy proxy running on port 8443 terminates mTLS.
- **Language**: Go microservice for JWT validation and issuance.
- **JWT Chaining**: Strictly validates the signature, issuer, and expiration of the incoming `G1-JWT`.
- **Rate Limiting**: Token-bucket rate limiter per IP (20 req / 3 sec, burst 20).
- **Kafka Auditing**: Publishes success/failure/rate-limit events to the `auth-events` Kafka topic.
