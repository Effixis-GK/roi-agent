#!/bin/bash

# ROI Agent - macOS PKG Template Builder
# Creates PKG without org-specific config (.env will be injected on download)

set -e

echo "📦 ROI Agent - PKG Template Builder"
echo "========================================"

# 設定
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build/pkg"
VERSION="${1:-1.2.0}"
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

mkdir -p "$BUILD_DIR"

PAYLOAD_DIR="$BUILD_DIR/payload_${ARCH_LABEL}"
SCRIPTS_DIR="$BUILD_DIR/scripts_${ARCH_LABEL}"

rm -rf "$PAYLOAD_DIR" "$SCRIPTS_DIR"
mkdir -p "$PAYLOAD_DIR"
mkdir -p "$SCRIPTS_DIR"

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

# 3. .env.template（Dashboard APIが.envを生成する際の参照用）
cat > "$RESOURCES_DIR/.env.template" << 'EOF'
# ROI Agent Configuration
# This file will be replaced with actual configuration on download
ROI_AGENT_BASE_URL=https://roi-service-607617540267.asia-northeast1.run.app/api/v1/device
ROI_AGENT_API_KEY=YOUR_API_KEY_HERE
ROI_AGENT_INTERVAL_MINUTES=10
EOF
echo "  ✅ .env.template"

# 4. README
cat > "$INSTALL_DIR/README.txt" << EOF
ROI Agent - macOS Background Service
=====================================

Version: $VERSION
Architecture: $ARCH_LABEL

Pre-configured for your organization.
No manual configuration required.

UNINSTALLATION:
  sudo /Applications/ROI Agent/bin/uninstall.sh

LOGS:
  ~/.roiagent/logs/agent.log

SUPPORT:
  https://roi-dashboard-607617540267.asia-northeast1.run.app
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
    <string>/tmp/roiagent.stdout</string>
    
    <key>StandardErrorPath</key>
    <string>/tmp/roiagent.stderr</string>
    
    <key>WorkingDirectory</key>
    <string>/Applications/ROI Agent</string>
</dict>
</plist>
EOF
echo "  ✅ com.roiagent.plist"

# 6. アンインストールスクリプト
cat > "$BIN_DIR/uninstall.sh" << 'EOF'
#!/bin/bash
echo "🗑️  Uninstalling ROI Agent..."

CURRENT_USER=$(stat -f "%Su" /dev/console)
USER_HOME=$(eval echo ~$CURRENT_USER)

sudo -u "$CURRENT_USER" launchctl bootout "gui/$(id -u $CURRENT_USER)/com.roiagent" 2>/dev/null || true
rm -f "$USER_HOME/Library/LaunchAgents/com.roiagent.plist"

pkill -f "roi-agent" 2>/dev/null || true
pkill -f "data-sender" 2>/dev/null || true

rm -rf "/Applications/ROI Agent"

echo ""
read -p "Remove user data ($USER_HOME/.roiagent)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf "$USER_HOME/.roiagent"
    echo "  ✅ User data removed"
fi

echo "✅ ROI Agent uninstalled"
EOF
chmod +x "$BIN_DIR/uninstall.sh"
echo "  ✅ uninstall.sh"

# 7. バージョン情報
cat > "$INSTALL_DIR/version.json" << EOF
{
  "version": "$VERSION",
  "architecture": "$ARCH_LABEL",
  "build_date": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "format": "PKG_TEMPLATE"
}
EOF
echo "  ✅ version.json"

# 8. postinstallスクリプト（.envは既に含まれている前提）
cat > "$SCRIPTS_DIR/postinstall" << 'EOF'
#!/bin/bash
# ROI Agent - Post Installation Script

set -e

echo "🚀 Setting up ROI Agent..."

CURRENT_USER=$(stat -f "%Su" /dev/console)
USER_HOME=$(eval echo ~$CURRENT_USER)
USER_ID=$(id -u "$CURRENT_USER")

# LaunchAgent設定
LAUNCH_AGENTS_DIR="$USER_HOME/Library/LaunchAgents"
mkdir -p "$LAUNCH_AGENTS_DIR"
chown "$CURRENT_USER:staff" "$LAUNCH_AGENTS_DIR"

PLIST_SRC="/Applications/ROI Agent/Resources/com.roiagent.plist"
PLIST_DST="$LAUNCH_AGENTS_DIR/com.roiagent.plist"

cp "$PLIST_SRC" "$PLIST_DST"
chown "$CURRENT_USER:staff" "$PLIST_DST"
chmod 644 "$PLIST_DST"

# ログディレクトリ作成
LOG_DIR="$USER_HOME/.roiagent/logs"
DATA_DIR="$USER_HOME/.roiagent/data"

mkdir -p "$LOG_DIR" "$DATA_DIR"
chown -R "$CURRENT_USER:staff" "$USER_HOME/.roiagent"

# .envファイルの存在確認
ENV_FILE="/Applications/ROI Agent/Resources/.env"
if [ ! -f "$ENV_FILE" ]; then
    echo "⚠️  Warning: Configuration file (.env) not found"
    echo "This PKG should be downloaded from ROI Dashboard"
    # テンプレートから作成（フォールバック）
    cp "/Applications/ROI Agent/Resources/.env.template" "$ENV_FILE"
fi

chmod 600 "$ENV_FILE"
chown "$CURRENT_USER:staff" "$ENV_FILE"

echo "✅ Configuration verified"

# LaunchAgentをロード
sudo -u "$CURRENT_USER" launchctl bootstrap "gui/$USER_ID" "$PLIST_DST" 2>/dev/null || true
sudo -u "$CURRENT_USER" launchctl enable "gui/$USER_ID/com.roiagent" 2>/dev/null || true
sudo -u "$CURRENT_USER" launchctl kickstart -k "gui/$USER_ID/com.roiagent" 2>/dev/null || true

echo ""
echo "✅ ROI Agent installed and started!"
echo "📊 Check logs: tail -f $LOG_DIR/agent.log"
echo ""

exit 0
EOF
chmod +x "$SCRIPTS_DIR/postinstall"
echo "  ✅ postinstall script"

echo ""
echo "📦 Building PKG template..."

PKG_ID="com.roiagent.pkg"
PKG_VERSION="$VERSION"
PKG_NAME="ROI-Agent-macOS-${ARCH_LABEL}-${VERSION}.pkg"
PKG_OUTPUT="$BUILD_DIR/$PKG_NAME"

# PKGビルド
pkgbuild \
    --root "$PAYLOAD_DIR" \
    --scripts "$SCRIPTS_DIR" \
    --identifier "$PKG_ID" \
    --version "$PKG_VERSION" \
    --install-location "/" \
    "$PKG_OUTPUT"

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ PKG template created successfully!"
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
    
    # クリーンアップ
    rm -rf "$PAYLOAD_DIR" "$SCRIPTS_DIR"
    
    echo "🎉 Build completed!"
else
    echo "❌ Failed to create PKG"
    exit 1
fi
