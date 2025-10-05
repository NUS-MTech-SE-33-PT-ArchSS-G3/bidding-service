#!/usr/bin/env sh

set -euo pipefail

REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASS="${REDIS_PASSWORD:-}"

auth() {
  if [ -n "$REDIS_PASS" ]; then
    printf " -a %s" "$REDIS_PASS"
  fi
}

redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" $(auth) SET \
  auction:a_seeded \
  '{"auctionId":"a_seeded","status":"open","endsAt":"2025-12-31T23:59:59Z","startingPrice":100.0,"currentPrice":120.0,"minIncrement":10.0,"version":2}'

redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" $(auth) SET \
  auction:a_demo \
  '{"auctionId":"a_demo","status":"open","endsAt":"2025-12-31T23:59:59Z","startingPrice":50.0,"currentPrice":0.0,"minIncrement":5.0,"version":0}'

echo "Redis seed done"