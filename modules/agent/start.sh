#!/usr/bin/env bash
set -euo pipefail

# Forward signals to children
_term() {
  echo "Caught SIGTERM, shutting down..."
  kill -TERM "$LG_PID" 2>/dev/null || true
  kill -TERM "$API_PID" 2>/dev/null || true
}
trap _term TERM INT

# Run Supabase migrations if configured
if [ -n "${SUPABASE_URL:-}" ] && [ -n "${SUPABASE_SERVICE_KEY:-}" ]; then
  echo "🔄 Running Supabase migrations..."
  if uv run python scripts/apply_migrations.py; then
    echo "✅ Migrations completed successfully"
  else
    echo "⚠️  Migrations failed, continuing anyway..."
  fi
else
  echo "⏭️  Skipping migrations (SUPABASE_URL or SUPABASE_SERVICE_KEY not set)"
fi

# Start LangGraph dev API (2024) - already has hot-reload built-in
uv run langgraph dev --host 0.0.0.0 --port 2024 &
LG_PID=$!

# Start your FastAPI websocket server (8000) with hot-reload
uv run python -m uvicorn agent.server:app \
  --host 0.0.0.0 \
  --port 8000 \
  --reload \
  --reload-delay 0.5 \
  --log-level warning &
API_PID=$!

# Wait on either to exit, then exit with that code
wait -n $LG_PID $API_PID
exit $?