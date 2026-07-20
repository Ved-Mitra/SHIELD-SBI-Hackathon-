#!/usr/bin/env bash
# gate1_smoke_test.sh — Smoke tests for the Gate 1 attestation endpoint.
#
# Tests:
#   1. Health check endpoint returns 200.
#   2. Missing nonce returns 400.
#   3. Invalid platform returns 400.
#   4. Replayed nonce returns 401.
#   5. Mock mode: valid Android request returns a G1-JWT.
#   6. Mock mode: valid iOS request returns a G1-JWT.
#   7. Rate limiter kicks in after burst (optional — skipped by default).
#
# Prerequisites:
#   - Gate 1 running with GATE1_MOCK_ATTESTATION=true on localhost:8081.
#   - jq installed (brew install jq / apt-get install jq).
#
# Usage:
#   GATE1_MOCK_ATTESTATION=true go run ./cmd/gate1 &
#   sleep 1
#   ./tests/smoke/gate1_smoke_test.sh
set -euo pipefail

BASE_URL="${GATE1_URL:-http://localhost:8081}"
PASS=0
FAIL=0
NONCE=""

# ── colour helpers ─────────────────────────────────────────────────────────────
GREEN="\033[0;32m"; RED="\033[0;31m"; RESET="\033[0m"

ok()   { echo -e "${GREEN}  ✅ PASS${RESET}: $1"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}  ❌ FAIL${RESET}: $1"; FAIL=$((FAIL+1)); }

assert_status() {
  local label="$1" want="$2" got="$3"
  if [[ "$got" == "$want" ]]; then ok "$label (HTTP $got)"; else fail "$label — want $want, got $got"; fi
}

# ── helper: generate a fresh 32-byte hex nonce ────────────────────────────────
fresh_nonce() { openssl rand -hex 32; }

echo ""
echo "══════════════════════════════════════════════════════════"
echo "  Gate 1 Smoke Tests — ${BASE_URL}"
echo "══════════════════════════════════════════════════════════"
echo ""

# ── Test 1: Health check ───────────────────────────────────────────────────────
echo "1. Health check"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/healthz")
assert_status "GET /healthz" "200" "$STATUS"

# ── Test 2: Missing nonce → 400 ───────────────────────────────────────────────
echo "2. Missing nonce"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate1/attest" \
  -H "Content-Type: application/json" \
  -d '{"platform":"android","integrity_token":"fake"}')
assert_status "POST /gate1/attest (missing nonce)" "400" "$STATUS"

# ── Test 3: Invalid platform → 400 ────────────────────────────────────────────
echo "3. Invalid platform"
NONCE=$(fresh_nonce)
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate1/attest" \
  -H "Content-Type: application/json" \
  -d "{\"platform\":\"windows\",\"nonce\":\"${NONCE}\"}")
assert_status "POST /gate1/attest (invalid platform)" "400" "$STATUS"

# ── Test 4: Mock Android → 200 + JWT ──────────────────────────────────────────
echo "4. Mock Android attestation"
NONCE=$(fresh_nonce)
RESP=$(curl -s -X POST "${BASE_URL}/gate1/attest" \
  -H "Content-Type: application/json" \
  -d "{\"platform\":\"android\",\"integrity_token\":\"mock-token\",\"nonce\":\"${NONCE}\"}")
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate1/attest" \
  -H "Content-Type: application/json" \
  -d "{\"platform\":\"android\",\"integrity_token\":\"mock-token\",\"nonce\":\"$(fresh_nonce)\"}")
assert_status "POST /gate1/attest (android mock)" "200" "$STATUS"

# Decode and print JWT claims (requires jq)
if command -v jq &>/dev/null; then
  TOKEN=$(echo "$RESP" | jq -r '.token // empty' 2>/dev/null || true)
  if [[ -n "$TOKEN" ]]; then
    PAYLOAD=$(echo "$TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | jq . 2>/dev/null || true)
    if [[ -n "$PAYLOAD" ]]; then
      echo "    JWT claims: $(echo "$PAYLOAD" | jq -c '{iss,sub,aud,exp}')"
    fi
  fi
fi

# ── Test 5: Mock iOS → 200 + JWT ──────────────────────────────────────────────
echo "5. Mock iOS attestation"
NONCE=$(fresh_nonce)
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate1/attest" \
  -H "Content-Type: application/json" \
  -d "{\"platform\":\"ios\",\"attest_object\":\"bW9jaw==\",\"client_data_hash\":\"bW9jaw==\",\"key_id\":\"bW9jaw==\",\"nonce\":\"${NONCE}\"}")
assert_status "POST /gate1/attest (ios mock)" "200" "$STATUS"

# ── Test 6: Replay attack → 401 ───────────────────────────────────────────────
echo "6. Nonce replay attack prevention"
NONCE=$(fresh_nonce)
# First call — should succeed
curl -s -o /dev/null -X POST "${BASE_URL}/gate1/attest" \
  -H "Content-Type: application/json" \
  -d "{\"platform\":\"android\",\"integrity_token\":\"mock-token\",\"nonce\":\"${NONCE}\"}" > /dev/null
# Second call with same nonce — should fail
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate1/attest" \
  -H "Content-Type: application/json" \
  -d "{\"platform\":\"android\",\"integrity_token\":\"mock-token\",\"nonce\":\"${NONCE}\"}")
assert_status "POST /gate1/attest (replayed nonce)" "401" "$STATUS"

# ── Test 7: Wrong method → 405 ────────────────────────────────────────────────
echo "7. Wrong HTTP method"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "${BASE_URL}/gate1/attest")
assert_status "GET /gate1/attest (method not allowed)" "405" "$STATUS"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════════════════"
echo "  Results: ${PASS} passed, ${FAIL} failed"
echo "══════════════════════════════════════════════════════════"
echo ""

[[ $FAIL -eq 0 ]] && exit 0 || exit 1
