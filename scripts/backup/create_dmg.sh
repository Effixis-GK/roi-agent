#!/bin/bash

# ROI Agent - DMG Installer Creator
# Creates a distributable DMG file for easy installation

set -e

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
BUILD_DIR="$BASE_DIR/build"
APP_NAME="ROI Agent"
BUNDLE_NAME="ROI Agent.app"
DMG_NAME="ROI-Agent-Installer"
TEMP_DMG="$BUILD_DIR/temp.dmg"
FINAL_DMG="$BUILD_DIR/$DMG_NAME.dmg"

echo "=== Creating ROI Agent DMG Installer ==="

# Check if app bundle exists
if [ ! -d "$BUILD_DIR/$BUNDLE_NAME" ]; then
    echo "❌ App bundle not found. Please run ./build_app.sh first"
    exit 1
fi

# Clean previous DMG
rm -f "$TEMP_DMG" "$FINAL_DMG"

# Create temporary directory for DMG contents
DMG_TEMP_DIR="$BUILD_DIR/dmg_temp"
rm -rf "$DMG_TEMP_DIR"
mkdir -p "$DMG_TEMP_DIR"

# Copy app bundle to DMG temp directory
cp -R "$BUILD_DIR/$BUNDLE_NAME" "$DMG_TEMP_DIR/"

# Create README for DMG
cat > "$DMG_TEMP_DIR/README.txt" << 'EOF'
ROI Agent - Productivity Monitoring Tool

INSTALLATION:
1. Drag "ROI Agent.app" to your Applications folder
2. Double-click "ROI Agent" in Applications to launch
3. Grant accessibility permissions when prompted
4. The app will start monitoring and open your dashboard

USAGE:
- The app runs in the background monitoring your application usage
- Access dashboard: Double-click ROI Agent or visit http://localhost:5002
- View productivity analytics and time tracking data

UNINSTALL:
- Move "ROI Agent.app" from Applications to Trash
- Remove data folder: ~/.roiagent (optional)

For support and documentation: https://github.com/your-repo/roi-agent
EOF

# Create symbolic link to Applications folder
ln -s /Applications "$DMG_TEMP_DIR/Applications"

# Calculate DMG size (app size + 50MB buffer)
APP_SIZE=$(du -sm "$BUILD_DIR/$BUNDLE_NAME" | cut -f1)
DMG_SIZE=$((APP_SIZE + 50))

# Create temporary DMG
hdiutil create -srcfolder "$DMG_TEMP_DIR" -volname "$APP_NAME" -fs HFS+ -fsargs "-c c=64,a=16,e=16" -format UDRW -size ${DMG_SIZE}m "$TEMP_DMG"

# Mount the temporary DMG
MOUNT_DIR=$(hdiutil attach -readwrite -noverify -noautoopen "$TEMP_DMG" | grep -E '^/dev/' | sed 1q | awk '{print $3}')

echo "Mounted DMG at: $MOUNT_DIR"

# Set DMG window properties using AppleScript
cat > "$BUILD_DIR/dmg_setup.applescript" << EOF
tell application "Finder"
    tell disk "$APP_NAME"
        open
        set current view of container window to icon view
        set toolbar visible of container window to false
        set statusbar visible of container window to false
        set the bounds of container window to {400, 100, 900, 400}
        set theViewOptions to the icon view options of container window
        set arrangement of theViewOptions to not arranged
        set icon size of theViewOptions to 100
        set background picture of theViewOptions to file ".background:background.png"
        set position of item "$BUNDLE_NAME" of container window to {150, 200}
        set position of item "Applications" of container window to {350, 200}
        close
        open
        update without registering applications
        delay 2
    end tell
end tell
EOF

# Apply DMG settings (suppress errors if Finder is not available)
osascript "$BUILD_DIR/dmg_setup.applescript" 2>/dev/null || echo "⚠️  Could not set DMG window properties"

# Unmount the DMG
hdiutil detach "$MOUNT_DIR"

# Convert to final compressed DMG
hdiutil convert "$TEMP_DMG" -format UDZO -imagekey zlib-level=9 -o "$FINAL_DMG"

# Clean up
rm -f "$TEMP_DMG"
rm -rf "$DMG_TEMP_DIR"
rm -f "$BUILD_DIR/dmg_setup.applescript"

echo ""
echo "✅ DMG Installer created successfully!"
echo ""
echo "📦 Location: $FINAL_DMG"
echo "📊 Size: $(du -h "$FINAL_DMG" | cut -f1)"
echo ""
echo "🚀 To distribute:"
echo "1. Share the DMG file with users"
echo "2. Users double-click the DMG to mount"
echo "3. Users drag ROI Agent to Applications folder"
echo "4. Users launch ROI Agent from Applications"
echo ""
