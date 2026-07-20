# Envoy Notes (prototype)

## Responsibilities
- Enforce mTLS between client and Envoy.
- Pass verified client identity to Gate-2.
- Restrict TLS versions and ciphers in production.

## Rotation
- Rotate CA and client certs by environment.
- Re-issue client certs on compromise or expiry.

## Logs
- Use Envoy access logs for request auditing.
- Tie Envoy request IDs to Gate-2 logs.

