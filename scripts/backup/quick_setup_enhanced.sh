#!/bin/bash

# ROI Agent Enhanced - Network Monitoring Quick Setup
# ネットワーク監視機能付きワンコマンドセットアップ

set -e

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
BUILD_DIR="$BASE_DIR/build"
APP_NAME="ROI Agent Enhanced"

echo "🚀 ROI Agent Enhanced セットアップ開始"
echo "ネットワーク監視機能付きバージョン"
echo "=================================="

# 1. 前提条件チェック
echo "1. 前提条件チェック..."

# Go言語チェック
if ! command -v go &> /dev/null; then
    echo "❌ Go言語がインストールされていません"
    echo "   Homebrewでインストール: brew install go"
    exit 1
else
    echo "✅ Go: $(go version | cut -d' ' -f3)"
fi

# Python3チェック
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3がインストールされていません"
    exit 1
else
    echo "✅ Python: $(python3 --version | cut -d' ' -f2)"
fi

# 必要なPythonパッケージをチェック
echo "2. Python依存関係チェック..."
cd "$BASE_DIR/web"

if [ ! -d "venv" ]; then
    echo "   Python仮想環境を作成中..."
    python3 -m venv venv
fi

source venv/bin/activate
pip install --upgrade pip > /dev/null 2>&1
pip install flask requests > /dev/null 2>&1
echo "✅ Python依存関係: OK"

# 3. ネットワーク監視エージェントのビルド
echo "3. ネットワーク監視エージェント ビルド..."
cd "$BASE_DIR"

# 実行権限付与
chmod +x build_enhanced.sh
chmod +x network_debug_tools.py

# エージェントビルド
./build_enhanced.sh

echo "✅ アプリケーションビルド完了"

# 4. 初期テスト実行
echo "4. 初期機能テスト..."

# ネットワーク監視機能テスト
echo "   ネットワーク機能テスト実行中..."
python3 network_debug_tools.py testdata > /dev/null 2>&1
echo "✅ テストデータ生成完了"

# システム診断（詳細は表示しない）
python3 network_debug_tools.py network > /dev/null 2>&1
echo "✅ ネットワーク接続テスト完了"

# 5. アプリケーションをApplicationsにインストール
echo "5. アプリケーション インストール..."

APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"
INSTALL_PATH="/Applications/$APP_NAME.app"

if [ -d "$INSTALL_PATH" ]; then
    echo "   既存アプリを削除中..."
    rm -rf "$INSTALL_PATH"
fi

echo "   アプリをApplicationsフォルダにコピー中..."
cp -R "$APP_BUNDLE" "/Applications/"

echo "✅ インストール完了: $INSTALL_PATH"

# 6. 初回起動テスト
echo "6. 初回起動テスト..."

# アプリケーション権限確認
if "$INSTALL_PATH/Contents/MacOS/roi-agent" status > /dev/null 2>&1; then
    echo "✅ アプリケーション起動テスト: 成功"
else
    echo "⚠️  アプリケーション起動: 権限設定が必要"
fi

# 7. セットアップ完了とガイダンス
echo ""
echo "🎉 ROI Agent Enhanced セットアップ完了！"
echo "========================================"
echo ""
echo "📱 機能一覧:"
echo "   ✅ アプリケーション使用時間監視"
echo "   ✅ ネットワーク通信監視 (HTTP/HTTPS + カスタムポート)"
echo "   ✅ ドメインごとのアクセス時間統計"
echo "   ✅ リアルタイムダッシュボード"
echo "   ✅ 統合ビュー (アプリ + ネットワーク)"
echo ""
echo "🚀 使用開始方法:"
echo ""
echo "1. アプリケーション起動:"
echo "   方法A: Finderで「$APP_NAME」をダブルクリック"
echo "   方法B: ターミナルで以下を実行"
echo "   '$INSTALL_PATH/Contents/MacOS/roi-agent start'"
echo ""
echo "2. ダッシュボードアクセス:"
echo "   http://localhost:5002"
echo ""
echo "3. 基本コマンド:"
echo "   開始:     '$INSTALL_PATH/Contents/MacOS/roi-agent start'"
echo "   停止:     '$INSTALL_PATH/Contents/MacOS/roi-agent stop'"
echo "   状況確認:  '$INSTALL_PATH/Contents/MacOS/roi-agent status'"
echo "   ダッシュボード: '$INSTALL_PATH/Contents/MacOS/roi-agent dashboard'"
echo ""
echo "🔧 デバッグ・トラブルシューティング:"
echo "   完全診断:  'cd $BASE_DIR && python3 network_debug_tools.py full'"
echo "   ネットワークテスト: 'python3 network_debug_tools.py network'"
echo "   ログ確認:  '$INSTALL_PATH/Contents/MacOS/roi-agent logs'"
echo ""

# 8. 重要な注意事項
echo "⚠️  重要な設定:"
echo ""
echo "1. アクセシビリティ権限が必要です:"
echo "   システム環境設定 > セキュリティとプライバシー > プライバシー > アクセシビリティ"
echo "   「$APP_NAME」を追加して権限を付与してください"
echo ""
echo "2. ネットワーク監視の精度向上:"
echo "   より詳細な監視のため、管理者権限での実行を推奨"
echo "   'sudo $INSTALL_PATH/Contents/MacOS/roi-agent start'"
echo ""
echo "3. ファイアウォール設定:"
echo "   ポート5002でのローカル通信を許可してください"
echo ""

# 9. オプション: 自動起動設定の提案
echo "🔄 自動起動設定 (オプション):"
echo "   ログイン時に自動でROI Agentを開始したい場合:"
echo ""
echo "   システム環境設定 > ユーザとグループ > ログイン項目"
echo "   「$APP_NAME」を追加"
echo ""

# 10. 次のステップ提案
echo "📈 次のステップ:"
echo ""
echo "1. 権限設定後、アプリを起動してダッシュボードを確認"
echo "2. 数時間使用してデータ収集をテスト"
echo "3. ネットワーク監視機能で通信パターンを分析"
echo "4. 生産性レポートで時間の使い方を最適化"
echo ""

# 11. 問題が発生した場合のサポート
echo "🆘 サポート情報:"
echo ""
echo "問題が発生した場合:"
echo "1. 診断実行: 'cd $BASE_DIR && python3 network_debug_tools.py full'"
echo "2. ログ確認: '$INSTALL_PATH/Contents/MacOS/roi-agent logs'"
echo "3. 再セットアップ: '$BASE_DIR/quick_setup_enhanced.sh'"
echo ""

# 12. オプション: 即座に起動するか確認
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
read -p "🚀 今すぐROI Agent Enhancedを起動しますか？ (y/n): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "ROI Agent Enhanced を起動中..."
    
    # バックグラウンドで起動
    "$INSTALL_PATH/Contents/MacOS/roi-agent" start &
    
    # 少し待ってからダッシュボードを開く
    sleep 3
    echo "ダッシュボードを開いています..."
    open "http://localhost:5002"
    
    echo ""
    echo "✅ ROI Agent Enhanced が起動しました！"
    echo "   ダッシュボード: http://localhost:5002"
    echo "   停止: '$INSTALL_PATH/Contents/MacOS/roi-agent stop'"
else
    echo ""
    echo "手動で起動する場合:"
    echo "'$INSTALL_PATH/Contents/MacOS/roi-agent start'"
fi

echo ""
echo "🎉 セットアップ完了！生産性向上を楽しんでください！"
