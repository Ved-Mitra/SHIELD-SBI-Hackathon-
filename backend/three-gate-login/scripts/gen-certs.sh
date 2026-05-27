#!/usr/bin/env bash
set -euo pipefail

CERT_DIR="certs/dev"
mkdir -p "${CERT_DIR}"

openssl genrsa -out "${CERT_DIR}/ca.key" 4096
openssl req -x509 -new -nodes -key "${CERT_DIR}/ca.key" -sha256 -days 365 \
  -subj "/C=IN/ST=MH/L=Mumbai/O=SHIELD/OU=Gate2/CN=shield-dev-ca" \
  -out "${CERT_DIR}/ca.crt"

openssl genrsa -out "${CERT_DIR}/server.key" 2048
openssl req -new -key "${CERT_DIR}/server.key" \
  -subj "/C=IN/ST=MH/L=Mumbai/O=SHIELD/OU=Gate2/CN=shield-gate2.local" \
  -out "${CERT_DIR}/server.csr"

openssl x509 -req -in "${CERT_DIR}/server.csr" -CA "${CERT_DIR}/ca.crt" -CAkey "${CERT_DIR}/ca.key" \
  -CAcreateserial -out "${CERT_DIR}/server.crt" -days 365 -sha256

openssl genrsa -out "${CERT_DIR}/client.key" 2048
openssl req -new -key "${CERT_DIR}/client.key" \
  -subj "/C=IN/ST=MH/L=Mumbai/O=SHIELD/OU=Client/CN=shield-mobile" \
  -out "${CERT_DIR}/client.csr"

openssl x509 -req -in "${CERT_DIR}/client.csr" -CA "${CERT_DIR}/ca.crt" -CAkey "${CERT_DIR}/ca.key" \
  -CAcreateserial -out "${CERT_DIR}/client.crt" -days 365 -sha256

echo "Generated certs in ${CERT_DIR}"

