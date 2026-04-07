#!/bin/bash
# Start the Anthropic-to-Ollama proxy for Claude Code
# Usage: ./start_claude_proxy.sh [port]

PORT=${1:-8082}
PIDFILE="/tmp/anthropic_ollama_proxy.pid"
LOGFILE="/tmp/anthropic_proxy.log"

# Check if already running
if [ -f "$PIDFILE" ]; then
    OLD_PID=$(cat "$PIDFILE")
    if ps -p "$OLD_PID" > /dev/null 2>&1; then
        echo "Proxy already running on port $PORT (PID: $OLD_PID)"
        echo "To restart: kill $OLD_PID && $0 $PORT"
        exit 0
    fi
    rm -f "$PIDFILE"
fi

# Kill any existing proxy on the port
pkill -f "anthropic_ollama_proxy.py" 2>/dev/null
sleep 1

# Start the proxy
PROXY_SCRIPT="$HOME/.claude/anthropic_ollama_proxy.py"
if [ ! -f "$PROXY_SCRIPT" ]; then
    echo "ERROR: Proxy script not found at $PROXY_SCRIPT"
    exit 1
fi

python3 "$PROXY_SCRIPT" "$PORT" > "$LOGFILE" 2>&1 &
PROXY_PID=$!
echo "$PROXY_PID" > "$PIDFILE"

# Wait for it to start
sleep 2

# Verify it's running
if curl -s "http://localhost:$PORT/health" > /dev/null 2>&1; then
    echo "✅ Anthropic-to-Ollama proxy started on http://localhost:$PORT"
    echo "   PID: $PROXY_PID"
    echo "   Log: $LOGFILE"
    echo ""
    echo "Available cloud models:"
    curl -s "http://localhost:$PORT/v1/models" | python3 -c "
import json, sys
data = json.load(sys.stdin)
for m in data.get('data', []):
    mid = m.get('id', '')
    if 'cloud' in mid.lower():
        print(f'   - {mid}')
"
    echo ""
    echo "To stop: kill $PROXY_PID"
else
    echo "❌ Failed to start proxy. Check $LOGFILE for errors."
    rm -f "$PIDFILE"
    exit 1
fi
