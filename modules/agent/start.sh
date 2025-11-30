#!/usr/bin/env bash
set -euo pipefail

# Forward signals to children
_term() {
  echo "Caught SIGTERM, shutting down..."
  kill -TERM "$LG_PID" 2>/dev/null || true
  kill -TERM "$API_PID" 2>/dev/null || true
}
trap _term TERM INT

# Check migration status and apply if needed (local development only)
# Note: Supabase is only used in production via Azure Pipeline
echo "ℹ️  Checking migration status..."
uv run python scripts/check_migration_status.py || echo "⚠️  Could not check migration status"

# Auto-apply migrations for local Postgres
echo "ℹ️  Checking if migrations need to be applied to local Postgres..."
uv run python scripts/apply_migrations.py || echo "⚠️  Could not apply migrations automatically"

# Run seeders after migrations (only inserts if data doesn't exist)
echo "ℹ️  Running database seeders..."
uv run python scripts/run_seeders.py || echo "⚠️  Could not run seeders"

# Upload seed document files to storage
echo "ℹ️  Uploading seed document files..."
uv run python scripts/seed_documents.py || echo "⚠️  Could not upload seed documents"

# Start LangGraph dev API (2024) - already has hot-reload built-in
# Quiet LangGraph runtime queue stats by setting log level
export LOG_LEVEL=WARNING
uv run langgraph dev --host 0.0.0.0 --port 2024 &
LG_PID=$!

# Start your FastAPI websocket server (8000) with hot-reload
uv run python -m uvicorn server:app \
  --host 0.0.0.0 \
  --port 8000 \
  --reload \
  --reload-delay 0.5 \
  --log-level warning &
API_PID=$!

# Wait on either to exit, then exit with that code
wait -n $LG_PID $API_PID
exit $?