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

echo "Starting LangGraph server on port 2024..."
echo "Starting FastAPI server on port $HTTP_PORT..."

# Start LangGraph API (2024) - production mode
# Use python -m since langgraph command may not be in PATH
python -m langgraph up --port 2024 &
LG_PID=$!

# Start FastAPI websocket server - production mode
python -m uvicorn agent.server:app \
  --host 0.0.0.0 \
  --port "$HTTP_PORT" \
  --log-level info &
API_PID=$!

# Wait on either to exit, then exit with that code
wait -n $LG_PID $API_PID
exit $?