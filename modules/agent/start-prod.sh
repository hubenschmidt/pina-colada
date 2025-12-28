#!/usr/bin/env bash
set -euo pipefail

exec 2>&1
set -x

echo "========================================="
echo "Starting Go agent in production mode"
echo "========================================="
echo "Go version: $(go version 2>/dev/null || echo 'compiled binary')"
echo "Working directory: $(pwd)"
echo "Environment variables:"
env | grep -E '^(PORT|DATABASE_URL|ENV)' || true
echo "========================================="

_term() {
  echo "Caught SIGTERM, shutting down..."
  kill -TERM "$AGENT_PID" 2>/dev/null || true
}
trap _term TERM INT

PORT=${PORT:-8080}
echo "PORT set to: $PORT"

echo "Running migrations..."
/migrate up || echo "Could not run migrations"

echo "Starting Go agent on port $PORT..."
/agent-go &
AGENT_PID=$!
echo "Agent PID: $AGENT_PID"

wait $AGENT_PID
EXIT_CODE=$?
echo "Process exited with code: $EXIT_CODE"
exit $EXIT_CODE
