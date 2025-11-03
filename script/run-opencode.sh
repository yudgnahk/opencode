#!/bin/bash
# Script to run OpenCode server and TUI together

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_PATH="$PROJECT_ROOT/bin/opencode"

# Check if binary exists
if [ ! -f "$BIN_PATH" ]; then
    echo "Error: opencode binary not found at $BIN_PATH"
    echo "Please run 'make build' first"
    exit 1
fi

# Start server in background
echo "Starting OpenCode server..."
"$BIN_PATH" serve > /tmp/opencode-server.log 2>&1 &
SERVER_PID=$!

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Stopping OpenCode server..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
}

trap cleanup EXIT INT TERM

# Wait for server to be ready
echo "Waiting for server to start..."
for i in {1..30}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "Server is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "Error: Server failed to start within 30 seconds"
        echo "Check logs at /tmp/opencode-server.log"
        exit 1
    fi
    sleep 1
done

# Start TUI
echo "Starting OpenCode TUI..."
export OPENCODE_SERVER=http://localhost:8080
"$BIN_PATH" tui "$@"
