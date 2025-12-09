#!/bin/bash

# ROI Agent - macOS PKG Template Builder
# Creates PKG without org-specific config (.env will be injected on download)

set -e

echo "📦 ROI Agent - PKG Template Builder"
echo "========================================"

# 設定
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build/pkg"
VERSION="${1:-1.2.8}"
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

# 3. .env.template
cat > "$RESOURCES_DIR/.env.template" << 'EOF'
# ROI Agent Configuration
ROI_AGENT_BASE_URL=https://test-607617540267.asia-northeast1.run.app/api/v1/device
ROI_AGENT_API_KEY=YOUR_API_KEY_HERE
ROI_AGENT_INTERVAL_MINUTES=10
ROI_AGENT_ENABLED=true
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
  /var/log/roiagent/roiagent.log
  /var/log/roiagent/roiagent-error.log

SUPPORT:
  https://roi-dashboard-607617540267.asia-northeast1.run.app
EOF
echo "  ✅ README.txt"

# 4.5 VERSION file (used by data-sender to report agent version)
echo "$VERSION" > "$RESOURCES_DIR/VERSION"
echo "  ✅ VERSION ($VERSION)"

# 5. LaunchDaemon plist (Resourcesに配置してpostinstallでコピー)
cat > "$RESOURCES_DIR/com.roiagent.daemon.plist" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.roiagent.daemon</string>
    
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
    <string>/var/log/roiagent/roiagent.log</string>
    
    <key>StandardErrorPath</key>
    <string>/var/log/roiagent/roiagent-error.log</string>
    
    <key>WorkingDirectory</key>
    <string>/Applications/ROI Agent</string>
    
    <key>EnvironmentVariables</key>
    <dict>
        <key>HOME</key>
        <string>/var/root</string>
    </dict>
</dict>
</plist>
EOF
echo "  ✅ com.roiagent.daemon.plist"

# 6. アンインストールスクリプト
cat > "$BIN_DIR/uninstall.sh" << 'EOF'
#!/bin/bash
echo "🗑️  Uninstalling ROI Agent..."

# LaunchDaemonを停止・削除
launchctl bootout system/com.roiagent.daemon 2>/dev/null || true
rm -f /Library/LaunchDaemons/com.roiagent.daemon.plist

# プロセスを強制終了
pkill -f "roi-agent" 2>/dev/null || true
pkill -f "data-sender" 2>/dev/null || true

# アプリケーションを削除
rm -rf "/Applications/ROI Agent"

echo ""
read -p "Remove system data (/var/lib/roiagent, /var/root/.roiagent)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf /var/lib/roiagent
    rm -rf /var/root/.roiagent
    echo "  ✅ System data removed"
fi

echo ""
read -p "Remove logs (/var/log/roiagent/)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf /var/log/roiagent
    echo "  ✅ Logs removed"
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
  "format": "PKG_DAEMON",
  "requires_root": true
}
EOF
echo "  ✅ version.json"

# 8. postinstallスクリプト
cat > "$SCRIPTS_DIR/postinstall" << 'EOF'
#!/bin/bash
# ROI Agent - Post Installation Script (LaunchDaemon version)

set -e

echo "🚀 Setting up ROI Agent (system-wide)..."

# 共有設定ディレクトリ作成 (roi-agent と data-sender の両方がアクセス)
mkdir -p /var/lib/roiagent
chmod 755 /var/lib/roiagent

# ログディレクトリとデータディレクトリ作成 (レガシー互換性)
mkdir -p /var/root/.roiagent/logs
mkdir -p /var/root/.roiagent/data
mkdir -p /var/root/.roiagent/transmission

# システムログディレクトリ作成
mkdir -p /var/log/roiagent
touch /var/log/roiagent/roiagent.log
touch /var/log/roiagent/roiagent-error.log
chmod 755 /var/log/roiagent
chmod 644 /var/log/roiagent/*.log

# .envファイルの確認
ENV_FILE="/Applications/ROI Agent/Resources/.env"
if [ ! -f "$ENV_FILE" ]; then
    echo "⚠️  Warning: Configuration file (.env) not found"
    echo "This PKG should be downloaded from ROI Dashboard"
    cp "/Applications/ROI Agent/Resources/.env.template" "$ENV_FILE"
fi

# Set readable permissions so data-sender can read when run as user
chmod 644 "$ENV_FILE"
chown root:staff "$ENV_FILE"

# Copy to shared config directory for user-level access
SHARED_ENV_FILE="/var/lib/roiagent/.env"
cp "$ENV_FILE" "$SHARED_ENV_FILE"
chmod 644 "$SHARED_ENV_FILE"

echo "✅ Configuration verified (readable by all users)"

# LaunchDaemonのplistファイルをコピー
SOURCE_PLIST="/Applications/ROI Agent/Resources/com.roiagent.daemon.plist"
PLIST_FILE="/Library/LaunchDaemons/com.roiagent.daemon.plist"

if [ ! -f "$SOURCE_PLIST" ]; then
    echo "❌ Error: Source plist not found at $SOURCE_PLIST"
    echo "PKG installation may have failed"
    exit 1
fi

echo "✅ Source plist found, copying to LaunchDaemons..."

# plistをコピー
cp "$SOURCE_PLIST" "$PLIST_FILE"

# plistファイルのパーミッション設定
chown root:wheel "$PLIST_FILE"
chmod 644 "$PLIST_FILE"

echo "✅ LaunchDaemon plist installed and configured"

# 既存のLaunchDaemonを停止（エラーを無視）
launchctl bootout system/com.roiagent.daemon 2>/dev/null || true

# LaunchDaemonをロード（macOS Big Sur以降対応）
if launchctl bootstrap system "$PLIST_FILE" 2>/dev/null; then
    echo "✅ LaunchDaemon loaded (bootstrap)"
elif launchctl load "$PLIST_FILE" 2>/dev/null; then
    echo "✅ LaunchDaemon loaded (legacy load)"
else
    echo "⚠️  Warning: Could not load LaunchDaemon, trying manual start..."
    launchctl start com.roiagent.daemon 2>/dev/null || true
fi

# LaunchDaemonを有効化
launchctl enable system/com.roiagent.daemon 2>/dev/null || true

# サービスを起動
launchctl kickstart -k system/com.roiagent.daemon 2>/dev/null || true

echo ""
echo "✅ ROI Agent installed and started!"
echo "📊 Check logs: tail -f /var/log/roiagent/roiagent.log"
echo "🔒 Running with root privileges for DNS monitoring"
echo ""
echo "🔍 Verify installation:"
echo "   sudo launchctl list | grep roiagent"
echo "   ps aux | grep roi-agent"
echo ""

exit 0
EOF
chmod +x "$SCRIPTS_DIR/postinstall"
echo "  ✅ postinstall script"

echo ""
echo "📦 Building PKG..."

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
    echo "✅ PKG created successfully!"
    echo ""
    echo "📊 Package Information:"
    echo "   File: $PKG_NAME"
    echo "   Size: $(du -h "$PKG_OUTPUT" | cut -f1)"
    echo ""
    
    # SHA256計算
    SHA256=$(shasum -a 256 "$PKG_OUTPUT" | awk '{print $1}')
    echo "   SHA256: $SHA256"
    echo "$SHA256" > "$PKG_OUTPUT.sha256"
    echo ""
    
    # クリーンアップ
    rm -rf "$PAYLOAD_DIR" "$SCRIPTS_DIR"
    
    echo "🎉 Build completed!"
    echo "⚠️  Note: This PKG installs ROI Agent as a LaunchDaemon (system-wide, runs as root)"
else
    echo "❌ Failed to create PKG"
    exit 1
fi
