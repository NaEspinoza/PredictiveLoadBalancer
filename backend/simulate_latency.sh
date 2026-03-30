#!/usr/bin/env sh
# Simple helper to curl the backend repeatedly to observe latency
URL=${1:-http://localhost:9001/}
INTERVAL=${2:-1}

while true; do
  date
  curl -w "\n" -s "$URL" || echo "request failed"
  sleep "$INTERVAL"
done
