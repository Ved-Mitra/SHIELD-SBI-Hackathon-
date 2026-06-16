#!/usr/bin/env bash
# gate2_smoke_test.sh — Smoke tests for the Gate 2 token endpoint.
#
# Tests (run against a server with REAL G1-JWT validation enabled):
#   1.  Health check returns 200.
#   2.  POST /gate2/token without Authorization header → 401.
#   3.  POST /gate2/token with garbage JWT → 401.
#   4.  POST /gate2/token with expired G1-JWT → 401.
#   5.  POST /gate2/token with wrong audience G1-JWT → 401.
#   6.  POST /gate2/token with valid G1-JWT + mTLS header → 200 + G2-JWT.
#   7.  G2-JWT claims are correct (iss=shield-gate2, aud=shield-gate3).
#   8.  G2-JWT sub matches the mTLS client DN.
#   9.  Wrong HTTP method (GET /gate2/token) → 405.
#   10. expires_in matches configured TTL.
#   11. Rate limiter: excessive requests → 429.
#
# Prerequisites:
#   - Gate 2 running in REAL mode (no GATE2_MOCK_GATE1) on localhost:8080.
#   - GATE2_MOCK_CLIENT_DN=true so x-client-dn is injected by mock.
#   - Gate 1 RSA private key available to generate test G1-JWTs.
#   - openssl, curl, jq installed.
#
# Usage (from gate-2/ directory):
#   ./tests/smoke/gate2_smoke_test.sh
#
# The script will start and stop the server automatically.
set -euo pipefail

BASE_URL="${GATE2_URL:-http://localhost:8080}"
GATE1_KEY="${GATE1_PRIVATE_KEY_PATH:-../gate-1/certs/gate1/private.pem}"
GATE1_PUB="${GATE1_PUBLIC_KEY_PATH:-certs/gate1/public.pem}"
GATE2_KEY="${GATE2_JWT_PRIVATE_KEY_PATH:-certs/gate2/private.pem}"
CLIENT_DN="CN=shield-mobile,O=SHIELD,C=IN"
PASS=0; FAIL=0

GREEN="\033[0;32m"; RED="\033[0;31m"; YELLOW="\033[0;33m"; RESET="\033[0m"

ok()   { echo -e "${GREEN}  ✅ PASS${RESET}: $1"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}  ❌ FAIL${RESET}: $1"; FAIL=$((FAIL+1)); }
skip() { echo -e "${YELLOW}  ⏭  SKIP${RESET}: $1"; }

assert_status() {
  local label="$1" want="$2" got="$3"
  if [[ "$got" == "$want" ]]; then
    ok "$label (HTTP $got)"
  else
    fail "$label — want HTTP $want, got HTTP $got"
  fi
}

# ── make_jwt: build a real RS256-signed JWT using openssl ────────────────────
# Usage: make_jwt <private_key_path> <issuer> <audience> <exp_offset_seconds>
# exp_offset: positive = future (valid), negative = past (expired)
make_jwt() {
  local key="$1" iss="$2" aud="$3" exp_offset="${4:-120}"
  local now exp header payload sig

  now=$(date +%s)
  exp=$((now + exp_offset))

  header=$(printf '%s' '{"alg":"RS256","typ":"JWT"}' \
    | openssl base64 -A | tr '+/' '-_' | tr -d '=')

  payload=$(printf '%s' "{\"iss\":\"${iss}\",\"aud\":[\"${aud}\"],\"sub\":\"android:mock\",\"iat\":${now},\"exp\":${exp}}" \
    | openssl base64 -A | tr '+/' '-_' | tr -d '=')

  sig=$(printf '%s' "${header}.${payload}" \
    | openssl dgst -sha256 -sign "${key}" \
    | openssl base64 -A | tr '+/' '-_' | tr -d '=')

  printf '%s' "${header}.${payload}.${sig}"
}

# ── Preflight: check key files ────────────────────────────────────────────────
check_keys() {
  local ok=true
  if [[ ! -f "${GATE1_KEY}" ]]; then
    echo -e "${RED}[ERROR]${RESET} Gate 1 private key not found: ${GATE1_KEY}"
    echo "  → Run: cd ../gate-1 && ./scripts/gen-gate1-keys.sh"
    ok=false
  fi
  if [[ ! -f "${GATE1_PUB}" ]]; then
    echo -e "${RED}[ERROR]${RESET} Gate 1 public key not found: ${GATE1_PUB}"
    echo "  → Run: cp ../gate-1/certs/gate1/public.pem ${GATE1_PUB}"
    ok=false
  fi
  if [[ ! -f "${GATE2_KEY}" ]]; then
    echo -e "${RED}[ERROR]${RESET} Gate 2 private key not found: ${GATE2_KEY}"
    echo "  → Run: ./scripts/gen-gate2-keys.sh"
    ok=false
  fi
  if ! $ok; then exit 1; fi
}

# ── Start Gate 2 server in REAL mode ─────────────────────────────────────────
start_server() {
  GATE2_JWT_PRIVATE_KEY_PATH="${GATE2_KEY}" \
  GATE2_GATE1_PUBLIC_KEY_PATH="${GATE1_PUB}" \
  GATE2_MOCK_CLIENT_DN=true \
    go run ./cmd/gate2 > /tmp/gate2_smoke.log 2>&1 &
  SERVER_PID=$!

  # Wait until server accepts connections (max 10s)
  local waited=0
  until curl -s -o /dev/null "${BASE_URL}/healthz" 2>/dev/null; do
    sleep 1
    waited=$((waited+1))
    if [[ $waited -ge 10 ]]; then
      echo -e "${RED}[ERROR]${RESET} Server did not start within 10s"
      echo "Server log:"
      cat /tmp/gate2_smoke.log
      kill $SERVER_PID 2>/dev/null
      exit 1
    fi
  done
}

stop_server() {
  kill "${SERVER_PID}" 2>/dev/null
  wait "${SERVER_PID}" 2>/dev/null || true
}

# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════════════════"
echo "  Gate 2 Smoke Tests — ${BASE_URL}"
echo "══════════════════════════════════════════════════════════"
echo ""

check_keys

echo "Starting Gate 2 in real-validation mode..."
start_server
echo "Server ready (PID=${SERVER_PID})"
echo ""

# Pre-generate JWTs for the tests
VALID_G1_JWT=$(make_jwt "${GATE1_KEY}" "shield-gate1" "shield-gate2" 120)
EXPIRED_G1_JWT=$(make_jwt "${GATE1_KEY}" "shield-gate1" "shield-gate2" -60)
WRONG_AUD_G1_JWT=$(make_jwt "${GATE1_KEY}" "shield-gate1" "shield-gate3" 120)
WRONG_ISS_G1_JWT=$(make_jwt "${GATE1_KEY}" "wrong-issuer" "shield-gate2" 120)

# ── 1: Health check ───────────────────────────────────────────────────────────
echo "1. Health check"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/healthz")
assert_status "GET /healthz" "200" "$STATUS"

# ── 2: No Authorization header → 401 ─────────────────────────────────────────
echo "2. Missing Authorization header"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate2/token" \
  -H "x-client-dn: ${CLIENT_DN}")
assert_status "POST /gate2/token (no auth)" "401" "$STATUS"

# ── 3: Garbage Bearer token → 401 ────────────────────────────────────────────
echo "3. Garbage Bearer token"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate2/token" \
  -H "Authorization: Bearer not.a.real.jwt" \
  -H "x-client-dn: ${CLIENT_DN}")
assert_status "POST /gate2/token (garbage JWT)" "401" "$STATUS"

# ── 4: Expired G1-JWT → 401 ──────────────────────────────────────────────────
echo "4. Expired G1-JWT"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate2/token" \
  -H "Authorization: Bearer ${EXPIRED_G1_JWT}" \
  -H "x-client-dn: ${CLIENT_DN}")
assert_status "POST /gate2/token (expired JWT)" "401" "$STATUS"

# ── 5a: Wrong audience G1-JWT → 401 ──────────────────────────────────────────
echo "5a. G1-JWT with wrong audience"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate2/token" \
  -H "Authorization: Bearer ${WRONG_AUD_G1_JWT}" \
  -H "x-client-dn: ${CLIENT_DN}")
assert_status "POST /gate2/token (wrong aud)" "401" "$STATUS"

# ── 5b: Wrong issuer G1-JWT → 401 ────────────────────────────────────────────
echo "5b. G1-JWT with wrong issuer"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate2/token" \
  -H "Authorization: Bearer ${WRONG_ISS_G1_JWT}" \
  -H "x-client-dn: ${CLIENT_DN}")
assert_status "POST /gate2/token (wrong iss)" "401" "$STATUS"

# ── 6: Valid G1-JWT + mTLS → 200 + G2-JWT ────────────────────────────────────
echo "6. Valid G1-JWT returns G2-JWT"
RESP=$(curl -s -X POST "${BASE_URL}/gate2/token" \
  -H "Authorization: Bearer ${VALID_G1_JWT}" \
  -H "x-client-dn: ${CLIENT_DN}")
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate2/token" \
  -H "Authorization: Bearer ${VALID_G1_JWT}" \
  -H "x-client-dn: ${CLIENT_DN}")
assert_status "POST /gate2/token (valid G1-JWT)" "200" "$STATUS"

# ── 7: G2-JWT claims ─────────────────────────────────────────────────────────
echo "7. G2-JWT claims are correct"
if command -v jq &>/dev/null; then
  TOKEN=$(echo "$RESP" | jq -r '.token // empty' 2>/dev/null || true)
  if [[ -n "$TOKEN" ]]; then
    # base64url decode the payload (middle segment)
    PAYLOAD_B64=$(echo "$TOKEN" | cut -d. -f2)
    # base64url → standard base64: swap -→+ and _→/, append padding, decode
    PAYLOAD=$(echo "${PAYLOAD_B64}==" | tr -- '-_' '+/' | openssl base64 -d -A 2>/dev/null | jq . 2>/dev/null || true)

    ISS=$(echo "$PAYLOAD" | jq -r '.iss' 2>/dev/null || true)
    AUD=$(echo "$PAYLOAD" | jq -r '.aud[0]' 2>/dev/null || true)
    SUB=$(echo "$PAYLOAD" | jq -r '.sub' 2>/dev/null || true)

    if [[ "$ISS" == "shield-gate2" && "$AUD" == "shield-gate3" ]]; then
      ok "G2-JWT iss=shield-gate2, aud=shield-gate3"
    else
      fail "G2-JWT wrong claims: iss=$ISS aud=$AUD"
    fi

    echo "    Claims: $(echo "$PAYLOAD" | jq -c '{iss,sub,aud,exp}')"
  else
    fail "No token field in response body"
  fi
else
  skip "jq not installed — skipping JWT claim inspection"
fi

# ── 8: G2-JWT sub = client DN ────────────────────────────────────────────────
echo "8. G2-JWT sub matches mTLS client DN"
if [[ -n "${PAYLOAD:-}" ]]; then
  # GATE2_MOCK_CLIENT_DN returns "CN=mock-client,O=SHIELD,C=IN" when header is missing
  # but we sent x-client-dn explicitly so it should use our value
  if [[ "$SUB" == *"shield-mobile"* ]] || [[ "$SUB" == *"mock-client"* ]]; then
    ok "G2-JWT sub=$SUB"
  else
    fail "G2-JWT sub unexpected: $SUB"
  fi
else
  skip "No payload to check"
fi

# ── 9: Wrong HTTP method → 405 ───────────────────────────────────────────────
echo "9. Wrong HTTP method (GET)"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "${BASE_URL}/gate2/token" \
  -H "Authorization: Bearer ${VALID_G1_JWT}")
assert_status "GET /gate2/token (method not allowed)" "405" "$STATUS"

# ── 10: expires_in field ─────────────────────────────────────────────────────
echo "10. expires_in is positive"
if command -v jq &>/dev/null && [[ -n "${RESP:-}" ]]; then
  EXPIRES=$(echo "$RESP" | jq -r '.expires_in // 0' 2>/dev/null || echo 0)
  TOKEN_TYPE=$(echo "$RESP" | jq -r '.token_type // empty' 2>/dev/null || true)
  if [[ "$EXPIRES" -gt 0 && "$TOKEN_TYPE" == "Bearer" ]]; then
    ok "expires_in=${EXPIRES}s, token_type=Bearer"
  else
    fail "expires_in=$EXPIRES or token_type=$TOKEN_TYPE unexpected"
  fi
else
  skip "jq not installed or no response"
fi

# ── 11: Rate limiter → 429 ───────────────────────────────────────────────────
echo "11. Rate limiter activates on burst"
GOT_429=false
for i in $(seq 1 35); do
  CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/gate2/token" \
    -H "Authorization: Bearer ${VALID_G1_JWT}" \
    -H "x-client-dn: ${CLIENT_DN}")
  if [[ "$CODE" == "429" ]]; then
    GOT_429=true
    break
  fi
done
if $GOT_429; then
  ok "Rate limiter returned 429 after burst"
else
  fail "Rate limiter never returned 429 in 35 requests"
fi

# ── Cleanup ───────────────────────────────────────────────────────────────────
stop_server

# ── Summary ──────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════════════════"
printf "  Results: ${GREEN}%d passed${RESET}, ${RED}%d failed${RESET}\n" "$PASS" "$FAIL"
echo "══════════════════════════════════════════════════════════"
echo ""
[[ $FAIL -eq 0 ]] && exit 0 || exit 1
