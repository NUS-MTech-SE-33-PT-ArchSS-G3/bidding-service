#!/usr/bin/env sh

set -euo pipefail

BROKERS="${KAFKA_BROKERS:-kafka:9092}"

echo '{"auctionId":"a_evt_1","startingPrice":70.0,"minIncrement":7.0,"endsAt":"2025-12-31T23:59:59Z","version":0}' \
  | kcat -b "$BROKERS" -t auction.opened -P

echo "Kafka seed done"