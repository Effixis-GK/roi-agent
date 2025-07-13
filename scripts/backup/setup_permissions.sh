#!/bin/bash

# ROI Agent Enhanced - Final Setup and Permissions
# 最終セットアップと権限設定

set -e

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"

echo "🔧 ROI Agent Enhanced - 最終セットアップ"
echo "========================================="

cd "$BASE_DIR"

# 実行権限付与
echo "1. スクリプトファイルに実行権限を付与中..."
chmod +x build_enhanced.sh
chmod +x quick_setup_enhanced.sh
chmod +x network_debug_tools.py

if [ -f "build_agent.sh" ]; then
    chmod +x build_agent.sh
fi

if [ -f "start_web.sh" ]; then
    chmod +x start_web.sh
fi

if [ -f "dev_tools.sh" ]; then
    chmod +x dev_tools.sh
fi

echo "✅ スクリプト権限設定完了"

# Go module初期化
echo "2. Go module設定確認中..."
cd "$BASE_DIR/agent"

if [ ! -f "go.mod" ]; then
    echo "   go.mod を作成中..."
    go mod init roi-agent
fi

go mod tidy
echo "✅ Go module設定完了"

# Python仮想環境確認
echo "3. Python環境確認中..."
cd "$BASE_DIR/web"

if [ ! -d "venv" ]; then
    echo "   Python仮想環境を作成中..."
    python3 -m venv venv
fi

source venv/bin/activate
pip install --upgrade pip >/dev/null 2>&1
pip install flask requests >/dev/null 2>&1
echo "✅ Python環境設定完了"

# ディレクトリ構造確認
echo "4. ディレクトリ構造確認中..."
cd "$BASE_DIR"

mkdir -p build
mkdir -p data
mkdir -p logs
mkdir -p config

# ユーザーデータディレクトリ作成
USER_DATA_DIR="$HOME/.roiagent"
mkdir -p "$USER_DATA_DIR/data"
mkdir -p "$USER_DATA_DIR/logs"

echo "✅ ディレクトリ構造確認完了"

# 設定ファイル確認
echo "5. 設定ファイル確認中..."
if [ ! -f "config/config.yaml" ]; then
    echo "   デフォルト設定ファイルを作成中..."
    cat > config/config.yaml << 'EOF'
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
fi

echo "✅ 設定ファイル確認完了"

# テストデータ生成
echo "6. 初期テストデータ生成中..."
cd "$BASE_DIR"
python3 network_debug_tools.py testdata >/dev/null 2>&1
echo "✅ テストデータ生成完了"

# システム要件チェック
echo "7. システム要件最終チェック..."

# Go確認
if command -v go >/dev/null 2>&1; then
    echo "   ✅ Go: $(go version | cut -d' ' -f3)"
else
    echo "   ❌ Go言語が見つかりません"
    echo "      インストール: brew install go"
fi

# Python確認
if command -v python3 >/dev/null 2>&1; then
    echo "   ✅ Python: $(python3 --version | cut -d' ' -f2)"
else
    echo "   ❌ Python3が見つかりません"
fi

# macOS バージョン確認
echo "   ✅ macOS: $(sw_vers -productVersion)"

echo ""
echo "🎉 ROI Agent Enhanced セットアップ完了！"
echo "========================================="
echo ""
echo "📁 プロジェクト構成:"
echo "   ├── agent/               Go監視エージェント（ネットワーク対応）"
echo "   ├── web/                 Flask Web UI（拡張版）"
echo "   ├── build/               ビルド出力"
echo "   ├── config/              設定ファイル"
echo "   ├── data/                テストデータ"
echo "   └── logs/                ログファイル"
echo ""
echo "🚀 次のステップ:"
echo ""
echo "1. ワンコマンドビルド＆インストール:"
echo "   ./quick_setup_enhanced.sh"
echo ""
echo "2. 手動ビルド:"
echo "   ./build_enhanced.sh"
echo ""
echo "3. ネットワーク機能テスト:"
echo "   python3 network_debug_tools.py full"
echo ""
echo "4. 完成したアプリの場所:"
echo "   /Applications/ROI Agent Enhanced.app"
echo ""
echo "📊 新機能 - ネットワーク監視:"
echo "   ✅ HTTP/HTTPS通信監視"
echo "   ✅ ドメイン別アクセス時間"
echo "   ✅ 帯域幅使用量追跡"
echo "   ✅ アプリ統合分析"
echo "   ✅ リアルタイムダッシュボード"
echo ""
echo "🌐 ダッシュボード: http://localhost:5002"
echo "   - アプリケーションタブ: 従来の使用時間監視"
echo "   - ネットワークタブ: 新しい通信監視機能"
echo "   - 統合ビュー: アプリ+ネットワーク総合分析"
echo ""
echo "Ready to monitor your productivity! 🎯"
