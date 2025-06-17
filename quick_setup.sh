#!/bin/bash

# ROI Agent - Quick Setup and Distribution
# One-command setup for building and packaging the app

set -e

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
BUILD_DIR="$BASE_DIR/build"

echo "🚀 ROI Agent - Quick Setup and Build"
echo "===================================="

cd "$BASE_DIR"

# Step 1: Build the Go agent
echo "📦 Step 1: Building Go agent..."
chmod +x build_agent.sh
./build_agent.sh

# Step 2: Build the macOS app bundle
echo "🍎 Step 2: Creating macOS app bundle..."
chmod +x build_app.sh
./build_app.sh

# Step 3: Create DMG installer
echo "💿 Step 3: Creating DMG installer..."
chmod +x create_dmg.sh
./create_dmg.sh

# Step 4: Show results
echo ""
echo "✅ ROI Agent build completed successfully!"
echo ""
echo "📁 Build artifacts:"
echo "   App Bundle: $BUILD_DIR/ROI Agent.app"
echo "   DMG Installer: $BUILD_DIR/ROI-Agent-Installer.dmg"
echo ""
echo "🎯 Quick actions:"
echo "   Test app:     open '$BUILD_DIR/ROI Agent.app'"
echo "   Open DMG:     open '$BUILD_DIR/ROI-Agent-Installer.dmg'"
echo "   Install app:  cp -R '$BUILD_DIR/ROI Agent.app' /Applications/"
echo ""
echo "📚 Usage after installation:"
echo "   • Double-click ROI Agent in Applications"
echo "   • Grant accessibility permissions when prompted"
echo "   • Dashboard opens automatically at http://localhost:5002"
echo "   • App runs in background monitoring usage"
echo ""
echo "🔧 Command line control:"
echo "   /Applications/ROI\ Agent.app/Contents/MacOS/roi-agent [start|stop|status|dashboard]"
echo ""

# Optional: Auto-install if requested
read -p "🤔 Install ROI Agent to Applications folder now? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "📥 Installing to Applications..."
    cp -R "$BUILD_DIR/ROI Agent.app" /Applications/
    echo "✅ Installed! You can now launch ROI Agent from Applications folder."
    
    read -p "🚀 Launch ROI Agent now? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        open "/Applications/ROI Agent.app"
        echo "🎉 ROI Agent launched! Dashboard will open shortly."
    fi
fi

echo ""
echo "🎉 Setup complete! Enjoy tracking your productivity with ROI Agent!"
