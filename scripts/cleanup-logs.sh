#!/bin/bash
# Cleanup old logs (older than 2 days)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_ROOT/logs"

mkdir -p "$LOG_DIR"

echo "Cleaning up old log files..."
find "$LOG_DIR" -name "*.log" -type f -mtime +2 -delete
echo "Log cleanup complete"
