# SHIELD — Three-Gate Login: Backend Implementation Guide

> **Context:** This document covers the full backend design and implementation plan for all three
> authentication gates in the SHIELD system. Gate 2 (mTLS) is partially implemented. Gates 1 and 3
> need to be built. Read this before writing any new code.

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Gate 1 — App Authenticity (Play Integrity / App Attest)](#2-gate-1--app-authenticity)
3. [Gate 2 — Channel Authenticity (mTLS + JWT)](#3-gate-2--channel-authenticity-mtls--jwt)
4. [Gate 3 — User Authenticity (FIDO2 WebAuthn)](#4-gate-3--user-authenticity-fido2-webauthn)
5. [Cross-Gate Token Chain](#5-cross-gate-token-chain)
6. [Shared Infrastructure](#6-shared-infrastructure)
7. [Security Hardening Checklist](#7-security-hardening-checklist)
8. [Recommended Folder Structure](#8-recommended-folder-structure)

---

## 1. System Overview

The three-gate pipeline is a **sequential, cryptographic authentication chain**. A user/device must
pass **all three gates in order**. Each gate issues a short-lived token that is required as input
to the next gate. No gate can be skipped.

```
Mobile Client
     │
     │  Gate 1: Attest the APP binary is genuine (not a phishing clone)
     ▼
┌─────────────────────────┐
│  Gate 1 Backend         │  POST /gate1/attest
│  (Go or Python/FastAPI) │  Input : Play Integrity token / Apple App Attest token
│                         │  Output: short-lived G1-JWT  (aud = "shield-gate2")
└───────────┬─────────────┘
            │  G1-JWT in Authorization header
            │  Gate 2: Attest the CHANNEL is genuine (mTLS cert pinned)
            ▼
┌─────────────────────────┐
│  Gate 2 Backend (Go)    │  POST /gate2/token    ← already partially built
│  behind Envoy mTLS      │  Input : mTLS client cert + G1-JWT
│                         │  Output: short-lived G2-JWT  (aud = "shield-gate3")
└───────────┬─────────────┘
            │  G2-JWT in Authorization header
            │  Gate 3: Attest the USER is genuine (FIDO2 hardware key / biometric)
            ▼
┌─────────────────────────┐
│  Gate 3 Backend (Go)    │  POST /gate3/register   (first time)
│                         │  POST /gate3/authenticate/begin
│                         │  POST /gate3/authenticate/finish
│                         │  Output: session token + user identity
└─────────────────────────┘
            │
            ▼
      YONO Banking Session established
```

### Token Lifetime Policy

| Token   | TTL    | Signing Algorithm | Purpose                       |
|---------|--------|-------------------|-------------------------------|
| G1-JWT  | 2 min  | RS256             | Proves app binary is genuine  |
| G2-JWT  | 10 min | RS256             | Proves channel is mTLS-pinned |
| Session | 30 min | opaque / Redis    | Active banking session        |

> **Why RS256 not HS256?**
> Gate 2 currently uses HS256 with a shared secret. This means Gate 3 must hold the same secret,
> creating a single point of compromise. RS256 uses a private key (Gate N signs) and public key
> (Gate N+1 verifies). Each gate service only needs its own private key and the *previous* gate's
> public key. **Switch Gate 2 to RS256 before demo.**

---

## 2. Gate 1 — App Authenticity

### Purpose
Verifies that the mobile binary making the request is the **legitimate, unmodified SBI YONO app**
published on the Google Play Store or Apple App Store — not a phishing clone or repackaged APK.

### Platform Providers

| Platform | Technology              | Trust Root          |
|----------|-------------------------|---------------------|
| Android  | Google Play Integrity API | Google's servers  |
| iOS      | Apple App Attest        | Apple's servers     |

### API Contract

#### `POST /gate1/attest`

**Request Body (Android):**
```json
{
  "platform": "android",
  "integrity_token": "<base64-encoded Play Integrity token>",
  "nonce": "<32-byte hex client-generated nonce>"
}
```

**Request Body (iOS):**
```json
{
  "platform": "ios",
  "attest_object": "<base64url-encoded CBOR attestation object>",
  "client_data_hash": "<SHA-256 of nonce, base64>",
  "key_id": "<base64url key identifier>"
}
```

**Success Response `200 OK`:**
```json
{
  "token": "<G1-JWT>",
  "expires_in": 120,
  "token_type": "Bearer"
}
```

**Error Responses:**
- `400 Bad Request` — missing/malformed fields
- `401 Unauthorized` — attestation verdict failed
- `429 Too Many Requests` — rate limit exceeded
- `502 Bad Gateway` — upstream Google/Apple API unreachable

### Android: Play Integrity Verification Flow

```
Client                          Gate 1 Backend              Google Play Integrity
  │                                   │                              │
  │── POST /gate1/attest ────────────►│                              │
  │   { platform, integrity_token,    │                              │
  │     nonce }                       │                              │
  │                                   │── DecryptIntegrityToken ────►│
  │                                   │   (using Google API key)     │
  │                                   │◄─ Verdict JSON ──────────────│
  │                                   │                              │
  │                                   │  Validate:
  │                                   │  ✓ requestDetails.nonce == our nonce
  │                                   │  ✓ appIntegrity.appRecognitionVerdict == "PLAYS_RECOGNIZED"
  │                                   │  ✓ deviceIntegrity.deviceRecognitionVerdict has "MEETS_DEVICE_INTEGRITY"
  │                                   │  ✓ accountDetails.appLicensingVerdict == "LICENSED"
  │                                   │  ✓ requestDetails.requestPackageName == "com.sbi.yono"
  │                                   │
  │◄── G1-JWT ────────────────────────│
```

**Key Verdict Fields to Check:**
```
appIntegrity.appRecognitionVerdict:
  PLAY_RECOGNIZED      → ✅ pass
  UNRECOGNIZED_VERSION → ❌ reject (sideloaded or unofficial version)
  UNEVALUATED          → ❌ reject

deviceIntegrity.deviceRecognitionVerdict:
  Must contain MEETS_DEVICE_INTEGRITY
  MEETS_STRONG_INTEGRITY preferred (hardware-backed keystore)
```

### iOS: App Attest Verification Flow

```
Client                          Gate 1 Backend              Apple Attestation Service
  │                                   │                              │
  │  (First time: Register key)       │                              │
  │── POST /gate1/attest ────────────►│                              │
  │   { attest_object, key_id,        │── VerifyAttestation ────────►│
  │     client_data_hash }            │   (using apple-app-attest    │
  │                                   │    server library)           │
  │                                   │◄─ Credential ID, public key ─│
  │                                   │                              │
  │                                   │  Validate:
  │                                   │  ✓ App ID == "TEAMID.com.sbi.yono"
  │                                   │  ✓ rpIdHash matches our app
  │                                   │  ✓ counter == 0 (new attestation)
  │                                   │  ✓ Store public key in DB against key_id
  │                                   │
  │◄── G1-JWT ────────────────────────│
```

### Go Implementation Sketch

**File:** `backend/gate1/internal/handler/attest.go`

```go
package handler

import (
    "encoding/json"
    "net/http"
    "time"

    "shield/gate1/internal/config"
    "shield/gate1/internal/integrity"  // Play Integrity client
    "shield/gate1/internal/appattest" // Apple App Attest client
    "shield/gate1/internal/jwt"
    "shield/gate1/internal/nonce"
)

type AttestHandler struct {
    Config config.Config
}

func (h AttestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var req AttestRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    // 1. Validate nonce freshness (prevent replay)
    if !nonce.Consume(r.Context(), req.Nonce) {
        http.Error(w, "invalid or replayed nonce", http.StatusUnauthorized)
        return
    }

    // 2. Verify with platform provider
    var subject string
    var err error
    switch req.Platform {
    case "android":
        subject, err = integrity.Verify(r.Context(), req.IntegrityToken, req.Nonce, h.Config.PackageName)
    case "ios":
        subject, err = appattest.Verify(r.Context(), req.AttestObject, req.ClientDataHash, req.KeyID, h.Config.AppID)
    default:
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    // 3. Issue G1-JWT (RS256)
    token, err := jwt.IssueRS256(jwt.IssueInput{
        PrivateKey: h.Config.Gate1PrivateKey,
        Issuer:     "shield-gate1",
        Audience:   "shield-gate2",
        Subject:    subject, // e.g., "android:com.sbi.yono"
        TTL:        2 * time.Minute,
    })
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]any{
        "token":      token,
        "expires_in": 120,
        "token_type": "Bearer",
    })
}
```

### Config / Environment Variables (Gate 1)

| Variable                                  | Default              | Description                                      |
|-------------------------------------------|----------------------|--------------------------------------------------|
| `GATE1_ADDR`                              | `:8081`              | Listen address                                   |
| `GATE1_PLAY_INTEGRITY_DECRYPTION_KEY`     | —                    | **Required.** Base64 Play Integrity decryption key |
| `GATE1_PLAY_INTEGRITY_VERIFICATION_KEY`   | —                    | **Required.** Base64 Play Integrity verification key |
| `GATE1_ANDROID_PACKAGE_NAME`              | `com.sbi.yono`       | Expected Android package name                    |
| `GATE1_IOS_APP_ID`                        | `<TEAMID>.com.sbi.yono` | Expected iOS App ID                           |
| `GATE1_JWT_PRIVATE_KEY_PATH`              | `certs/gate1/private.pem` | RS256 private key (PEM)                    |
| `GATE1_NONCE_STORE_ADDR`                  | `redis:6379`         | Redis address for nonce replay prevention        |

### Nonce / Replay Prevention

- Client generates a 32-byte cryptographically random nonce before each attest call.
- Backend: on receipt, check if nonce exists in Redis. If yes → reject (replay). If no → store
  with 5-minute TTL, then proceed.
- This prevents an attacker from capturing a valid integrity token and replaying it.

---

## 3. Gate 2 — Channel Authenticity (mTLS + JWT)

> Gate 2 is **partially implemented**. The Go service and Envoy config exist. The following section
> covers what needs to change and what the production-ready version looks like.

### Current State

- ✅ Envoy mTLS termination — client cert required, DN forwarded as `x-client-dn`
- ✅ Go service issues JWT from client DN
- ✅ Basic middleware: request ID, logging
- ❌ JWT uses **HS256** — must be upgraded to RS256
- ❌ No validation of the **G1-JWT** from Gate 1
- ❌ No rate limiting or replay protection
- ❌ Server started with bare `http.ListenAndServe` (no timeouts)
- ❌ No fail-fast on missing `GATE2_JWT_SECRET` env var

### Required Changes

#### 3.1 Upgrade JWT Signing to RS256

Replace `jwt/issue.go`:

```go
// BEFORE (HS256 — symmetric, insecure for multi-service)
token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
return token.SignedString(input.Secret)  // []byte secret

// AFTER (RS256 — asymmetric)
token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
return token.SignedString(input.PrivateKey)  // *rsa.PrivateKey
```

Generate Gate 2 RSA key pair (add to `scripts/gen-certs.sh`):
```bash
openssl genrsa -out certs/gate2/private.pem 2048
openssl rsa -in certs/gate2/private.pem -pubout -out certs/gate2/public.pem
```

#### 3.2 Validate Incoming G1-JWT

Add middleware to `server.go` that verifies the Bearer token from Gate 1 before the token handler
runs.

**File:** `internal/middleware/gate1_auth.go`

```go
package middleware

import (
    "crypto/rsa"
    "fmt"
    "net/http"
    "strings"

    jwtlib "github.com/golang-jwt/jwt/v5"
)

func Gate1Auth(gate1PublicKey *rsa.PublicKey, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

        _, err := jwtlib.ParseWithClaims(tokenStr, &jwtlib.RegisteredClaims{},
            func(t *jwtlib.Token) (interface{}, error) {
                if _, ok := t.Method.(*jwtlib.SigningMethodRSA); !ok {
                    return nil, fmt.Errorf("unexpected signing method")
                }
                return gate1PublicKey, nil
            },
            jwtlib.WithAudience("shield-gate2"),
            jwtlib.WithIssuer("shield-gate1"),
        )
        if err != nil {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

#### 3.3 Fix Server Timeouts in `main.go`

```go
// Replace bare http.ListenAndServe with:
srv := &http.Server{
    Addr:         cfg.Addr,
    Handler:      h,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
}
log.Printf("gate-2 listening on %s", cfg.Addr)
if err := srv.ListenAndServe(); err != nil {
    log.Fatalf("server stopped: %v", err)
}
```

#### 3.4 Fail-Fast on Missing Secret

```go
// In config/config.go Load():
keyPath := os.Getenv("GATE2_JWT_PRIVATE_KEY_PATH")
if keyPath == "" {
    log.Fatal("GATE2_JWT_PRIVATE_KEY_PATH must be set — refusing to start with dev default")
}
```

### Updated Environment Variables (Gate 2)

| Variable                      | Default                   | Description                                     |
|-------------------------------|---------------------------|-------------------------------------------------|
| `GATE2_ADDR`                  | `:8080`                   | Listen address                                  |
| `GATE2_JWT_PRIVATE_KEY_PATH`  | —                         | **Required.** RS256 private key (PEM)           |
| `GATE2_JWT_TTL`               | `10m`                     | Token lifetime                                  |
| `GATE2_JWT_ISSUER`            | `shield-gate2`            | JWT `iss` claim                                 |
| `GATE2_JWT_AUDIENCE`          | `shield-gate3`            | JWT `aud` claim                                 |
| `GATE2_CLIENT_ID_HEADER`      | `x-client-dn`             | mTLS DN header from Envoy                       |
| `GATE2_GATE1_PUBLIC_KEY_PATH` | `certs/gate1/public.pem`  | Gate 1 RS256 public key for G1-JWT validation   |

---

## 4. Gate 3 — User Authenticity (FIDO2 WebAuthn)

### Purpose
Verifies the **human user** is the legitimate account owner via a hardware-bound credential — a
device biometric (Face ID, fingerprint) or a physical FIDO2 security key. This makes the login
"un-phishable" because the credential is cryptographically bound to the origin domain and hardware;
it cannot be stolen or replayed on a phishing site.

### Why WebAuthn Defeats Phishing
A WebAuthn credential is signed by the device's **Secure Enclave / TPM** and includes the **RP ID
(Relying Party ID)** in the signed payload. Even if a user is tricked into visiting
`yono-sbi.phishing.com`, the credential created for `yono.sbi.co.in` will **never authenticate**
against a different origin. Credential harvesting is mathematically impossible.

### API Contract

Gate 3 exposes four endpoints:

#### `POST /gate3/register/begin`
Starts FIDO2 registration (first-time setup).

**Request Headers:** `Authorization: Bearer <G2-JWT>`

**Request Body:**
```json
{
  "username": "user@sbi.example",
  "display_name": "Priya Sharma"
}
```

**Response `200 OK`:**
```json
{
  "publicKey": {
    "challenge": "<base64url random 32 bytes>",
    "rp": { "name": "SBI YONO", "id": "yono.sbi.co.in" },
    "user": {
      "id": "<base64url user handle>",
      "name": "user@sbi.example",
      "displayName": "Priya Sharma"
    },
    "pubKeyCredParams": [
      { "type": "public-key", "alg": -7 },
      { "type": "public-key", "alg": -257 }
    ],
    "authenticatorSelection": {
      "authenticatorAttachment": "platform",
      "userVerification": "required",
      "residentKey": "required"
    },
    "timeout": 60000,
    "attestation": "direct"
  }
}
```

#### `POST /gate3/register/finish`
Completes registration. Client sends the signed attestation object from the device.

**Request Body:**
```json
{
  "id": "<credential ID, base64url>",
  "rawId": "<same, base64url>",
  "response": {
    "clientDataJSON": "<base64url>",
    "attestationObject": "<base64url CBOR>"
  },
  "type": "public-key"
}
```

**Success Response `201 Created`:**
```json
{ "credential_id": "<base64url>", "registered": true }
```

#### `POST /gate3/authenticate/begin`
Starts FIDO2 authentication.

**Request Headers:** `Authorization: Bearer <G2-JWT>`

**Request Body:**
```json
{ "username": "user@sbi.example" }
```

**Response `200 OK`:**
```json
{
  "publicKey": {
    "challenge": "<base64url random 32 bytes>",
    "rpId": "yono.sbi.co.in",
    "allowCredentials": [
      { "type": "public-key", "id": "<credential ID, base64url>" }
    ],
    "userVerification": "required",
    "timeout": 60000
  }
}
```

#### `POST /gate3/authenticate/finish`
Completes authentication. Validates the signature from the device against the stored public key.

**Request Body:**
```json
{
  "id": "<credential ID>",
  "rawId": "<credential ID, base64url>",
  "response": {
    "clientDataJSON": "<base64url>",
    "authenticatorData": "<base64url>",
    "signature": "<base64url>",
    "userHandle": "<base64url>"
  },
  "type": "public-key"
}
```

**Success Response `200 OK`:**
```json
{
  "session_token": "<opaque session token>",
  "expires_in": 1800,
  "user_id": "<internal user identifier>"
}
```

### Go Implementation using `go-webauthn/webauthn`

**Dependency:**
```bash
go get github.com/go-webauthn/webauthn@latest
```

**File:** `backend/gate3/internal/handler/webauthn.go`

```go
package handler

import (
    "encoding/json"
    "net/http"

    "github.com/go-webauthn/webauthn/webauthn"
    "shield/gate3/internal/session"
    "shield/gate3/internal/userstore"
)

type WebAuthnHandler struct {
    WebAuthn  *webauthn.WebAuthn
    Sessions  session.Store    // Redis-backed challenge session store
    UserStore userstore.Store  // Postgres-backed credential store
}

func (h *WebAuthnHandler) RegisterBegin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username    string `json:"username"`
        DisplayName string `json:"display_name"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    user, err := h.UserStore.GetOrCreate(r.Context(), req.Username, req.DisplayName)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    creation, sessionData, err := h.WebAuthn.BeginRegistration(user)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    // Store challenge session in Redis (TTL 60s)
    h.Sessions.Save(r.Context(), "reg:"+req.Username, sessionData)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(creation)
}

func (h *WebAuthnHandler) RegisterFinish(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")
    user, _ := h.UserStore.Get(r.Context(), username)

    sessionData, err := h.Sessions.Load(r.Context(), "reg:"+username)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    credential, err := h.WebAuthn.FinishRegistration(user, *sessionData, r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    h.UserStore.AddCredential(r.Context(), username, credential)
    h.Sessions.Delete(r.Context(), "reg:"+username)

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]any{
        "credential_id": credential.ID,
        "registered":    true,
    })
}

func (h *WebAuthnHandler) AuthBegin(w http.ResponseWriter, r *http.Request) {
    var req struct{ Username string `json:"username"` }
    json.NewDecoder(r.Body).Decode(&req)

    user, err := h.UserStore.Get(r.Context(), req.Username)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    assertion, sessionData, err := h.WebAuthn.BeginLogin(user)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    h.Sessions.Save(r.Context(), "auth:"+req.Username, sessionData)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(assertion)
}

func (h *WebAuthnHandler) AuthFinish(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")
    user, _ := h.UserStore.Get(r.Context(), username)
    sessionData, _ := h.Sessions.Load(r.Context(), "auth:"+username)

    credential, err := h.WebAuthn.FinishLogin(user, *sessionData, r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    // Update sign counter in DB — prevents authenticator cloning attacks
    h.UserStore.UpdateCounter(r.Context(), username, credential.ID, credential.Authenticator.SignCount)
    h.Sessions.Delete(r.Context(), "auth:"+username)

    // Issue opaque session token stored in Redis
    sessionToken := session.IssueSessionToken(r.Context(), username)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "session_token": sessionToken,
        "expires_in":    1800,
        "user_id":       user.WebAuthnID(),
    })
}
```

### WebAuthn Configuration

```go
// In gate3/internal/config/webauthn.go
wconfig := &webauthn.Config{
    RPDisplayName: "SBI YONO",
    RPID:          "yono.sbi.co.in",
    RPOrigins: []string{
        "https://yono.sbi.co.in",
        "android:apk-key-hash:<SHA256-of-signing-cert>", // Android origin
    },
    AuthenticatorSelection: protocol.AuthenticatorSelection{
        AuthenticatorAttachment: protocol.Platform,       // Device biometric only
        UserVerification:        protocol.VerificationRequired,
        ResidentKey:             protocol.ResidentKeyRequirementRequired,
    },
    Timeout: 60000,
    Debug:   false,
}
```

### Credential Storage Schema (PostgreSQL)

```sql
-- Users table
CREATE TABLE webauthn_users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username     TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Credentials table (one user may have multiple registered devices)
CREATE TABLE webauthn_credentials (
    id             BYTEA PRIMARY KEY,         -- credential.ID (bytes)
    user_id        UUID REFERENCES webauthn_users(id) ON DELETE CASCADE,
    public_key     BYTEA NOT NULL,            -- credential.PublicKey (CBOR)
    sign_count     BIGINT NOT NULL DEFAULT 0, -- replay counter
    attestation    JSONB,                     -- raw attestation for audit
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    last_used_at   TIMESTAMPTZ
);

CREATE INDEX ON webauthn_credentials(user_id);
```

### Gate 3 Config / Environment Variables

| Variable                     | Default             | Description                                  |
|------------------------------|---------------------|----------------------------------------------|
| `GATE3_ADDR`                 | `:8082`             | Listen address                               |
| `GATE3_RP_ID`                | `yono.sbi.co.in`   | WebAuthn Relying Party ID                    |
| `GATE3_RP_ORIGIN`            | —                   | **Required.** Comma-separated allowed origins |
| `GATE3_GATE2_PUBLIC_KEY_PATH`| `certs/gate2/public.pem` | Gate 2 RS256 public key               |
| `GATE3_SESSION_STORE_ADDR`   | `redis:6379`        | Redis for WebAuthn challenge sessions        |
| `GATE3_DB_DSN`               | —                   | **Required.** PostgreSQL connection string   |
| `GATE3_SESSION_TOKEN_SECRET` | —                   | **Required.** Session token signing secret   |

---

## 5. Cross-Gate Token Chain

### Full Authentication Sequence Diagram

```
Client App          Gate 1              Gate 2 (Envoy+Go)    Gate 3 (Go+WebAuthn)
    │                  │                       │                      │
    │ 1. Generate nonce│                       │                      │
    │ 2. Request OS    │                       │                      │
    │    integrity tkn │                       │                      │
    │──POST /gate1/attest►                     │                      │
    │  {platform,token,│ 3. Verify with        │                      │
    │   nonce}         │    Google/Apple        │                      │
    │◄─ G1-JWT ────────│                       │                      │
    │                  │                       │                      │
    │ 4. mTLS connect (cert pinned)            │                      │
    │──POST /gate2/token ─────────────────────►│                      │
    │  Authorization: Bearer <G1-JWT>          │                      │
    │  [mTLS client cert verified by Envoy]    │                      │
    │                  │       5. Validate G1-JWT                     │
    │                  │       6. Extract DN from x-client-dn         │
    │◄─ G2-JWT ────────────────────────────────│                      │
    │                  │                       │                      │
    │──POST /gate3/authenticate/begin ────────────────────────────────►
    │  Authorization: Bearer <G2-JWT>          │                      │
    │                  │                       │    7. Validate G2-JWT│
    │                  │                       │    8. Return challenge│
    │◄─ WebAuthn Challenge ───────────────────────────────────────────│
    │                  │                       │                      │
    │ 9. User biometric / FIDO2 key signs challenge                   │
    │──POST /gate3/authenticate/finish ───────────────────────────────►
    │  { signed assertion }                    │                      │
    │                  │                       │   10. Verify sig     │
    │                  │                       │   11. Update counter │
    │◄─ Session Token ────────────────────────────────────────────────│
    │                  │                       │                      │
    │ ✅ Banking session established           │                      │
```

### Token Validation at Each Gate

| Gate   | Validates                                      | Issues                           |
|--------|------------------------------------------------|----------------------------------|
| Gate 1 | Play Integrity / App Attest from Google/Apple  | G1-JWT (RS256, aud=gate2)        |
| Gate 2 | mTLS client cert (Envoy) + G1-JWT              | G2-JWT (RS256, aud=gate3)        |
| Gate 3 | G2-JWT + WebAuthn assertion signature          | Opaque session token (Redis)     |

---

## 6. Shared Infrastructure

### Redis Usage

| Key Pattern       | TTL    | Content                  | Consumer              |
|-------------------|--------|--------------------------|-----------------------|
| `nonce:<hex>`     | 5 min  | `"1"` (existence flag)   | Gate 1 nonce dedup    |
| `reg:<username>`  | 60 s   | WebAuthn SessionData JSON | Gate 3 registration  |
| `auth:<username>` | 60 s   | WebAuthn SessionData JSON | Gate 3 authentication |
| `sess:<token>`    | 30 min | User identity JSON        | Session validation   |

### Certificate Hierarchy

```
Root CA  (self-signed, offline)
  ├── Gate 1 RSA Key Pair     (signs G1-JWT)
  │     certs/gate1/private.pem  /  certs/gate1/public.pem
  ├── Gate 2 RSA Key Pair     (signs G2-JWT)
  │     certs/gate2/private.pem  /  certs/gate2/public.pem
  └── mTLS Certificates       (for Envoy — already in certs/dev/)
        ca.crt / server.crt / server.key / client.crt / client.key
```

**Cross-gate public key distribution:**
- Gate 2 holds `certs/gate1/public.pem` → verifies incoming G1-JWTs.
- Gate 3 holds `certs/gate2/public.pem` → verifies incoming G2-JWTs.
- Private keys never leave their own service container.

### Docker Compose Services (add to `infrastructure/docker-compose.yml`)

```yaml
services:
  gate1:
    build: ../backend/gate1
    ports: ["8081:8081"]
    environment:
      - GATE1_PLAY_INTEGRITY_DECRYPTION_KEY=${PLAY_INTEGRITY_DECRYPTION_KEY}
      - GATE1_PLAY_INTEGRITY_VERIFICATION_KEY=${PLAY_INTEGRITY_VERIFICATION_KEY}
      - GATE1_JWT_PRIVATE_KEY_PATH=/certs/gate1/private.pem
      - GATE1_NONCE_STORE_ADDR=redis:6379
    volumes:
      - ../backend/three-gate-login/certs:/certs:ro
    depends_on: [redis]

  gate2:
    build: ../backend/three-gate-login
    ports: ["8080:8080"]
    environment:
      - GATE2_JWT_PRIVATE_KEY_PATH=/certs/gate2/private.pem
      - GATE2_GATE1_PUBLIC_KEY_PATH=/certs/gate1/public.pem
    volumes:
      - ../backend/three-gate-login/certs:/certs:ro

  gate2-envoy:
    image: envoyproxy/envoy:v1.29-latest
    ports: ["8443:8443", "9901:9901"]
    volumes:
      - ../backend/three-gate-login/deploy/envoy/envoy.yaml:/etc/envoy/envoy.yaml:ro
      - ../backend/three-gate-login/certs/dev:/certs/dev:ro
    depends_on: [gate2]

  gate3:
    build: ../backend/gate3
    ports: ["8082:8082"]
    environment:
      - GATE3_RP_ID=yono.sbi.co.in
      - GATE3_RP_ORIGIN=https://yono.sbi.co.in
      - GATE3_GATE2_PUBLIC_KEY_PATH=/certs/gate2/public.pem
      - GATE3_SESSION_STORE_ADDR=redis:6379
      - GATE3_DB_DSN=postgres://shield:${POSTGRES_PASSWORD}@postgres:5432/shield_webauthn
    volumes:
      - ../backend/three-gate-login/certs:/certs:ro
    depends_on: [redis, postgres]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=shield_webauthn
      - POSTGRES_USER=shield
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

---

## 7. Security Hardening Checklist

### Before Hackathon Demo

- [ ] Gate 2: Switch JWT signing from HS256 → RS256
- [ ] Gate 2: Add G1-JWT validation middleware (`middleware/gate1_auth.go`)
- [ ] Gate 2: Add server `ReadTimeout` / `WriteTimeout` / `IdleTimeout`
- [ ] Gate 2: Fail-fast if private key env var is missing (no silent dev default)
- [ ] Gate 1: Implement Play Integrity endpoint (Android path minimum)
- [ ] Gate 3: Implement WebAuthn register + authenticate endpoints
- [ ] All: Nonce / replay protection via Redis
- [ ] All: Rate limiting (10 req/min per IP on attestation endpoints)
- [ ] All: TLS 1.3 minimum enforced at Envoy

### Before Production (Post-Hackathon)

- [ ] Migrate JWT private keys to Google Cloud KMS or HashiCorp Vault
- [ ] Replace Redis with Redis Cluster + encryption at rest
- [ ] WebAuthn: alert on sign counter regression (authenticator cloning detection)
- [ ] Forward all auth events to Kafka topic `shield.auth.events` for SIEM
- [ ] Account lockout after N consecutive Gate 1/3 failures
- [ ] Gate 1: Add iOS App Attest assertion (continuous re-attestation) endpoint
- [ ] mTLS: Add OCSP stapling for client certificate revocation checking
- [ ] Penetration test: replay, credential stuffing, downgrade attack scenarios

---

## 8. Recommended Folder Structure

```
backend/
├── three-gate-login/              ← Gate 2 (current repo) — mTLS + JWT
│   ├── cmd/gate2/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/
│   │   │   ├── health.go
│   │   │   └── token.go
│   │   ├── jwt/
│   │   │   ├── issue.go           ← upgrade to RS256
│   │   │   └── verify.go
│   │   ├── middleware/
│   │   │   ├── gate1_auth.go      ← NEW: validate incoming G1-JWT
│   │   │   ├── logging.go
│   │   │   └── request_id.go
│   │   ├── model/
│   │   ├── mtls/
│   │   └── server/
│   ├── certs/
│   │   ├── dev/                   ← mTLS certs (existing)
│   │   ├── gate1/                 ← NEW: gate1_public.pem
│   │   └── gate2/                 ← NEW: gate2_private.pem, gate2_public.pem
│   ├── deploy/envoy/
│   └── backend.md                 ← this file
│
├── gate1/                         ← NEW: App Attestation service
│   ├── cmd/gate1/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/attest.go
│   │   ├── integrity/             ← Play Integrity API client
│   │   ├── appattest/             ← Apple App Attest client
│   │   ├── jwt/                   ← RS256 issue only
│   │   ├── nonce/                 ← Redis nonce store
│   │   └── server/
│   ├── Dockerfile
│   └── go.mod
│
└── gate3/                         ← NEW: WebAuthn FIDO2 service
    ├── cmd/gate3/main.go
    ├── internal/
    │   ├── config/
    │   ├── handler/
    │   │   └── webauthn.go
    │   ├── middleware/
    │   │   └── gate2_auth.go      ← validate incoming G2-JWT
    │   ├── session/               ← Redis WebAuthn session store
    │   ├── userstore/             ← Postgres credential store
    │   └── server/
    ├── migrations/
    │   └── 001_create_webauthn_tables.sql
    ├── Dockerfile
    └── go.mod
```

---

*Last updated: 2026-06-13 | Author: SHIELD Team — PhishKillers, IIT Jodhpur*
