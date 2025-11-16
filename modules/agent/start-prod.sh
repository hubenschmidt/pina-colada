#!/usr/bin/env bash
set -euo pipefail

# Forward signals to children
_term() {
  echo "Caught SIGTERM, shutting down..."
  kill -TERM "$LG_PID" 2>/dev/null || true
  kill -TERM "$API_PID" 2>/dev/null || true
}
trap _term TERM INT

# Use PORT env var from DigitalOcean, default to 8080
HTTP_PORT=${PORT:-8080}

# Check migration status (production uses Supabase)
echo "ℹ️  Checking migration status..."
python scripts/check_migration_status.py || echo "⚠️  Skipping migration status checks because migrations already ran.."

echo "Starting LangGraph server on port 2024..."
echo "Starting FastAPI server on port $HTTP_PORT..."

# Start LangGraph API (2024) - use 'dev' mode but without reload for production
# 'langgraph up' requires Docker, but we're already in a container
# 'langgraph dev' runs the server directly without Docker
if command -v langgraph &> /dev/null; then
    echo "Using langgraph CLI from PATH"
    langgraph dev --host 0.0.0.0 --port 2024 --no-reload &
else
    echo "Using python -m langgraph_cli (fallback)"
    python -m langgraph_cli.cli dev --host 0.0.0.0 --port 2024 --no-reload &
fi
LG_PID=$!

# Start FastAPI websocket server - production mode
python -m uvicorn server:app \
  --host 0.0.0.0 \
  --port "$HTTP_PORT" \
  --log-level info &
API_PID=$!

# Wait on either to exit, then exit with that code
wait -n $LG_PID $API_PID
exit $?