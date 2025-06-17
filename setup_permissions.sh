#!/bin/bash

# ROI Agent - Final Setup Script
# スクリプトファイルに実行権限を付与

echo "🔧 ROI Agent - ファイル権限設定"
echo "================================="

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
cd "$BASE_DIR"

# スクリプトファイルに実行権限を付与
echo "📝 スクリプトファイルに実行権限を付与中..."

chmod +x build_agent.sh
chmod +x start_web.sh
chmod +x build_app.sh
chmod +x create_dmg.sh
chmod +x quick_setup.sh
chmod +x debug_tools.py

echo "✅ 実行権限付与完了"
echo ""
echo "🎯 利用可能なコマンド:"
echo "  ./quick_setup.sh      # ワンコマンドセットアップ"
echo "  ./build_app.sh        # アプリケーションビルド"
echo "  ./create_dmg.sh       # DMGインストーラー作成"
echo "  python debug_tools.py # デバッグツール"
echo ""
echo "📖 詳細な使用方法:"
echo "  README.md (日本語版)"
echo "  README-ja.md (詳細版)"
echo ""
echo "🚀 今すぐ開始:"
echo "  ./quick_setup.sh"
echo ""
