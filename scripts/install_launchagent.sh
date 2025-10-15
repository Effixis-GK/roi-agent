#!/bin/bash

# ROI Agent - macOS LaunchAgent installer
# Configures ROI Agent to start automatically on user login

set -e

INSTALL_DIR="/Applications/ROI Agent"
LAUNCH_AGENT_PLIST="$HOME/Library/LaunchAgents/com.roiagent.plist"

echo "🚀 ROI Agent - LaunchAgent Setup"
echo "================================="

# Check if ROI Agent is installed
if [ ! -d "$INSTALL_DIR" ]; then
    echo "❌ Error: ROI Agent not found at $INSTALL_DIR"
    echo "   Please install ROI Agent first"
    exit 1
fi

# Create LaunchAgents directory if it doesn't exist
mkdir -p "$HOME/Library/LaunchAgents"

# Create LaunchAgent plist
cat > "$LAUNCH_AGENT_PLIST" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.roiagent</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>/Applications/ROI Agent/bin/start.sh</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
    
    <key>StandardOutPath</key>
    <string>$HOME/.roiagent/logs/launchagent.log</string>
    
    <key>StandardErrorPath</key>
    <string>$HOME/.roiagent/logs/launchagent.err</string>
    
    <key>WorkingDirectory</key>
    <string>/Applications/ROI Agent</string>
    
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
</dict>
</plist>
EOF

echo "✅ LaunchAgent plist created: $LAUNCH_AGENT_PLIST"

# Load LaunchAgent
launchctl load "$LAUNCH_AGENT_PLIST"

if [ $? -eq 0 ]; then
    echo "✅ LaunchAgent loaded successfully"
    echo ""
    echo "ROI Agent will now start automatically on login"
    echo ""
    echo "To manually control:"
    echo "  Start:  launchctl start com.roiagent"
    echo "  Stop:   launchctl stop com.roiagent"
    echo "  Remove: launchctl unload $LAUNCH_AGENT_PLIST"
else
    echo "❌ Failed to load LaunchAgent"
    exit 1
fi
