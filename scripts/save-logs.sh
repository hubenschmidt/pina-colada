#!/bin/bash
# Save current docker compose logs to daily log files

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_ROOT/logs"
DATE_SUFFIX=$(date +%Y-%m-%d)

# Create logs directory
mkdir -p "$LOG_DIR"

cd "$PROJECT_ROOT"

# Save all logs combined
ALL_LOGS="$LOG_DIR/docker-compose-$DATE_SUFFIX.log"
echo "Saving all logs to: $ALL_LOGS"
docker compose logs --no-color > "$ALL_LOGS" 2>&1

# Save individual service logs
CLIENT_LOGS="$LOG_DIR/client-$DATE_SUFFIX.log"
echo "Saving client logs to: $CLIENT_LOGS"
docker compose logs --no-color client > "$CLIENT_LOGS" 2>&1

AGENT_LOGS="$LOG_DIR/agent-$DATE_SUFFIX.log"
echo "Saving agent logs to: $AGENT_LOGS"
docker compose logs --no-color agent > "$AGENT_LOGS" 2>&1

POSTGRES_LOGS="$LOG_DIR/postgres-$DATE_SUFFIX.log"
echo "Saving postgres logs to: $POSTGRES_LOGS"
docker compose logs --no-color postgres > "$POSTGRES_LOGS" 2>&1

echo ""
echo "Logs saved successfully:"
echo "  All services: $ALL_LOGS ($(du -h "$ALL_LOGS" | cut -f1))"
echo "  Client:       $CLIENT_LOGS ($(du -h "$CLIENT_LOGS" | cut -f1))"
echo "  Agent:        $AGENT_LOGS ($(du -h "$AGENT_LOGS" | cut -f1))"
echo "  Postgres:     $POSTGRES_LOGS ($(du -h "$POSTGRES_LOGS" | cut -f1))"
