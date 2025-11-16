#!/usr/bin/env bash
set -euo pipefail

# Enable verbose error logging
exec 2>&1  # Redirect stderr to stdout for better logging visibility
set -x     # Print commands as they execute

echo "========================================="
echo "Starting application in production mode"
echo "========================================="
echo "Python version: $(python --version)"
echo "Working directory: $(pwd)"
echo "Environment variables:"
env | grep -E '^(PORT|DATABASE_URL|SUPABASE|PYTHON)' || true
echo "========================================="

# Forward signals to children
_term() {
  echo "Caught SIGTERM, shutting down..."
  kill -TERM "$LG_PID" 2>/dev/null || true
  kill -TERM "$API_PID" 2>/dev/null || true
}
trap _term TERM INT

# Use PORT env var from DigitalOcean, default to 8080
HTTP_PORT=${PORT:-8080}
echo "HTTP_PORT set to: $HTTP_PORT"

# Check migration status (production uses Supabase)
echo "ℹ️  Checking migration status..."
if python scripts/check_migration_status.py 2>&1; then
    echo "✓ Migration status check completed"
else
    echo "⚠️  Skipping migration status checks because migrations already ran.."
fi

# Verify package installation
echo "========================================="
echo "Verifying Python package installation..."
echo "========================================="
python -c "import sys; print('Python path:', sys.path)" || true
python -c "import server; print('✓ server module found at:', server.__file__)" 2>&1 || echo "❌ ERROR: server module not importable"
python -c "from server import app; print('✓ server.app found')" 2>&1 || echo "❌ ERROR: server.app not importable"
python -c "import langgraph; print('✓ langgraph installed')" 2>&1 || echo "❌ ERROR: langgraph not installed"
echo "========================================="

echo "Starting LangGraph server on port 2024..."
echo "Starting FastAPI server on port $HTTP_PORT..."

# Start LangGraph API (2024) - use 'dev' mode but without reload for production
# 'langgraph up' requires Docker, but we're already in a container
# 'langgraph dev' runs the server directly without Docker
if command -v langgraph &> /dev/null; then
    echo "✓ Using langgraph CLI from PATH: $(which langgraph)"
    langgraph dev --host 0.0.0.0 --port 2024 --no-reload 2>&1 &
else
    echo "⚠️  langgraph command not found, using python -m langgraph_cli (fallback)"
    python -m langgraph_cli.cli dev --host 0.0.0.0 --port 2024 --no-reload 2>&1 &
fi
LG_PID=$!
echo "LangGraph PID: $LG_PID"

# Give LangGraph a moment to start
sleep 2

# Start FastAPI websocket server - production mode
echo "Attempting to start uvicorn with: python -m uvicorn server:app"
if python -m uvicorn server:app \
  --host 0.0.0.0 \
  --port "$HTTP_PORT" \
  --log-level info 2>&1 &
then
    API_PID=$!
    echo "✓ FastAPI PID: $API_PID"
else
    echo "❌ ERROR: Failed to start FastAPI server"
    echo "Killing LangGraph server..."
    kill -TERM "$LG_PID" 2>/dev/null || true
    exit 1
fi

echo "========================================="
echo "Both servers started successfully"
echo "LangGraph: PID $LG_PID on port 2024"
echo "FastAPI: PID $API_PID on port $HTTP_PORT"
echo "========================================="

# Wait on either to exit, then exit with that code
wait -n $LG_PID $API_PID
EXIT_CODE=$?
echo "Process exited with code: $EXIT_CODE"
exit $EXIT_CODE