#!/bin/bash

# ROI Agent Enhanced - Cleanup unused files
echo "🧹 Cleaning up unused files..."

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
cd "$BASE_DIR"

# Remove old binaries
echo "Removing old binaries..."
rm -f agent/monitor agent/monitor_network agent/monitor_real agent/monitor_enhanced

# Remove development scripts (keeping only essential ones)
echo "Removing unused development scripts..."
rm -f dev_status.sh
rm -f dev_tools.sh
rm -f project_status.sh
rm -f quick_setup.sh
rm -f start_real_data_mode.sh
rm -f stop_development.sh
rm -f stop_enhanced_monitoring.sh
rm -f stop_real_development.sh
rm -f test_enhanced_build.sh
rm -f debug_tools.py

# Remove old documentation
echo "Removing old documentation..."
rm -f README.md
rm -f README-ja.md
rm -f INSTALLATION.md

# Clean Python cache
echo "Cleaning Python cache..."
rm -rf web/__pycache__

# Clean data and logs directories
echo "Cleaning temporary data..."
rm -rf data/*
rm -rf logs/*

echo "✅ Cleanup completed!"
echo ""
echo "Remaining files:"
find . -type f -name "*.sh" -o -name "*.py" -o -name "*.go" -o -name "*.md" | grep -v venv | sort
