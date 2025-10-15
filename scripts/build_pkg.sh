#!/bin/bash

# ROI Agent - macOS PKG Installer Builder
# Production version - No Web UI, background service only

set -e

echo "📦 ROI Agent - PKG Installer Builder"
echo "========================================"

# 設定
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build/pkg"
VERSION="${1:-1.0.4}"
ARCH="${2:-arm64}"

if [ "$ARCH" = "arm64" ]; then
    ARCH_LABEL="arm64"
    GO_ARCH="arm64"
elif [ "$ARCH" = "x64" ]; then
    ARCH_LABEL="x64"
    GO_ARCH="amd64"
else
    echo "❌ Invalid architecture: $ARCH (use 'arm64' or 'x64')"
    exit 1
fi

echo "Building for: macOS $ARCH_LABEL"
echo "Version: $VERSION"
echo ""

# クリーンアップ
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# インストール先のディレクトリ構造
PAYLOAD_DIR="$BUILD_DIR/payload"
INSTALL_DIR="$PAYLOAD_DIR/Applications/ROI Agent"
RESOURCES_DIR="$INSTALL_DIR/Resources"
BIN_DIR="$INSTALL_DIR/bin"

mkdir -p "$INSTALL_DIR"
mkdir -p "$RESOURCES_DIR"
mkdir -p "$BIN_DIR"

echo "🔨 Building binaries..."

# 1. Agent バイナリ
cd "$PROJECT_ROOT/agent"
GOOS=darwin GOARCH=$GO_ARCH go build -ldflags="-s -w" -o "$BIN_DIR/roi-agent" main.go
chmod +x "$BIN_DIR/roi-agent"
echo "  ✅ roi-agent ($ARCH_LABEL)"

# 2. Data Sender バイナリ
cd "$PROJECT_ROOT/data-sender"
GOOS=darwin GOARCH=$GO_ARCH go build -ldflags="-s -w" -o "$BIN_DIR/data-sender" .
chmod +x "$BIN_DIR/data-sender"
echo "  ✅ data-sender ($ARCH_LABEL)"

cd "$PROJECT_ROOT"

# 3. 設定ファイルテンプレート
cat > "$RESOURCES_DIR/.env.template" << 'EOF'
# ROI Agent Configuration
# This file will be configured during download from ROI Dashboard
ROI_AGENT_BASE_URL=__BASE_URL__
ROI_AGENT_API_KEY=__API_KEY__
ROI_AGENT_INTERVAL_MINUTES=10
EOF
echo "  ✅ .env.template"

# 4. README
cat > "$INSTALL_DIR/README.txt" << EOF
ROI Agent - macOS Background Service
=====================================

Version: $VERSION
Architecture: $ARCH_LABEL

INSTALLATION:
This installer will place ROI Agent in:
  /Applications/ROI Agent/

The agent runs as a background service and automatically:
- Monitors application usage
- Tracks network connections
- Sends data to your organization's dashboard

COMPONENTS:
- roi-agent: Main monitoring agent
- data-sender: Data transmission service

USAGE:
After installation, the service starts automatically as a LaunchAgent.

Manual control:
  Start:  launchctl start com.roiagent
  Stop:   launchctl stop com.roiagent
  Status: launchctl list | grep roiagent

UNINSTALLATION:
  /Applications/ROI Agent/bin/uninstall.sh

LOGS:
  ~/.roiagent/logs/agent.log
  ~/.roiagent/logs/launchagent.log

SUPPORT:
Visit: https://roi-dashboard-607617540267.asia-northeast1.run.app
EOF
echo "  ✅ README.txt"

# 5. LaunchAgent plist
cat > "$RESOURCES_DIR/com.roiagent.plist" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.roiagent</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>sudo</string>
        <string>/Applications/ROI Agent/bin/roi-agent</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
    
    <key>StandardOutPath</key>
    <string>/Users/Shared/.roiagent/logs/launchagent.log</string>
    
    <key>StandardErrorPath</key>
    <string>/Users/Shared/.roiagent/logs/launchagent.err</string>
    
    <key>WorkingDirectory</key>
    <string>/Applications/ROI Agent</string>
</dict>
</plist>
EOF
echo "  ✅ com.roiagent.plist"

# 6. インストールスクリプト
cat > "$BIN_DIR/install-launchagent.sh" << 'EOF'
#!/bin/bash
# ROI Agent - Install LaunchAgent

PLIST_SRC="/Applications/ROI Agent/Resources/com.roiagent.plist"
PLIST_DST="$HOME/Library/LaunchAgents/com.roiagent.plist"

echo "Installing ROI Agent LaunchAgent..."

# Create LaunchAgents directory
mkdir -p "$HOME/Library/LaunchAgents"

# Copy plist
cp "$PLIST_SRC" "$PLIST_DST"

# Load LaunchAgent
launchctl load "$PLIST_DST"

echo "✅ LaunchAgent installed and loaded"
echo "   ROI Agent will start automatically on login"
EOF
chmod +x "$BIN_DIR/install-launchagent.sh"
echo "  ✅ install-launchagent.sh"

# 7. アンインストールスクリプト
cat > "$BIN_DIR/uninstall.sh" << 'EOF'
#!/bin/bash
# ROI Agent - Uninstall Script

echo "🗑️  Uninstalling ROI Agent..."

# Stop and unload LaunchAgent
launchctl unload ~/Library/LaunchAgents/com.roiagent.plist 2>/dev/null
rm -f ~/Library/LaunchAgents/com.roiagent.plist

# Stop any running processes
sudo pkill -f "roi-agent" 2>/dev/null
sudo pkill -f "data-sender" 2>/dev/null
sudo pkill -f "tcpdump.*port 53" 2>/dev/null

# Remove application
sudo rm -rf "/Applications/ROI Agent"

# Ask about user data
read -p "Remove user data (~/.roiagent)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf ~/.roiagent
    echo "  ✅ User data removed"
fi

echo "✅ ROI Agent uninstalled"
EOF
chmod +x "$BIN_DIR/uninstall.sh"
echo "  ✅ uninstall.sh"

# 8. バージョン情報
cat > "$INSTALL_DIR/version.json" << EOF
{
  "version": "$VERSION",
  "architecture": "$ARCH_LABEL",
  "build_date": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "format": "PKG",
  "components": {
    "agent": "roi-agent",
    "data_sender": "data-sender"
  },
  "features": [
    "Application usage monitoring",
    "Network connection tracking",
    "Automatic data transmission",
    "Background service (LaunchAgent)"
  ]
}
EOF
echo "  ✅ version.json"

echo ""
echo "📦 Building PKG installer..."

PKG_ID="com.roiagent.pkg"
PKG_VERSION="$VERSION"
PKG_NAME="ROI-Agent-macOS-${ARCH_LABEL}-${VERSION}.pkg"
PKG_OUTPUT="$BUILD_DIR/$PKG_NAME"

# PKGビルド
pkgbuild \
    --root "$PAYLOAD_DIR" \
    --identifier "$PKG_ID" \
    --version "$PKG_VERSION" \
    --install-location "/" \
    "$PKG_OUTPUT"

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ PKG created successfully!"
    echo ""
    echo "📊 Package Information:"
    echo "   File: $PKG_NAME"
    echo "   Size: $(du -h "$PKG_OUTPUT" | cut -f1)"
    echo "   Path: $PKG_OUTPUT"
    echo ""
    
    # SHA256計算
    SHA256=$(shasum -a 256 "$PKG_OUTPUT" | awk '{print $1}')
    echo "   SHA256: $SHA256"
    echo "$SHA256" > "$PKG_OUTPUT.sha256"
    echo ""
    
    echo "🎉 Build completed!"
else
    echo "❌ Failed to create PKG"
    exit 1
fi
