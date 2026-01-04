#!/bin/bash
# ROI Agent - Post Installation Script
# Reads configuration from installer environment variables

set -e

echo "🚀 Setting up ROI Agent..."

CURRENT_USER=$(stat -f "%Su" /dev/console)
USER_HOME=$(eval echo ~$CURRENT_USER)
USER_ID=$(id -u "$CURRENT_USER")

echo "Installing for user: $CURRENT_USER (UID: $USER_ID)"

# LaunchAgentsディレクトリを作成
LAUNCH_AGENTS_DIR="$USER_HOME/Library/LaunchAgents"
mkdir -p "$LAUNCH_AGENTS_DIR"
chown "$CURRENT_USER:staff" "$LAUNCH_AGENTS_DIR"

# plistファイルをコピー
PLIST_SRC="/Applications/ROI Agent/Resources/com.roiagent.plist"
PLIST_DST="$LAUNCH_AGENTS_DIR/com.roiagent.plist"

cp "$PLIST_SRC" "$PLIST_DST"
chown "$CURRENT_USER:staff" "$PLIST_DST"
chmod 644 "$PLIST_DST"

echo "✅ LaunchAgent plist installed"

# ログディレクトリを作成
LOG_DIR="$USER_HOME/.roiagent/logs"
DATA_DIR="$USER_HOME/.roiagent/data"

mkdir -p "$LOG_DIR"
mkdir -p "$DATA_DIR"
chown -R "$CURRENT_USER:staff" "$USER_HOME/.roiagent"

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

# Set readable permissions so data-sender can read when run as user
chmod 644 "$ENV_FILE"
chown "$CURRENT_USER:staff" "$ENV_FILE"

echo "✅ Configuration file created (readable by user)"

# LaunchAgentをロード
echo "Loading LaunchAgent..."
sudo -u "$CURRENT_USER" launchctl bootstrap "gui/$USER_ID" "$PLIST_DST" 2>/dev/null || true
sudo -u "$CURRENT_USER" launchctl enable "gui/$USER_ID/com.roiagent" 2>/dev/null || true
sudo -u "$CURRENT_USER" launchctl kickstart -k "gui/$USER_ID/com.roiagent" 2>/dev/null || true

echo ""
echo "✅ ROI Agent installed and started!"
echo ""
echo "📊 Check logs: tail -f $LOG_DIR/agent.log"
echo ""

exit 0
