# Gate-2 Architecture (prototype)

## Boundary
- Envoy terminates TLS and validates client certs.
- The Go app trusts Envoy headers for client identity.

## Flow
1. Client connects to Envoy with mTLS.
2. Envoy validates certs using Gate-2 CA.
3. Envoy forwards request to Go app.
4. Go app issues a short-lived JWT.

## Identity
- Primary identity: `x-client-dn` from Envoy.
- Gate-2 uses this as JWT subject.

## JWT
- Issuer: `shield-gate2`.
- Audience: `shield-gate3`.
- TTL: 5-10 minutes.

