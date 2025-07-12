#!/bin/bash
echo "🛑 ROI Agent Enhanced - tcpdump DNS監視停止"
echo "======================================="

echo "プロセス停止中..."

# Stop tcpdump processes
sudo pkill -f "tcpdump.*port 53" || true

# Stop Go agent
pkill -f "main.go" || true

# Stop Web UI
pkill -f "enhanced_app.py" || true

sleep 3
echo "✅ 停止完了"

echo ""
echo "📊 収集されたDNS監視データ:"
echo "============================"
TODAY=$(date +%Y-%m-%d)
REAL_DATA_FILE="$HOME/.roiagent/data/combined_$TODAY.json"

if [ -f "$REAL_DATA_FILE" ]; then
    echo "データファイル: $REAL_DATA_FILE"
    
    if command -v jq > /dev/null 2>&1; then
        echo ""
        echo "📈 統計情報:"
        echo "  アクティブアプリ数: $(jq '[.apps[] | select(.is_active == true)] | length' "$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo "  アクティブネットワーク接続数: $(jq '[.network[] | select(.is_active == true)] | length' "$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo "  ユニークドメイン数: $(jq '.network_total.unique_domains' "$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo "  総接続時間: $(jq '.network_total.total_duration' "$REAL_DATA_FILE" 2>/dev/null || echo "0")秒"
        echo ""
        echo "🌐 検出されたアクティブドメイン:"
        jq -r '.network | to_entries[] | select(.value.is_active == true) | "  " + .value.domain + ":" + (.value.port | tostring) + " (" + .value.protocol + ")"' "$REAL_DATA_FILE" 2>/dev/null | head -10 || echo "  データなし"
        echo ""
        echo "🎯 フォーカス中のアプリ:"
        jq -r '.apps | to_entries[] | select(.value.is_focused == true) | "  " + .key + " (フォーカス時間: " + (.value.focus_time | tostring) + "秒)"' "$REAL_DATA_FILE" 2>/dev/null || echo "  データなし"
    fi
else
    echo "データファイルが見つかりません"
fi

echo ""
echo "📄 最新ログ:"
echo "-----------"
echo "エージェント:"
tail -5 "$HOME/.roiagent/logs/agent.log" 2>/dev/null || echo "ログなし"
echo ""
echo "Web UI:"
tail -5 "$HOME/.roiagent/logs/webui.log" 2>/dev/null || echo "ログなし"

echo ""
echo "🔍 DNS監視状況:"
echo "---------------"
if [ -f "$HOME/.roiagent/logs/agent.log" ]; then
    DNS_COUNT=$(grep -c "DNS Query detected" "$HOME/.roiagent/logs/agent.log" 2>/dev/null || echo "0")
    echo "  DNS検出回数: $DNS_COUNT"
    if [ "$DNS_COUNT" -gt 0 ]; then
        echo "  最近の検出例:"
        grep "DNS Query detected" "$HOME/.roiagent/logs/agent.log" | tail -3 | sed 's/^/    /' || echo "    なし"
    fi
    
    NETWORK_COUNT=$(grep -c "Network update" "$HOME/.roiagent/logs/agent.log" 2>/dev/null || echo "0")
    echo "  ネットワーク更新回数: $NETWORK_COUNT"
    
    APP_COUNT=$(grep -c "App update" "$HOME/.roiagent/logs/agent.log" 2>/dev/null || echo "0")
    echo "  アプリ更新回数: $APP_COUNT"
else
    echo "  ログファイルなし"
fi

echo ""
echo "ℹ️  データは $HOME/.roiagent/data/ に保存されています"
echo "📊 Web UI: http://localhost:5002 でデータを確認できます"
