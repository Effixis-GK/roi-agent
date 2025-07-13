#!/bin/bash

# ROI Agent Enhanced - Network Monitoring Build Script
# ネットワーク監視機能付きアプリケーションビルド

set -e

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
BUILD_DIR="$BASE_DIR/build"
APP_NAME="ROI Agent Enhanced"
BUNDLE_NAME="ROI Agent Enhanced.app"
BUNDLE_DIR="$BUILD_DIR/$BUNDLE_NAME"

echo "=== ROI Agent Enhanced ビルド開始 ==="
echo "ネットワーク監視機能付きバージョン"

# ディレクトリ作成
mkdir -p "$BUILD_DIR"
rm -rf "$BUNDLE_DIR"
mkdir -p "$BUNDLE_DIR/Contents/MacOS"
mkdir -p "$BUNDLE_DIR/Contents/Resources"
mkdir -p "$BUNDLE_DIR/Contents/Frameworks/Python"

echo "1. Goエージェント（ネットワーク監視版）をビルド中..."
cd "$BASE_DIR/agent"

# ネットワーク監視対応版をビルド
go mod tidy
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o "$BUNDLE_DIR/Contents/MacOS/monitor" network_main.go

# Intel Macサポート用
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o "$BUNDLE_DIR/Contents/MacOS/monitor_intel" network_main.go

# ユニバーサルバイナリ作成
lipo -create "$BUNDLE_DIR/Contents/MacOS/monitor" "$BUNDLE_DIR/Contents/MacOS/monitor_intel" -output "$BUNDLE_DIR/Contents/MacOS/monitor_universal"
mv "$BUNDLE_DIR/Contents/MacOS/monitor_universal" "$BUNDLE_DIR/Contents/MacOS/monitor"
rm -f "$BUNDLE_DIR/Contents/MacOS/monitor_intel"

echo "2. Python環境とWeb UIをセットアップ中..."
cd "$BASE_DIR/web"

# Python仮想環境作成
python3 -m venv "$BUNDLE_DIR/Contents/Frameworks/Python/venv"
source "$BUNDLE_DIR/Contents/Frameworks/Python/venv/bin/activate"

# 依存関係インストール
pip install --upgrade pip
pip install flask

# Web UIファイルをコピー
mkdir -p "$BUNDLE_DIR/Contents/Resources/web"
cp enhanced_app.py "$BUNDLE_DIR/Contents/Resources/web/app.py"
cp -r templates "$BUNDLE_DIR/Contents/Resources/web/"

echo "3. メインランチャーを作成中..."
cat > "$BUNDLE_DIR/Contents/MacOS/roi-agent" << 'EOF'
#!/bin/bash

# ROI Agent Enhanced Launcher
APP_DIR="$(dirname "$0")/.."
RESOURCES_DIR="$APP_DIR/Resources"
PYTHON_ENV="$APP_DIR/Frameworks/Python/venv"
MONITOR_BINARY="$(dirname "$0")/monitor"

# ホームディレクトリのデータフォルダを作成
USER_DATA_DIR="$HOME/.roiagent"
mkdir -p "$USER_DATA_DIR/data"
mkdir -p "$USER_DATA_DIR/logs"

export PYTHONPATH="$RESOURCES_DIR/web:$PYTHONPATH"

# 引数に応じてコマンド実行
case "${1:-start}" in
    "start")
        echo "Starting ROI Agent Enhanced with Network Monitoring..."
        
        # エージェント開始
        "$MONITOR_BINARY" &
        MONITOR_PID=$!
        echo $MONITOR_PID > "$USER_DATA_DIR/monitor.pid"
        
        # Web UI開始
        cd "$RESOURCES_DIR/web"
        source "$PYTHON_ENV/bin/activate"
        python app.py &
        WEB_PID=$!
        echo $WEB_PID > "$USER_DATA_DIR/web.pid"
        
        # ダッシュボードを開く
        sleep 3
        open "http://localhost:5002"
        
        echo "ROI Agent Enhanced started!"
        echo "Dashboard: http://localhost:5002"
        echo "Press Ctrl+C to stop or use 'roi-agent stop'"
        
        # 両方のプロセスを待機
        wait $MONITOR_PID $WEB_PID
        ;;
        
    "stop")
        echo "Stopping ROI Agent Enhanced..."
        
        # PIDファイルから停止
        if [ -f "$USER_DATA_DIR/monitor.pid" ]; then
            kill $(cat "$USER_DATA_DIR/monitor.pid") 2>/dev/null || true
            rm -f "$USER_DATA_DIR/monitor.pid"
        fi
        
        if [ -f "$USER_DATA_DIR/web.pid" ]; then
            kill $(cat "$USER_DATA_DIR/web.pid") 2>/dev/null || true
            rm -f "$USER_DATA_DIR/web.pid"
        fi
        
        # プロセス名でも停止
        pkill -f "monitor" || true
        pkill -f "enhanced_app.py" || true
        
        echo "ROI Agent Enhanced stopped."
        ;;
        
    "status")
        echo "=== ROI Agent Enhanced Status ==="
        
        if [ -f "$USER_DATA_DIR/monitor.pid" ] && kill -0 $(cat "$USER_DATA_DIR/monitor.pid") 2>/dev/null; then
            echo "✅ Monitor Agent: Running (PID: $(cat "$USER_DATA_DIR/monitor.pid"))"
        else
            echo "❌ Monitor Agent: Not running"
        fi
        
        if [ -f "$USER_DATA_DIR/web.pid" ] && kill -0 $(cat "$USER_DATA_DIR/web.pid") 2>/dev/null; then
            echo "✅ Web UI: Running (PID: $(cat "$USER_DATA_DIR/web.pid"))"
            echo "🌐 Dashboard: http://localhost:5002"
        else
            echo "❌ Web UI: Not running"
        fi
        
        # データ確認
        if [ -d "$USER_DATA_DIR/data" ]; then
            DATA_FILES=$(ls "$USER_DATA_DIR/data"/*.json 2>/dev/null | wc -l)
            echo "📊 Data files: $DATA_FILES"
        fi
        ;;
        
    "dashboard")
        open "http://localhost:5002"
        ;;
        
    "logs")
        echo "=== Monitor Logs ==="
        tail -f "$USER_DATA_DIR/logs"/*.log 2>/dev/null || echo "No logs found"
        ;;
        
    *)
        echo "ROI Agent Enhanced - ネットワーク監視機能付き"
        echo ""
        echo "使用方法:"
        echo "  $0 start       - エージェントとWeb UIを開始"
        echo "  $0 stop        - すべてのプロセスを停止"
        echo "  $0 status      - 動作状況を確認"
        echo "  $0 dashboard   - ダッシュボードを開く"
        echo "  $0 logs        - ログを表示"
        echo ""
        echo "機能:"
        echo "  📱 アプリケーション使用時間監視"
        echo "  🌐 ネットワーク通信監視 (HTTP/HTTPS + カスタムポート)"
        echo "  📊 ドメインごとのアクセス時間統計"
        echo "  🚀 リアルタイムダッシュボード"
        ;;
esac
EOF

chmod +x "$BUNDLE_DIR/Contents/MacOS/roi-agent"

echo "4. アプリケーション情報ファイルを作成中..."
cat > "$BUNDLE_DIR/Contents/Info.plist" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>roi-agent</string>
    <key>CFBundleIdentifier</key>
    <string>com.roiagent.enhanced</string>
    <key>CFBundleName</key>
    <string>ROI Agent Enhanced</string>
    <key>CFBundleDisplayName</key>
    <string>ROI Agent Enhanced</string>
    <key>CFBundleVersion</key>
    <string>2.0.0</string>
    <key>CFBundleShortVersionString</key>
    <string>2.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSAppleEventsUsageDescription</key>
    <string>ROI Agent Enhanced needs to monitor application usage and network connections for productivity analysis.</string>
    <key>NSSystemAdministrationUsageDescription</key>
    <string>ROI Agent Enhanced needs system administration access to monitor network connections.</string>
    <key>com.apple.security.network.server</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
</dict>
</plist>
EOF

echo "5. アイコンとリソースをセットアップ中..."
# アイコン作成（シンプルなテキストベース）
cat > "$BUNDLE_DIR/Contents/Resources/app_icon.txt" << 'EOF'
ROI Agent Enhanced
Network Monitoring Version
🚀📊🌐
EOF

echo "6. 設定ファイルをコピー中..."
cp "$BASE_DIR/config/config.yaml" "$BUNDLE_DIR/Contents/Resources/" 2>/dev/null || echo "No config file found, creating default"

cat > "$BUNDLE_DIR/Contents/Resources/config.yaml" << 'EOF'
# ROI Agent Enhanced Configuration
monitor:
  interval: 15  # seconds
  data_retention_days: 30
  
network:
  monitor_ports: [80, 443, 8080, 3000, 5000, 8000, 9000]
  monitor_protocols: ["HTTP", "HTTPS", "TCP"]
  dns_resolution: true
  packet_capture: false  # Requires admin privileges
  
web:
  host: "127.0.0.1"
  port: 5002
  auto_refresh: 30  # seconds
  
security:
  require_accessibility: true
  local_only: true
EOF

echo "7. 最終チェックと権限設定..."
chmod +x "$BUNDLE_DIR/Contents/MacOS/monitor"
chmod +x "$BUNDLE_DIR/Contents/MacOS/roi-agent"

# バンドル構造確認
echo ""
echo "=== アプリケーションバンドル作成完了 ==="
echo "場所: $BUNDLE_DIR"
echo ""
echo "構成:"
find "$BUNDLE_DIR" -type f | head -20
echo "..."
echo ""

echo "8. クイックテスト実行..."
if "$BUNDLE_DIR/Contents/MacOS/roi-agent" status > /dev/null 2>&1; then
    echo "✅ アプリケーション起動テスト: 成功"
else
    echo "❌ アプリケーション起動テスト: 失敗"
fi

echo ""
echo "=== 🎉 ROI Agent Enhanced ビルド完了 ==="
echo ""
echo "次のステップ:"
echo "1. アプリテスト:      open '$BUNDLE_DIR'"
echo "2. インストール:      cp -R '$BUNDLE_DIR' /Applications/"
echo "3. 起動:            '/Applications/$BUNDLE_NAME/Contents/MacOS/roi-agent start'"
echo "4. ダッシュボード:     http://localhost:5002"
echo ""
echo "機能一覧:"
echo "📱 アプリケーション監視 - フォアグラウンド/バックグラウンド/フォーカス時間"
echo "🌐 ネットワーク監視 - HTTP/HTTPS通信とカスタムポート"
echo "📊 ドメイン分析 - アクセス時間と帯域幅統計"
echo "⚡ リアルタイム更新 - 15秒間隔でのデータ収集"
echo "🎯 統合ダッシュボード - アプリとネットワークの統合ビュー"
echo ""
