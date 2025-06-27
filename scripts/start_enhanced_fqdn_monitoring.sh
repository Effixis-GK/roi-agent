#!/bin/bash

# ROI Agent Enhanced - Real FQDN Network Monitoring
# 実際のFQDN解決とパケットキャプチャを使用したネットワーク監視

set -e

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
cd "$BASE_DIR"

echo "🌐 ROI Agent Enhanced - 実FQDN監視モード"
echo "=========================================="

# 1. 環境チェック
echo "1. 拡張ネットワーク監視環境チェック..."

if ! command -v go &> /dev/null; then
    echo "❌ Go言語が見つかりません"
    exit 1
fi

if ! command -v python3 &> /dev/null; then
    echo "❌ Python3が見つかりません"
    exit 1
fi

echo "✅ 環境確認完了"

# 2. 拡張エージェントビルド
echo "2. 拡張ネットワーク監視エージェントをビルド中..."
cd "$BASE_DIR/agent"

if [ ! -f "go.mod" ]; then
    go mod init roi-agent-enhanced
fi

go mod tidy

# 拡張版をビルド
echo "   enhanced_network_main.go をビルド中..."
go build -o monitor_enhanced enhanced_network_main.go

if [ ! -f "monitor_enhanced" ]; then
    echo "❌ 拡張エージェントビルド失敗"
    exit 1
fi

echo "✅ 拡張エージェントビルド完了"

# 3. ネットワーク監視機能テスト
echo "3. ネットワーク監視機能事前テスト..."
cd "$BASE_DIR"

echo "   FQDN解決テスト:"
python3 network_fqdn_debug.py fqdn

echo "   現在の接続テスト:"
python3 network_fqdn_debug.py connections

# 4. Python環境セットアップ
echo "4. Python Web UI 環境セットアップ中..."
cd "$BASE_DIR/web"

if [ ! -d "venv" ]; then
    python3 -m venv venv
fi

source venv/bin/activate
pip install --upgrade pip > /dev/null 2>&1
pip install flask requests > /dev/null 2>&1

echo "✅ Python環境セットアップ完了"

# 5. データディレクトリ準備
echo "5. 実データディレクトリを準備中..."
USER_DATA_DIR="$HOME/.roiagent"
mkdir -p "$USER_DATA_DIR/data"
mkdir -p "$USER_DATA_DIR/logs"

# 既存データクリア
echo "   既存データをクリア中..."
rm -f "$USER_DATA_DIR/data"/combined_*.json

echo "✅ データディレクトリ準備完了"

# 6. 既存プロセス停止
echo "6. 既存プロセス確認・停止中..."

pkill -f "monitor" 2>/dev/null || true
pkill -f "enhanced_app.py" 2>/dev/null || true

sleep 2
echo "✅ プロセス確認完了"

# 7. 権限確認
echo "7. 拡張ネットワーク監視権限確認..."
cd "$BASE_DIR/agent"

# FQDN解決テスト
echo "   FQDN解決機能テスト:"
./monitor_enhanced test-fqdn

# アクセシビリティ権限確認
if ./monitor_enhanced check-permissions 2>/dev/null | grep -q "OK"; then
    echo "✅ アクセシビリティ権限: OK"
else
    echo "⚠️  アクセシビリティ権限が必要です"
    echo ""
    echo "📋 権限設定手順:"
    echo "   システム環境設定 > セキュリティとプライバシー > アクセシビリティ"
    echo "   ターミナル (または使用中のエディタ) を追加"
    echo ""
    read -p "権限設定後、Enterで続行..." -r
fi

# 8. 拡張監視モード起動
echo ""
echo "🚀 ROI Agent Enhanced - 実FQDN監視開始"
echo "======================================"
echo ""

# ログファイル準備
LOG_DIR="$USER_DATA_DIR/logs"
AGENT_LOG="$LOG_DIR/enhanced_agent_$(date +%Y%m%d_%H%M%S).log"
WEB_LOG="$LOG_DIR/enhanced_web_$(date +%Y%m%d_%H%M%S).log"

# 拡張エージェント起動
echo "📡 拡張ネットワーク監視エージェント起動中..."
cd "$BASE_DIR/agent"

echo "Starting ROI Agent Enhanced with REAL FQDN RESOLUTION" > "$AGENT_LOG"
echo "Features: Packet capture, FQDN resolution, Redirect following" >> "$AGENT_LOG"
echo "Timestamp: $(date)" >> "$AGENT_LOG"
echo "" >> "$AGENT_LOG"

nohup ./monitor_enhanced >> "$AGENT_LOG" 2>&1 &
AGENT_PID=$!
echo "   PID: $AGENT_PID"
echo "   ログ: $AGENT_LOG"

# 少し待機してFQDN解決開始を確認
sleep 8

# Web UI起動
echo "🌐 拡張Web UI起動中..."
cd "$BASE_DIR/web"
source venv/bin/activate

echo "Starting Enhanced Web UI with FQDN Network Data" > "$WEB_LOG"
echo "Timestamp: $(date)" >> "$WEB_LOG"
echo "" >> "$WEB_LOG"

nohup python enhanced_app.py >> "$WEB_LOG" 2>&1 &
WEB_PID=$!
echo "   PID: $WEB_PID"
echo "   ログ: $WEB_LOG"

# 起動完了まで待機
echo ""
echo "⏳ 拡張ネットワーク監視開始を待機中..."
sleep 10

# 実際のネットワーク監視確認
echo ""
echo "📊 拡張ネットワーク監視状況確認:"

# エージェント確認
if kill -0 $AGENT_PID 2>/dev/null; then
    echo "✅ 拡張監視エージェント: 動作中 (PID: $AGENT_PID)"
    
    # 実際にFQDN解決が行われているか確認
    sleep 5
    
    echo "   FQDN解決確認中..."
    if grep -q "Resolved" "$AGENT_LOG" 2>/dev/null; then
        echo "   ✅ FQDN解決: 動作中"
        grep "Resolved" "$AGENT_LOG" | tail -3 | sed 's/^/     /'
    else
        echo "   ⚠️  FQDN解決: まだ実行されていません（接続待機中）"
    fi
    
    TODAY=$(date +%Y-%m-%d)
    REAL_DATA_FILE="$USER_DATA_DIR/data/combined_$TODAY.json"
    
    sleep 3
    
    if [ -f "$REAL_DATA_FILE" ]; then
        echo "✅ 拡張データファイル: 作成済み ($REAL_DATA_FILE)"
        
        if command -v jq > /dev/null 2>&1; then
            APP_COUNT=$(jq '.apps | length' "$REAL_DATA_FILE" 2>/dev/null || echo "0")
            NET_COUNT=$(jq '.network | length' "$REAL_DATA_FILE" 2>/dev/null || echo "0")
            DNS_COUNT=$(jq '.dns_queries | length' "$REAL_DATA_FILE" 2>/dev/null || echo "0")
            UNIQUE_DOMAINS=$(jq '.network_total.unique_domains' "$REAL_DATA_FILE" 2>/dev/null || echo "0")
            
            echo "   収集中のアプリ数: $APP_COUNT"
            echo "   ネットワーク接続数: $NET_COUNT"
            echo "   DNSクエリ数: $DNS_COUNT"
            echo "   ユニークドメイン数: $UNIQUE_DOMAINS"
            
            if [ "$NET_COUNT" -gt 0 ]; then
                echo "   実際のFQDN例:"
                jq -r '.network | to_entries[] | "     " + .key + " (" + .value.app_name + ")"' "$REAL_DATA_FILE" 2>/dev/null | head -3 || echo "     データ解析中..."
            fi
        fi
    else
        echo "⚠️  拡張データファイル: 作成中（FQDN解決処理中）"
    fi
else
    echo "❌ 拡張監視エージェント: 起動失敗"
    echo "   ログ確認: tail -f $AGENT_LOG"
fi

# Web UI確認
if kill -0 $WEB_PID 2>/dev/null; then
    echo "✅ 拡張Web UI: 動作中 (PID: $WEB_PID)"
    
    sleep 3
    if curl -s http://localhost:5002/api/status > /dev/null; then
        echo "✅ 拡張HTTP接続: OK"
        
        # 拡張API確認
        STATUS_JSON=$(curl -s http://localhost:5002/api/status 2>/dev/null)
        if echo "$STATUS_JSON" | grep -q "unique_domains"; then
            UNIQUE_DOMAINS=$(echo "$STATUS_JSON" | jq '.unique_domains' 2>/dev/null || echo "0")
            DNS_CACHE_SIZE=$(echo "$STATUS_JSON" | jq '.dns_cache_size' 2>/dev/null || echo "0")
            echo "✅ 拡張機能API: 動作中"
            echo "   DNSキャッシュサイズ: $DNS_CACHE_SIZE"
            echo "   ユニークドメイン数: $UNIQUE_DOMAINS"
        else
            echo "⚠️  API応答: 基本版が動作中（拡張版ではない）"
        fi
    else
        echo "❌ HTTP接続: 失敗"
    fi
else
    echo "❌ 拡張Web UI: 起動失敗"
    echo "   ログ確認: tail -f $WEB_LOG"
fi

echo ""
echo "🎯 拡張FQDN監視モード起動完了!"
echo "================================"
echo ""
echo "🌐 拡張ダッシュボード:"
echo "   URL: http://localhost:5002"
echo "   機能: 実際のFQDN解決 + パケット解析 + リダイレクト追跡"
echo ""
echo "📡 拡張監視機能:"
echo "   ✅ 実際のFQDNを取得（IPアドレス→ドメイン名解決）"
echo "   ✅ HTTPリダイレクト追跡（最終アクセス先を特定）"
echo "   ✅ DNSクエリ監視"
echo "   ✅ アプリケーション別通信分析"
echo "   ✅ リアルタイムネットワーク接続状況"
echo ""
echo "🔍 拡張デバッグコマンド:"
echo "   FQDN解決確認:     python3 network_fqdn_debug.py fqdn"
echo "   現在の接続:       python3 network_fqdn_debug.py connections"
echo "   DNS監視:         python3 network_fqdn_debug.py dns"
echo "   リダイレクト:     python3 network_fqdn_debug.py redirects"
echo "   完全診断:        python3 network_fqdn_debug.py full"
echo ""
echo "📊 拡張API エンドポイント:"
echo "   状況確認:    curl http://localhost:5002/api/status"
echo "   ネットワーク: curl 'http://localhost:5002/api/data?type=network'"
echo "   統合データ:   curl 'http://localhost:5002/api/data?type=both'"
echo "   ドメイン分析: curl http://localhost:5002/api/network/domains"
echo ""
echo "📄 拡張ログファイル:"
echo "   エージェント: $AGENT_LOG"
echo "   Web UI: $WEB_LOG"
echo ""
echo "🛠️ 実時間監視コマンド:"
echo "   エージェントログ: tail -f $AGENT_LOG"
echo "   Web UIログ:     tail -f $WEB_LOG"
echo "   FQDN解決監視:    grep 'Resolved' $AGENT_LOG"
echo "   リアルタイム監視: python3 real_data_debug.py monitor"
echo ""
echo "🛑 停止方法:"
echo "   kill $AGENT_PID $WEB_PID"
echo "   または: ./stop_enhanced_monitoring.sh"
echo ""

# 拡張版停止スクリプト作成
cat > "$BASE_DIR/stop_enhanced_monitoring.sh" << EOF
#!/bin/bash
echo "🛑 ROI Agent Enhanced - 拡張FQDN監視停止"
echo "======================================="

echo "プロセス停止中..."
kill $AGENT_PID $WEB_PID 2>/dev/null || true
pkill -f "monitor_enhanced" || true
pkill -f "enhanced_app.py" || true

sleep 3
echo "✅ 停止完了"

echo ""
echo "📊 収集された拡張ネットワークデータ:"
echo "============================"
TODAY=\$(date +%Y-%m-%d)
REAL_DATA_FILE="$USER_DATA_DIR/data/combined_\$TODAY.json"

if [ -f "\$REAL_DATA_FILE" ]; then
    echo "拡張データファイル: \$REAL_DATA_FILE"
    
    if command -v jq > /dev/null 2>&1; then
        echo ""
        echo "📈 統計情報:"
        echo "  アプリ数: \$(jq '.apps | length' "\$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo "  ネットワーク接続数: \$(jq '.network | length' "\$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo "  DNSクエリ数: \$(jq '.dns_queries | length' "\$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo "  ユニークドメイン数: \$(jq '.network_total.unique_domains' "\$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo ""
        echo "🌐 実際に解決されたFQDN:"
        jq -r '.network | to_entries[] | select(.value.domain != .value.remote_ip) | "  " + .value.domain + " (" + .value.remote_ip + ")"' "\$REAL_DATA_FILE" 2>/dev/null | head -10 || echo "  データなし"
        echo ""
        echo "📡 DNSクエリ履歴:"
        jq -r '.dns_queries[] | "  " + .domain + " (" + .timestamp + ")"' "\$REAL_DATA_FILE" 2>/dev/null | tail -5 || echo "  データなし"
    fi
else
    echo "拡張データファイルが見つかりません"
fi

echo ""
echo "📄 最新ログ:"
echo "-----------"
echo "拡張エージェント:"
tail -5 "$AGENT_LOG" 2>/dev/null || echo "ログなし"
echo ""
echo "拡張Web UI:"
tail -5 "$WEB_LOG" 2>/dev/null || echo "ログなし"

echo ""
echo "🔍 FQDN解決状況:"
echo "---------------"
if [ -f "$AGENT_LOG" ]; then
    FQDN_COUNT=\$(grep -c "Resolved" "$AGENT_LOG" 2>/dev/null || echo "0")
    echo "  FQDN解決回数: \$FQDN_COUNT"
    if [ "\$FQDN_COUNT" -gt 0 ]; then
        echo "  最近の解決例:"
        grep "Resolved" "$AGENT_LOG" | tail -3 | sed 's/^/    /' || echo "    なし"
    fi
else
    echo "  ログファイルなし"
fi
EOF

chmod +x "$BASE_DIR/stop_enhanced_monitoring.sh"

# 使用方法の詳細表示
echo "💡 使用のヒント:"
echo "==============="
echo ""
echo "1. 実際のFQDN取得を確認するには:"
echo "   - ブラウザでいくつかのWebサイト（github.com, google.com等）を訪問"
echo "   - tail -f $AGENT_LOG でリアルタイム解決を確認"
echo ""
echo "2. リダイレクト追跡を確認するには:"
echo "   - http://github.com (httpsにリダイレクト) などにアクセス"
echo "   - ダッシュボードでリダイレクトチェーンを確認"
echo ""
echo "3. 高精度監視のためには:"
echo "   sudo ./start_enhanced_fqdn_monitoring.sh"
echo "   （管理者権限で実行するとより詳細な情報を取得可能）"
echo ""

# ブラウザを開く（オプション）
echo "🌐 拡張ダッシュボードを開きますか? (y/n): "
read -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "ブラウザで拡張ダッシュボードを開いています..."
    sleep 3
    open "http://localhost:5002"
    
    echo ""
    echo "📈 ダッシュボードの見方:"
    echo "  - ネットワークタブ: 実際のFQDNとリダイレクト情報"
    echo "  - 統合ビュー: アプリ使用とネットワーク通信の関連性"
    echo "  - ドメイン分析ボタン: 詳細な通信統計"
fi

echo ""
echo "🎉 拡張FQDN監視が開始されました!"
echo ""
echo "Happy Enhanced Network Monitoring! 🌐🎯✨"
