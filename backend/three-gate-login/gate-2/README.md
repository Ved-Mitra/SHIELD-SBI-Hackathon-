# three-gate-login (Gate-2 mTLS)

## Scope (prototype)
- Provide Gate-2 API under SHIELD domain only.
- Enforce mTLS at Envoy, pass verified client identity to app.
- Issue short-lived JWT from Gate-2 after successful mTLS.
- Keep config and docs minimal, but runnable locally.

## Trust boundary and flow
1. Client connects to Envoy with mTLS.
2. Envoy verifies client cert against Gate-2 CA.
3. Envoy forwards request to Go app with verified client identity headers.
4. Go app issues short-lived JWT bound to client identity.

## API (minimal)
- `POST /gate2/token`
  - Input: none (identity from mTLS header).
  - Output: JWT + `expires_in`.
- `GET /healthz`

## JWT rules (prototype)
- Issuer: `shield-gate2`.
- Subject: verified client identity from mTLS.
- TTL: 5-10 minutes.
- Audience: `shield-gate3`.
- Rotation: daily secret during prototype; move to KMS later.

## Envoy termination
- In-repo config: `deploy/envoy/envoy.yaml`.
- Ops notes: `docs/envoy-notes.md`.

## Local dev (minimal)
1. Generate dev certs: `./scripts/gen-certs.sh`.
2. Start the Go app: `go run ./cmd/gate2`.
3. Start Envoy: `envoy -c deploy/envoy/envoy.yaml`.
4. Call Gate-2:

```bash
curl -sS https://localhost:8443/gate2/token \
  --cacert certs/dev/ca.crt \
  --cert certs/dev/client.crt \
  --key certs/dev/client.key
```

## Config (env)
- `GATE2_ADDR` (default `:8080`)
- `GATE2_JWT_SECRET` (default `dev-secret-change`)
- `GATE2_JWT_TTL` (default `10m`)
- `GATE2_JWT_ISSUER` (default `shield-gate2`)
- `GATE2_JWT_AUDIENCE` (default `shield-gate3`)
- `GATE2_CLIENT_ID_HEADER` (default `x-client-dn`)

## Folder structure
```
three-gate-login/
  README.md
  go.mod
  cmd/
    gate2/
      main.go
  internal/
    config/
    handler/
    jwt/
    middleware/
    model/
    mtls/
    server/
  certs/
    dev/
  deploy/
    envoy/
      envoy.yaml
  docs/
    architecture.md
    envoy-notes.md
    mtls-dev.md
    references.md
  scripts/
    gen-certs.sh
  tests/
    smoke/
      gate2_token_test.sh
```

