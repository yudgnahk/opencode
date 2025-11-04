#!/bin/bash
# Script to run OpenCode server and TUI together

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_PATH="$PROJECT_ROOT/bin/opencode"

# Parse arguments
USE_RANDOM_PORT=false
PORT=8080
while [[ $# -gt 0 ]]; do
    case $1 in
        --random-port)
            USE_RANDOM_PORT=true
            shift
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        *)
            break
            ;;
    esac
done

# Check if binary exists
if [ ! -f "$BIN_PATH" ]; then
    echo "Error: opencode binary not found at $BIN_PATH"
    echo "Please run 'make build' first"
    exit 1
fi

# Find a random available port if requested
if [ "$USE_RANDOM_PORT" = true ]; then
    # Use Python to find a random available port
    PORT=$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()')
    echo "Using random port: $PORT"
fi

# Start server in background
echo "Starting OpenCode server on port $PORT..."
"$BIN_PATH" serve --port "$PORT" > /tmp/opencode-server.log 2>&1 &
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
    if curl -s "http://localhost:$PORT/health" > /dev/null 2>&1; then
        echo "Server is ready on http://localhost:$PORT!"
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
export OPENCODE_SERVER="http://localhost:$PORT"
"$BIN_PATH" tui "$@"
