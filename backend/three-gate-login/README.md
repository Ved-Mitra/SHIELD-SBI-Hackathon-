# Three-Gate Login System

## Overview
This directory contains the core authentication backend for the SHIELD project. To provide unparalleled security against credential theft and phishing, SHIELD employs a chained three-gate authentication architecture.

## The Three Gates
1. **[Gate-1: Device Authenticity](./gate-1)**: Validates that the requesting device is a genuine, uncompromised, unrooted mobile device (using Apple App Attest / Google Play Integrity).
2. **[Gate-2: Channel Authenticity](./gate-2)**: Validates that the communication channel is secure and explicitly tied to the YONO app using Mutual TLS (mTLS) via an Envoy proxy.
3. **[Gate-3: User Authenticity](./gate-3)**: Validates the human user using FIDO2/WebAuthn biometrics (phishing-resistant, hardware-backed keys).

## Authentication Flow
The gates must be passed sequentially.
- **Gate-1** issues a short-lived `G1-JWT`.
- **Gate-2** requires the `G1-JWT` over mTLS and issues a `G2-JWT`.
- **Gate-3** requires the `G2-JWT`, performs biometric authentication, and finally issues the `session_token` (stored in Redis).

## Event Auditing
Every request to any gate is logged to the Kafka `auth-events` topic, providing a complete, immutable audit trail of successes, failures, and brute-force attempts.
