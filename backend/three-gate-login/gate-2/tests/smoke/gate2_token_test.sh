#!/usr/bin/env bash
set -euo pipefail

curl -sS https://localhost:8443/gate2/token \
  --cacert certs/dev/ca.crt \
  --cert certs/dev/client.crt \
  --key certs/dev/client.key | cat

