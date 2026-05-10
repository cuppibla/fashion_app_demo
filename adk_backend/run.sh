#!/bin/bash
set -e

cd "$(dirname "$0")"

go run . web \
  --write-timeout=300s \
  --read-timeout=60s \
  --idle-timeout=120s \
  api \
  --sse-write-timeout=300s \
  --webui_address=http://localhost:61982
