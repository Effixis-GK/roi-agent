#!/bin/bash

# ROI Agent Enhanced - Quick Project Reorganization
echo "🗂️ ROI Agent Enhanced - Project Reorganization"
echo "==============================================="

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
cd "$BASE_DIR"

# Step 1: Organize structure
echo "1. Organizing project structure..."
chmod +x organize_structure.sh
./organize_structure.sh

# Step 2: Update README
echo "2. Updating README..."
mv README_new.md README.md

# Step 3: Set permissions for organized files
echo "3. Setting permissions..."
chmod +x scripts/*.sh 2>/dev/null || true
chmod +x debug/*.py 2>/dev/null || true

# Step 4: Clean up temporary files
echo "4. Cleaning up..."
rm -f organize_structure.sh
rm -f run_cleanup.sh cleanup.sh setup_final_structure.sh

echo ""
echo "✅ Project reorganization completed!"
echo ""
echo "📁 Final Clean Structure:"
echo "========================"
echo ""
echo "roi-agent/"
echo "├── agent/                          # Core監視エージェント" 
echo "├── web/                            # Web UI"
echo "├── config/                         # 設定ファイル"
echo "├── scripts/                        # 🆕 Shell Scripts"
echo "│   ├── build_enhanced.sh"
echo "│   ├── quick_setup_enhanced.sh"
echo "│   ├── start_enhanced_fqdn_monitoring.sh"
echo "│   ├── setup_permissions.sh"
echo "│   └── create_dmg.sh"
echo "├── debug/                          # 🆕 Debug Tools"
echo "│   ├── network_fqdn_debug.py"
echo "│   └── real_data_debug.py"
echo "├── build/                          # ビルド出力"
echo "├── data/                           # 実データ"
echo "├── logs/                           # ログファイル"
echo "└── README.md                       # 新しいドキュメント"
echo ""
echo "🚀 新しい使用方法:"
echo "=================="
echo ""
echo "📱 開発者モード（コンソール）:"
echo "   ./scripts/start_enhanced_fqdn_monitoring.sh"
echo ""
echo "🏗️ Macアプリ作成:"
echo "   ./scripts/quick_setup_enhanced.sh"
echo ""
echo "🔍 デバッグ:"
echo "   python3 debug/network_fqdn_debug.py full"
echo "   python3 debug/real_data_debug.py monitor"
echo ""
echo "✨ すべて実データのみを使用 - テストデータなし"
echo ""
