#!/bin/bash
# View saved log files

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_ROOT/logs"
DATE_SUFFIX=$(date +%Y-%m-%d)

SERVICE="${1:-all}"
DATE="${2:-$DATE_SUFFIX}"

case "$SERVICE" in
  client)
    LOG_FILE="$LOG_DIR/client-$DATE.log"
    ;;
  agent)
    LOG_FILE="$LOG_DIR/agent-$DATE.log"
    ;;
  postgres)
    LOG_FILE="$LOG_DIR/postgres-$DATE.log"
    ;;
  all)
    LOG_FILE="$LOG_DIR/docker-compose-$DATE.log"
    ;;
  *)
    echo "Usage: $0 [client|agent|postgres|all] [YYYY-MM-DD]"
    echo ""
    echo "Examples:"
    echo "  $0                    # View all logs from today"
    echo "  $0 client             # View client logs from today"
    echo "  $0 agent              # View agent logs from today"
    echo "  $0 postgres           # View postgres logs from today"
    echo "  $0 agent 2025-11-16   # View agent logs from specific date"
    echo ""
    echo "Available log files:"
    ls -lh "$LOG_DIR"/*.log 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'
    exit 1
    ;;
esac

if [ ! -f "$LOG_FILE" ]; then
  echo "Error: Log file not found: $LOG_FILE"
  echo ""
  echo "Available log files:"
  ls -lh "$LOG_DIR"/*.log 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'
  exit 1
fi

echo "Viewing: $LOG_FILE"
echo "========================================"
cat "$LOG_FILE"
