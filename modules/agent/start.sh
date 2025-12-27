#!/usr/bin/env bash
set -euo pipefail

_term() {
  echo "Caught SIGTERM, shutting down..."
  kill -TERM "$AIR_PID" 2>/dev/null || true
}
trap _term TERM INT

echo "Running migrations..."
go run ./cmd/migrate up || echo "Could not run migrations"

echo "Running seeders..."
go run ./cmd/seed || echo "Could not run seeders"

echo "Uploading seed documents..."
go run ./cmd/seed-documents || echo "Could not upload seed documents"

echo "Starting Go agent with hot-reload..."
air -c .air.toml &
AIR_PID=$!

wait $AIR_PID
exit $?
