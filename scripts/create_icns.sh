#!/bin/bash

# Create AppIcon.icns from icon.png for Cloud Storage upload

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

ICON_PNG="$PROJECT_ROOT/public/icon.png"
OUTPUT_DIR="$PROJECT_ROOT/build"
ICONSET_DIR="$OUTPUT_DIR/AppIcon.iconset"
OUTPUT_ICNS="$OUTPUT_DIR/AppIcon.icns"

echo "🎨 Creating AppIcon.icns from icon.png"

# Check if icon.png exists
if [ ! -f "$ICON_PNG" ]; then
    echo "❌ Error: icon.png not found at $ICON_PNG"
    exit 1
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"
mkdir -p "$ICONSET_DIR"

echo "📦 Generating iconset..."

# Generate all required sizes
sips -z 16 16 "$ICON_PNG" --out "$ICONSET_DIR/icon_16x16.png" > /dev/null 2>&1
sips -z 32 32 "$ICON_PNG" --out "$ICONSET_DIR/icon_16x16@2x.png" > /dev/null 2>&1
sips -z 32 32 "$ICON_PNG" --out "$ICONSET_DIR/icon_32x32.png" > /dev/null 2>&1
sips -z 64 64 "$ICON_PNG" --out "$ICONSET_DIR/icon_32x32@2x.png" > /dev/null 2>&1
sips -z 128 128 "$ICON_PNG" --out "$ICONSET_DIR/icon_128x128.png" > /dev/null 2>&1
sips -z 256 256 "$ICON_PNG" --out "$ICONSET_DIR/icon_128x128@2x.png" > /dev/null 2>&1
sips -z 256 256 "$ICON_PNG" --out "$ICONSET_DIR/icon_256x256.png" > /dev/null 2>&1
sips -z 512 512 "$ICON_PNG" --out "$ICONSET_DIR/icon_256x256@2x.png" > /dev/null 2>&1
sips -z 512 512 "$ICON_PNG" --out "$ICONSET_DIR/icon_512x512.png" > /dev/null 2>&1
sips -z 1024 1024 "$ICON_PNG" --out "$ICONSET_DIR/icon_512x512@2x.png" > /dev/null 2>&1

echo "🔨 Creating .icns file..."
iconutil -c icns "$ICONSET_DIR" -o "$OUTPUT_ICNS"

# Clean up iconset
rm -rf "$ICONSET_DIR"

echo "✅ AppIcon.icns created: $OUTPUT_ICNS"
echo "📊 Size: $(du -h "$OUTPUT_ICNS" | cut -f1)"

# Upload to Cloud Storage
echo ""
echo "☁️  Uploading to Cloud Storage..."

BUCKET_NAME="roi-agent-releases"
VERSION="latest"

gsutil cp "$OUTPUT_ICNS" "gs://$BUCKET_NAME/$VERSION/macos/resources/AppIcon.icns"

echo "✅ Uploaded to gs://$BUCKET_NAME/$VERSION/macos/resources/AppIcon.icns"
echo "🎉 Done!"
