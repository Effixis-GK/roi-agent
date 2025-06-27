#!/bin/bash

# ROI Agent Enhanced - Organize project structure
echo "📁 Organizing project structure with dedicated folders..."

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
cd "$BASE_DIR"

# Create new folder structure
echo "Creating folders..."
mkdir -p scripts
mkdir -p debug
mkdir -p tools

# Move shell scripts to scripts folder
echo "Moving shell scripts..."
mv *.sh scripts/ 2>/dev/null || true

# Move Python debug tools to debug folder  
echo "Moving debug tools..."
mv network_fqdn_debug.py debug/ 2>/dev/null || true
mv real_data_debug.py debug/ 2>/dev/null || true

# Keep this organization script in root temporarily
cp scripts/organize_structure.sh . 2>/dev/null || true

echo "✅ Project reorganization completed!"
echo ""
echo "📁 New Structure:"
echo "roi-agent/"
echo "├── agent/                          # Core Go agent"
echo "│   ├── enhanced_network_main.go"
echo "│   └── go.mod"
echo "├── web/                           # Web UI"
echo "│   ├── enhanced_app.py"
echo "│   ├── requirements.txt"
echo "│   └── templates/"
echo "│       └── enhanced_index.html"
echo "├── config/                        # Configuration"
echo "│   └── config.yaml"
echo "├── scripts/                       # 🆕 Shell scripts"
echo "│   ├── build_enhanced.sh"
echo "│   ├── quick_setup_enhanced.sh"
echo "│   ├── start_enhanced_fqdn_monitoring.sh"
echo "│   ├── setup_permissions.sh"
echo "│   └── create_dmg.sh"
echo "├── debug/                         # 🆕 Debug tools"
echo "│   ├── network_fqdn_debug.py"
echo "│   └── real_data_debug.py"
echo "├── build/                         # Build output"
echo "├── data/                          # Real data"
echo "├── logs/                          # Log files"
echo "└── README.md"
echo ""
echo "🎯 New Usage Commands:"
echo "  App Build:     ./scripts/quick_setup_enhanced.sh"
echo "  Console Mode:  ./scripts/start_enhanced_fqdn_monitoring.sh"  
echo "  Debug Network: python3 debug/network_fqdn_debug.py full"
echo "  Debug Data:    python3 debug/real_data_debug.py monitor"
echo ""
