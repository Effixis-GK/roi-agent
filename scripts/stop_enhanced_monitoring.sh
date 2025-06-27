#!/bin/bash
echo "🛑 ROI Agent Enhanced - 拡張FQDN監視停止"
echo "======================================="

echo "プロセス停止中..."
kill 36466 36557 2>/dev/null || true
pkill -f "monitor_enhanced" || true
pkill -f "enhanced_app.py" || true

sleep 3
echo "✅ 停止完了"

echo ""
echo "📊 収集された拡張ネットワークデータ:"
echo "============================"
TODAY=$(date +%Y-%m-%d)
REAL_DATA_FILE="/Users/taktakeu/.roiagent/data/combined_$TODAY.json"

if [ -f "$REAL_DATA_FILE" ]; then
    echo "拡張データファイル: $REAL_DATA_FILE"
    
    if command -v jq > /dev/null 2>&1; then
        echo ""
        echo "📈 統計情報:"
        echo "  アプリ数: $(jq '.apps | length' "$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo "  ネットワーク接続数: $(jq '.network | length' "$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo "  DNSクエリ数: $(jq '.dns_queries | length' "$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo "  ユニークドメイン数: $(jq '.network_total.unique_domains' "$REAL_DATA_FILE" 2>/dev/null || echo "0")"
        echo ""
        echo "🌐 実際に解決されたFQDN:"
        jq -r '.network | to_entries[] | select(.value.domain != .value.remote_ip) | "  " + .value.domain + " (" + .value.remote_ip + ")"' "$REAL_DATA_FILE" 2>/dev/null | head -10 || echo "  データなし"
        echo ""
        echo "📡 DNSクエリ履歴:"
        jq -r '.dns_queries[] | "  " + .domain + " (" + .timestamp + ")"' "$REAL_DATA_FILE" 2>/dev/null | tail -5 || echo "  データなし"
    fi
else
    echo "拡張データファイルが見つかりません"
fi

echo ""
echo "📄 最新ログ:"
echo "-----------"
echo "拡張エージェント:"
tail -5 "/Users/taktakeu/.roiagent/logs/enhanced_agent_20250627_203938.log" 2>/dev/null || echo "ログなし"
echo ""
echo "拡張Web UI:"
tail -5 "/Users/taktakeu/.roiagent/logs/enhanced_web_20250627_203938.log" 2>/dev/null || echo "ログなし"

echo ""
echo "🔍 FQDN解決状況:"
echo "---------------"
if [ -f "/Users/taktakeu/.roiagent/logs/enhanced_agent_20250627_203938.log" ]; then
    FQDN_COUNT=$(grep -c "Resolved" "/Users/taktakeu/.roiagent/logs/enhanced_agent_20250627_203938.log" 2>/dev/null || echo "0")
    echo "  FQDN解決回数: $FQDN_COUNT"
    if [ "$FQDN_COUNT" -gt 0 ]; then
        echo "  最近の解決例:"
        grep "Resolved" "/Users/taktakeu/.roiagent/logs/enhanced_agent_20250627_203938.log" | tail -3 | sed 's/^/    /' || echo "    なし"
    fi
else
    echo "  ログファイルなし"
fi
