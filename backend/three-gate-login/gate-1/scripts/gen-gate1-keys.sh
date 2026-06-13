#!/usr/bin/env bash
# gen-gate1-keys.sh — Generate RSA key pair for Gate 1 JWT signing (RS256).
#
# Usage:
#   ./scripts/gen-gate1-keys.sh
#
# Output:
#   certs/gate1/private.pem  — RSA-2048 private key  (KEEP SECRET — never commit)
#   certs/gate1/public.pem   — RSA-2048 public key   (distribute to Gate 2)
#
# Gate 2 needs certs/gate1/public.pem to verify incoming G1-JWTs.
# Copy it to:  backend/three-gate-login/certs/gate1/public.pem
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
CERT_DIR="${ROOT_DIR}/certs/gate1"

mkdir -p "${CERT_DIR}"

echo "── Generating Gate 1 RSA-2048 key pair ─────────────────────────────"

# Generate private key (PKCS#8 format — directly readable by Go's x509.ParsePKCS8PrivateKey)
openssl genpkey -algorithm RSA \
  -pkeyopt rsa_keygen_bits:2048 \
  -out "${CERT_DIR}/private.pem"

# Derive public key from private key
openssl pkey \
  -in  "${CERT_DIR}/private.pem" \
  -pubout \
  -out "${CERT_DIR}/public.pem"

# Set restrictive permissions on private key
chmod 600 "${CERT_DIR}/private.pem"

echo ""
echo "✅  Keys generated:"
echo "    Private key : ${CERT_DIR}/private.pem  (chmod 600 — DO NOT COMMIT)"
echo "    Public  key : ${CERT_DIR}/public.pem"
echo ""
echo "📋  Next steps:"
echo "    1. Set GATE1_JWT_PRIVATE_KEY_PATH=${CERT_DIR}/private.pem"
echo "    2. Copy public key to Gate 2:"
echo "       cp ${CERT_DIR}/public.pem \\"
echo "          ../../three-gate-login/certs/gate1/public.pem"
echo "    3. Set GATE2_GATE1_PUBLIC_KEY_PATH in Gate 2 config."
echo ""

# Print key fingerprint for verification
echo "── Key fingerprint (SHA-256) ────────────────────────────────────────"
openssl pkey -in "${CERT_DIR}/private.pem" -pubout 2>/dev/null \
  | openssl dgst -sha256 -hex \
  | awk '{print "    " $2}'
echo ""
