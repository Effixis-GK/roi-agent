#!/bin/bash
# ROI Agent - Post Installation Script
# Installs as LaunchDaemon (root) for tcpdump access

set -e

echo "🚀 Setting up ROI Agent..."

# LaunchDaemonsディレクトリ（システム全体で動作）
LAUNCH_DAEMONS_DIR="/Library/LaunchDaemons"
mkdir -p "$LAUNCH_DAEMONS_DIR"

# plistファイルをLaunchDaemonsにコピー
PLIST_SRC="/Applications/ROI Agent/Resources/com.roiagent.plist"
PLIST_DST="$LAUNCH_DAEMONS_DIR/com.roiagent.plist"

cp "$PLIST_SRC" "$PLIST_DST"
chown root:wheel "$PLIST_DST"
chmod 644 "$PLIST_DST"

echo "✅ LaunchDaemon plist installed"

# ログディレクトリを作成（システム共有）
LOG_DIR="/Users/Shared/.roiagent/logs"
DATA_DIR="/Users/Shared/.roiagent/data"

mkdir -p "$LOG_DIR"
mkdir -p "$DATA_DIR"
chown -R root:wheel "/Users/Shared/.roiagent"
chmod -R 755 "/Users/Shared/.roiagent"

echo "✅ Log directory created"

# .envファイルを生成（環境変数から）
ENV_FILE="/Applications/ROI Agent/Resources/.env"

# インストーラーから渡される環境変数を使用
API_KEY="${ROI_AGENT_API_KEY:-YOUR_API_KEY_HERE}"
BASE_URL="${ROI_AGENT_BASE_URL:-https://roi-service-607617540267.asia-northeast1.run.app/api/v1/device}"

cat > "$ENV_FILE" << EOF
# ROI Agent Configuration
# Auto-configured during installation

ROI_AGENT_BASE_URL=$BASE_URL
ROI_AGENT_API_KEY=$API_KEY
ROI_AGENT_INTERVAL_MINUTES=10
EOF

chmod 600 "$ENV_FILE"
chown root:wheel "$ENV_FILE"

echo "✅ Configuration file created"

# LaunchDaemonをロード（root権限で動作）
echo "Loading LaunchDaemon as root..."
launchctl bootstrap system "$PLIST_DST" 2>/dev/null || true
launchctl enable system/com.roiagent 2>/dev/null || true
launchctl kickstart -k system/com.roiagent 2>/dev/null || true

echo ""
echo "✅ ROI Agent installed and started as system service!"
echo "   Running with root permissions for tcpdump DNS monitoring"
echo ""
echo "📊 Check logs: tail -f $LOG_DIR/agent.log"
echo "🛑 Stop service: sudo launchctl stop system/com.roiagent"
echo "🚀 Start service: sudo launchctl start system/com.roiagent"
echo ""

exit 0
