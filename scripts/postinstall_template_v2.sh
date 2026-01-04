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

# 共有設定ディレクトリを作成（roi-agent と data-sender の両方がアクセス）
SHARED_CONFIG_DIR="/var/lib/roiagent"
mkdir -p "$SHARED_CONFIG_DIR"
chmod 755 "$SHARED_CONFIG_DIR"

echo "✅ Shared config directory created"

# ログディレクトリを作成（システム共有）
LOG_DIR="/var/log/roiagent"
DATA_DIR="/var/root/.roiagent/data"

mkdir -p "$LOG_DIR"
mkdir -p "$DATA_DIR"
mkdir -p "/var/root/.roiagent/logs"
mkdir -p "/var/root/.roiagent/transmission"
chmod 755 "$LOG_DIR"

echo "✅ Log and data directories created"

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

# Set readable permissions so data-sender can read when run as user
chmod 644 "$ENV_FILE"
chown root:staff "$ENV_FILE"

echo "✅ Configuration file created (readable by all users)"

# Also create a config in shared directory for user-level access
SHARED_ENV_FILE="$SHARED_CONFIG_DIR/.env"
cp "$ENV_FILE" "$SHARED_ENV_FILE"
chmod 644 "$SHARED_ENV_FILE"

echo "✅ Shared configuration file created"

# LaunchDaemonをロード（root権限で動作）
echo "Loading LaunchDaemon as root..."
launchctl bootstrap system "$PLIST_DST" 2>/dev/null || true
launchctl enable system/com.roiagent 2>/dev/null || true
launchctl kickstart -k system/com.roiagent 2>/dev/null || true

echo ""
echo "✅ ROI Agent installed and started as system service!"
echo "   Running with root permissions for tcpdump DNS monitoring"
echo ""
echo "📊 Check logs: tail -f /var/log/roiagent/roiagent.log"
echo "🛑 Stop service: sudo launchctl stop system/com.roiagent"
echo "🚀 Start service: sudo launchctl start system/com.roiagent"
echo ""

exit 0
