#!/bin/bash

# ROI Agent - Mac App Builder
# Creates a native macOS application with custom icon

set -e

echo "🍎 Building ROI Agent for macOS..."

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Project root: $PROJECT_ROOT"

# Configuration
APP_NAME="ROI Agent"
APP_BUNDLE_ID="com.roiagent.monitor"
APP_VERSION="1.0.0"
BUILD_DIR="$PROJECT_ROOT/build"
APP_DIR="$BUILD_DIR/$APP_NAME.app"
CONTENTS_DIR="$APP_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"
APP_ICON_PATH="$PROJECT_ROOT/public/icon.png"

# Clean and create build directory
echo "🧹 Cleaning build directory..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed"
    exit 1
fi

# Check if Python is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Error: Python3 is not installed"
    exit 1
fi

# Build Go agent
echo "🔨 Building Go agent..."
cd "$PROJECT_ROOT/agent"
GOOS=darwin GOARCH=amd64 go build -o "$MACOS_DIR/roi-agent" main.go
chmod +x "$MACOS_DIR/roi-agent"

# Copy Python Web UI
echo "📦 Copying Web UI..."
cp -r "$PROJECT_ROOT/web" "$RESOURCES_DIR/"

# Copy scripts
echo "📜 Copying scripts..."
mkdir -p "$RESOURCES_DIR/scripts"
cp "$PROJECT_ROOT/scripts/start_enhanced_fqdn_monitoring.sh" "$RESOURCES_DIR/scripts/"
cp "$PROJECT_ROOT/scripts/stop_enhanced_monitoring.sh" "$RESOURCES_DIR/scripts/"
chmod +x "$RESOURCES_DIR/scripts/"*.sh

# Create Info.plist
echo "📄 Creating Info.plist..."
cat > "$CONTENTS_DIR/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>roi-agent-launcher</string>
    <key>CFBundleIdentifier</key>
    <string>$APP_BUNDLE_ID</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleVersion</key>
    <string>$APP_VERSION</string>
    <key>CFBundleShortVersionString</key>
    <string>$APP_VERSION</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleSignature</key>
    <string>????</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15.0</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.productivity</string>
    <key>NSAppleEventsUsageDescription</key>
    <string>ROI Agent needs AppleScript access to monitor application usage.</string>
    <key>NSSystemAdministrationUsageDescription</key>
    <string>ROI Agent needs administrator privileges for network monitoring.</string>
</dict>
</plist>
EOF

# Create launcher script
echo "🚀 Creating launcher script..."
cat > "$MACOS_DIR/roi-agent-launcher" << 'EOF'
#!/bin/bash

# ROI Agent Launcher
# This script launches both the Go agent and Python Web UI

# Get the directory where this app bundle is located
APP_DIR="$(dirname "$(dirname "$(dirname "$(realpath "$0")")")")" 
RESOURCES_DIR="$APP_DIR/Contents/Resources"
MAC_OS_DIR="$APP_DIR/Contents/MacOS"

# Change to resources directory
cd "$RESOURCES_DIR"

# Check accessibility permissions
"$MAC_OS_DIR/roi-agent" check-permissions
if [ $? -ne 0 ]; then
    osascript -e 'display dialog "ROI Agent needs Accessibility permissions.\n\nPlease go to:\nSystem Settings > Privacy & Security > Accessibility\n\nAnd add this application." buttons {"OK"} default button "OK"'
    exit 1
fi

# Check sudo permissions
if ! sudo -n true 2>/dev/null; then
    osascript -e 'display dialog "ROI Agent needs administrator privileges for network monitoring.\n\nPlease enter your password when prompted." buttons {"OK"} default button "OK"'
fi

# Create logs directory
mkdir -p "$HOME/.roiagent/logs"

# Start the agent in background
"$MAC_OS_DIR/roi-agent" > "$HOME/.roiagent/logs/agent.log" 2>&1 &
AGENT_PID=$!

# Wait a moment for agent to start
sleep 3

# Start the web UI in background
cd "$RESOURCES_DIR/web"
python3 enhanced_app.py > "$HOME/.roiagent/logs/webui.log" 2>&1 &
WEBUI_PID=$!

# Wait for web server to start
sleep 5

# Open dashboard in browser
open http://localhost:5002

# Show notification
osascript -e 'display notification "ROI Agent is now running. Dashboard opened in browser." with title "ROI Agent"'

# Keep the launcher running
wait $AGENT_PID $WEBUI_PID
EOF

chmod +x "$MACOS_DIR/roi-agent-launcher"

# Handle app icon
if [ -f "$APP_ICON_PATH" ]; then
    echo "🎨 Adding custom app icon..."
    
    # Check if we have sips command to convert icon
    if command -v sips &> /dev/null; then
        # Create different sizes for icon set
        ICONSET_DIR="$BUILD_DIR/AppIcon.iconset"
        mkdir -p "$ICONSET_DIR"
        
        # Generate all required icon sizes
        sips -z 16 16 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_16x16.png"
        sips -z 32 32 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_16x16@2x.png"
        sips -z 32 32 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_32x32.png"
        sips -z 64 64 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_32x32@2x.png"
        sips -z 128 128 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_128x128.png"
        sips -z 256 256 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_128x128@2x.png"
        sips -z 256 256 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_256x256.png"
        sips -z 512 512 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_256x256@2x.png"
        sips -z 512 512 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_512x512.png"
        sips -z 1024 1024 "$APP_ICON_PATH" --out "$ICONSET_DIR/icon_512x512@2x.png"
        
        # Convert to .icns file
        if command -v iconutil &> /dev/null; then
            iconutil -c icns "$ICONSET_DIR" -o "$RESOURCES_DIR/AppIcon.icns"
            echo "✅ Custom app icon created"
        else
            echo "⚠️  iconutil not found, using PNG icon directly"
            cp "$APP_ICON_PATH" "$RESOURCES_DIR/AppIcon.png"
        fi
        
        # Clean up iconset
        rm -rf "$ICONSET_DIR"
    else
        echo "⚠️  sips not found, copying PNG icon directly"
        cp "$APP_ICON_PATH" "$RESOURCES_DIR/AppIcon.png"
    fi
else
    echo "⚠️  App icon not found at $APP_ICON_PATH"
    echo "   Place your app icon (512x512 PNG) at public/icon.png"
fi

# Create README for the app
cat > "$RESOURCES_DIR/README.txt" << EOF
ROI Agent - macOS Application & Network Monitor

This application monitors:
- Application usage time and focus time
- Network connections via DNS monitoring

Requirements:
- macOS with Accessibility permissions
- Administrator privileges for network monitoring

Dashboard: http://localhost:5002

Data Storage: ~/.roiagent/
Logs: ~/.roiagent/logs/

To stop monitoring:
1. Quit this application
2. Or run: sudo pkill -f "tcpdump.*port 53" && pkill -f "roi-agent"
EOF

echo "🎉 Mac app build complete!"
echo ""
echo "📍 App location: $APP_DIR"
echo "🖼️  Icon: $([ -f "$APP_ICON_PATH" ] && echo "Custom icon applied" || echo "No custom icon (place at public/icon.png)")"
echo ""
echo "🔧 To install:"
echo "   1. Copy '$APP_NAME.app' to /Applications/"
echo "   2. Right-click > Open (first time only)"
echo ""
echo "📋 Next steps:"
echo "   1. Grant Accessibility permissions when prompted"
echo "   2. Enter admin password for network monitoring"
echo "   3. Dashboard will open automatically"

# Optionally open the build directory
if command -v open &> /dev/null; then
    echo "📂 Opening build directory..."
    open "$BUILD_DIR"
fi
