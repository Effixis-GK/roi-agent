#!/bin/bash

# ROI Agent Enhanced - tcpdump DNS Monitoring Startup Script
# This script starts both the Go agent and Python Web UI

set -e

echo "=== ROI Agent Enhanced - Starting tcpdump DNS Monitoring ==="

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Project root: $PROJECT_ROOT"

# Check if required directories exist
if [ ! -d "$PROJECT_ROOT/agent" ]; then
    echo "❌ Error: agent directory not found at $PROJECT_ROOT/agent"
    exit 1
fi

if [ ! -d "$PROJECT_ROOT/web" ]; then
    echo "❌ Error: web directory not found at $PROJECT_ROOT/web"
    exit 1
fi

# Check if required files exist
if [ ! -f "$PROJECT_ROOT/agent/main.go" ]; then
    echo "❌ Error: main.go not found"
    exit 1
fi

if [ ! -f "$PROJECT_ROOT/web/enhanced_app.py" ]; then
    echo "❌ Error: enhanced_app.py not found"
    exit 1
fi

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check Python3 installation
if ! command -v python3 &> /dev/null; then
    echo "❌ Error: Python3 is not installed or not in PATH"
    echo "Please install Python3"
    exit 1
fi

# Check sudo permissions for tcpdump
echo "🔐 Checking sudo permissions for tcpdump..."
if ! sudo -n tcpdump --version &> /dev/null; then
    echo "⚠️  This script requires sudo permissions for tcpdump DNS monitoring."
    echo "   Please enter your password when prompted."
fi

echo "✅ Prerequisites check passed"

# Check for .env file in data-sender folder
echo ""
echo "📡 Data Transmission Setup:"
echo "==========================="

ENV_FILE="$PROJECT_ROOT/data-sender/.env"
if [ -f "$ENV_FILE" ]; then
    echo "✅ Found .env file at $ENV_FILE"
    echo "📄 Loading environment variables from .env file..."
    
    # Load environment variables from .env file
    while IFS= read -r line; do
        # Skip empty lines and comments
        if [[ "$line" =~ ^[[:space:]]*$ ]] || [[ "$line" =~ ^[[:space:]]*# ]]; then
            continue
        fi
        # Export the variable
        if [[ "$line" =~ ^[[:space:]]*([^=]+)=(.*)$ ]]; then
            export "${BASH_REMATCH[1]}"="${BASH_REMATCH[2]}"
            echo "📝 Loaded: ${BASH_REMATCH[1]}=${BASH_REMATCH[2]}"
        fi
    done < "$ENV_FILE"
    
    # Check if required variables are set
    if [ "$ROI_AGENT_BASE_URL" != "" ] && [ "$ROI_AGENT_API_KEY" != "" ]; then
        echo "✅ Data transmission will be enabled"
        echo "📡 Server URL: $ROI_AGENT_BASE_URL"
        echo "🔑 API Key: ${ROI_AGENT_API_KEY:0:8}..."
    else
        echo "⚠️  .env file found but missing required variables. Data transmission will be disabled."
    fi
else
    echo "⚠️  No .env file found at $ENV_FILE"
    echo "Do you want to enable data transmission? (y/n)"
    read -p "Choice (default: n): " enable_transmission
    
    if [ "$enable_transmission" = "y" ] || [ "$enable_transmission" = "Y" ]; then
        echo "📝 Please enter your server details:"
        read -p "Base URL (e.g., https://api.yourserver.com/v1/roi-agent): " base_url
        read -p "API Key: " api_key
        
        if [ "$base_url" != "" ] && [ "$api_key" != "" ]; then
            export ROI_AGENT_BASE_URL="$base_url"
            export ROI_AGENT_API_KEY="$api_key"
            
            # Save to .env file for future use
            echo "💾 Saving configuration to .env file..."
            mkdir -p "$(dirname "$ENV_FILE")"
            cat > "$ENV_FILE" << EOF
# ROI Agent Data Transmission Environment Variables
# Replace with your actual server URL and API key

ROI_AGENT_BASE_URL=$base_url
ROI_AGENT_API_KEY=$api_key
EOF
            echo "✅ Configuration saved to $ENV_FILE"
            echo "✅ Data transmission will be enabled"
        else
            echo "⚠️  Missing URL or API key. Data transmission will be disabled."
        fi
    else
        echo "ℹ️  Data transmission will be disabled"
    fi
fi

# Kill any existing processes
echo "🔄 Stopping any existing ROI Agent processes..."
sudo pkill -f "tcpdump.*port 53" || true
pkill -f "main.go" || true
pkill -f "enhanced_app.py" || true
sleep 2

# Create log directory
LOG_DIR="$HOME/.roiagent/logs"
mkdir -p "$LOG_DIR"

echo "📁 Log directory: $LOG_DIR"

# Start Go agent with sudo for tcpdump permissions
echo "🚀 Starting Go agent (tcpdump DNS Monitor)..."
cd "$PROJECT_ROOT/agent"

# Pass environment variables explicitly to sudo
SUDO_ENV=""
if [ "$ROI_AGENT_BASE_URL" != "" ]; then
    SUDO_ENV="$SUDO_ENV ROI_AGENT_BASE_URL=$ROI_AGENT_BASE_URL"
fi
if [ "$ROI_AGENT_API_KEY" != "" ]; then
    SUDO_ENV="$SUDO_ENV ROI_AGENT_API_KEY=$ROI_AGENT_API_KEY"
fi
if [ "$ROI_AGENT_INTERVAL_MINUTES" != "" ]; then
    SUDO_ENV="$SUDO_ENV ROI_AGENT_INTERVAL_MINUTES=$ROI_AGENT_INTERVAL_MINUTES"
    echo "📅 Setting transmission interval: $ROI_AGENT_INTERVAL_MINUTES minutes"
fi

echo "🔧 Passing environment variables to agent: $SUDO_ENV"
nohup sudo env $SUDO_ENV go run main.go > "$LOG_DIR/agent.log" 2>&1 &
AGENT_PID=$!
echo "   Agent PID: $AGENT_PID"

# Wait a moment for agent to start
sleep 3

# Start Python Web UI in background
echo "🌐 Starting Python Web UI..."
cd "$PROJECT_ROOT/web"
nohup python3 enhanced_app.py > "$LOG_DIR/webui.log" 2>&1 &
WEBUI_PID=$!
echo "   Web UI PID: $WEBUI_PID"

# Wait for web server to start
echo "⏳ Waiting for services to start..."
sleep 5

# Check if processes are still running
if ! sudo kill -0 "$AGENT_PID" 2>/dev/null; then
    echo "❌ Agent failed to start. Check logs:"
    echo "   tail -f $LOG_DIR/agent.log"
    exit 1
fi

if ! kill -0 "$WEBUI_PID" 2>/dev/null; then
    echo "❌ Web UI failed to start. Check logs:"
    echo "   tail -f $LOG_DIR/webui.log"
    exit 1
fi

# Test if web server is responding
echo "🔍 Testing web server..."
if curl -s http://localhost:5002 > /dev/null; then
    echo "✅ Web server is responding"
else
    echo "⚠️  Web server may not be ready yet, but continuing..."
fi

# Open dashboard in browser
echo "🎯 Opening dashboard in browser..."
if command -v open &> /dev/null; then
    open http://localhost:5002
elif command -v xdg-open &> /dev/null; then
    xdg-open http://localhost:5002
else
    echo "   Please open http://localhost:5002 manually in your browser"
fi

echo ""
echo "🎉 ROI Agent Enhanced started successfully!"
echo ""
echo "📊 Dashboard: http://localhost:5002"
echo "📁 Logs:"
echo "   Agent:  tail -f $LOG_DIR/agent.log"
echo "   Web UI: tail -f $LOG_DIR/webui.log"
echo ""
echo "🛑 To stop all services:"
echo "   sudo pkill -f \"tcpdump.*port 53\""
echo "   pkill -f main.go"
echo "   pkill -f enhanced_app.py"
echo ""
echo "ℹ️  tcpdump DNS monitoring requires sudo permissions."
echo "   The agent is running with elevated privileges for network access."
echo ""
echo "Press Ctrl+C to stop monitoring the logs, or run in background."

# Follow the logs (user can Ctrl+C to exit)
echo "📋 Following logs (Ctrl+C to exit)..."
sleep 2
tail -f "$LOG_DIR/agent.log" "$LOG_DIR/webui.log"
