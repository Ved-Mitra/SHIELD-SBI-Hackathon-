# mTLS Local Dev

## Generate certs
Use the provided script to create a dev CA, server cert, and client cert.

```bash
./scripts/gen-certs.sh
```

This creates files in `certs/dev/`:
- `ca.crt`
- `server.crt`, `server.key`
- `client.crt`, `client.key`

## Run Gate-2 app
```bash
go run ./cmd/gate2
```

## Run Envoy
```bash
envoy -c deploy/envoy/envoy.yaml
```

## Call Gate-2 using mTLS
```bash
curl -sS https://localhost:8443/gate2/token \
  --cacert certs/dev/ca.crt \
  --cert certs/dev/client.crt \
  --key certs/dev/client.key
```

