#!/bin/bash

# ROI Agent Cloud Storage Uploader - 完全版
# GitHub Actions対応 + Cloud Storage自動上書き

set -e

echo "🚀 ROI Agent - Cloud Storage Upload Script (v2.0)"
echo "=============================================="

# 設定
PROJECT_ROOT="/Users/taktakeu/Local/roi-agent"
BUCKET_NAME="roi-agent-releases"
VERSION="latest"

# GCPプロジェクト確認
echo "📋 Checking GCP project..."
PROJECT_ID=$(gcloud config get-value project)
echo "   Project: $PROJECT_ID"

# バケット存在確認
echo "📦 Checking bucket: gs://$BUCKET_NAME"
if ! gsutil ls -b gs://$BUCKET_NAME > /dev/null 2>&1; then
    echo "❌ Bucket not found. Creating..."
    gsutil mb -p $PROJECT_ID -l asia-northeast1 gs://$BUCKET_NAME
    echo "✅ Bucket created"
fi

cd "$PROJECT_ROOT"

# buildディレクトリ作成
mkdir -p build

echo ""
echo "🔨 Building Go binaries..."

# Agent (Intel)
echo "  - Building agent (Intel)..."
cd agent
GOOS=darwin GOARCH=amd64 go build -o ../build/roi-agent-intel main.go
chmod +x ../build/roi-agent-intel
echo "    ✅ roi-agent-intel ($(du -h ../build/roi-agent-intel | cut -f1))"

# Agent (ARM64)
echo "  - Building agent (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -o ../build/roi-agent-arm64 main.go
chmod +x ../build/roi-agent-arm64
echo "    ✅ roi-agent-arm64 ($(du -h ../build/roi-agent-arm64 | cut -f1))"

# Data Sender (Intel)
echo "  - Building data-sender (Intel)..."
cd ../data-sender
GOOS=darwin GOARCH=amd64 go build -o ../build/data-sender-intel main.go config.go logger.go processor.go sender.go types.go utils.go
chmod +x ../build/data-sender-intel
echo "    ✅ data-sender-intel ($(du -h ../build/data-sender-intel | cut -f1))"

# Data Sender (ARM64)
echo "  - Building data-sender (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -o ../build/data-sender-arm64 main.go config.go logger.go processor.go sender.go types.go utils.go
chmod +x ../build/data-sender-arm64
echo "    ✅ data-sender-arm64 ($(du -h ../build/data-sender-arm64 | cut -f1))"

cd "$PROJECT_ROOT"

echo ""
echo "📦 Creating resource archives..."

# Web UI
echo "  - Zipping web directory..."
cd web
zip -r ../build/web.zip . -x "*.pyc" -x "__pycache__/*" -x ".DS_Store"
cd ..
echo "    ✅ web.zip ($(du -h build/web.zip | cut -f1))"

# Scripts
echo "  - Zipping scripts directory..."
cd scripts
zip -r ../build/scripts.zip . -x ".DS_Store"
cd ..
echo "    ✅ scripts.zip ($(du -h build/scripts.zip | cut -f1))"

echo ""
echo "🎨 Creating AppIcon.icns..."

# AppIcon.icns作成（sipsとiconutilが利用可能な場合）
ICON_PNG="$PROJECT_ROOT/public/icon.png"
ICONSET_DIR="$PROJECT_ROOT/build/AppIcon.iconset"
OUTPUT_ICNS="$PROJECT_ROOT/build/AppIcon.icns"

if [ -f "$ICON_PNG" ]; then
    if command -v sips > /dev/null && command -v iconutil > /dev/null; then
        echo "  - Generating iconset..."
        mkdir -p "$ICONSET_DIR"
        
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
        
        iconutil -c icns "$ICONSET_DIR" -o "$OUTPUT_ICNS"
        rm -rf "$ICONSET_DIR"
        
        echo "    ✅ AppIcon.icns ($(du -h $OUTPUT_ICNS | cut -f1))"
    else
        echo "    ⚠️  sips/iconutil not available - skipping .icns generation"
        echo "    💡 Tip: Run this on macOS for full icon support"
    fi
else
    echo "    ⚠️  icon.png not found at $ICON_PNG"
fi

echo ""
echo "🗑️  Cleaning old files from Cloud Storage..."
echo "  - Removing old version files (keeping bucket structure)..."

# 古いファイルを削除（バケット全体ではなく、latestバージョンのみ）
gsutil -m rm -f gs://$BUCKET_NAME/$VERSION/macos/* 2>/dev/null || true
gsutil -m rm -f gs://$BUCKET_NAME/$VERSION/macos/resources/* 2>/dev/null || true
gsutil -m rm -f gs://$BUCKET_NAME/$VERSION/windows/* 2>/dev/null || true
gsutil -m rm -f gs://$BUCKET_NAME/$VERSION/*.json 2>/dev/null || true

echo "    ✅ Old files cleaned"

echo ""
echo "☁️  Uploading to Cloud Storage..."

# バイナリ
echo "  - Uploading binaries..."
gsutil -h "Cache-Control:no-cache, max-age=0" cp build/roi-agent-intel gs://$BUCKET_NAME/$VERSION/macos/
gsutil -h "Cache-Control:no-cache, max-age=0" cp build/roi-agent-arm64 gs://$BUCKET_NAME/$VERSION/macos/
gsutil -h "Cache-Control:no-cache, max-age=0" cp build/data-sender-intel gs://$BUCKET_NAME/$VERSION/macos/
gsutil -h "Cache-Control:no-cache, max-age=0" cp build/data-sender-arm64 gs://$BUCKET_NAME/$VERSION/macos/
echo "    ✅ Binaries uploaded"

# リソース
echo "  - Uploading resources..."
gsutil -h "Cache-Control:no-cache, max-age=0" cp build/web.zip gs://$BUCKET_NAME/$VERSION/macos/resources/
gsutil -h "Cache-Control:no-cache, max-age=0" cp build/scripts.zip gs://$BUCKET_NAME/$VERSION/macos/resources/

# アイコン（PNG）
if [ -f "$ICON_PNG" ]; then
    gsutil -h "Cache-Control:no-cache, max-age=0" cp "$ICON_PNG" gs://$BUCKET_NAME/$VERSION/macos/resources/icon.png
    echo "    ✅ icon.png uploaded"
fi

# アイコン（ICNS）
if [ -f "$OUTPUT_ICNS" ]; then
    gsutil -h "Cache-Control:no-cache, max-age=0" cp "$OUTPUT_ICNS" gs://$BUCKET_NAME/$VERSION/macos/resources/AppIcon.icns
    echo "    ✅ AppIcon.icns uploaded"
else
    echo "    ⚠️  AppIcon.icns not found - skipped"
fi

# version.json
echo "  - Creating version.json..."
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

cat > build/version.json << EOF
{
  "version": "$VERSION",
  "release_date": "$BUILD_DATE",
  "commit": "$BUILD_COMMIT",
  "changelog": "Complete .app bundle with AppIcon, Web UI, and launcher",
  "features": [
    "Application monitoring",
    "DNS/Network monitoring",
    "Web dashboard (localhost:5002)",
    "Data transmission (10min interval)",
    "Organization-specific API key"
  ],
  "platforms": {
    "macos": {
      "architectures": ["amd64", "arm64"],
      "requires_accessibility": true,
      "requires_sudo": true,
      "icon_support": true
    },
    "windows": {
      "architectures": ["amd64"]
    }
  }
}
EOF

gsutil -h "Cache-Control:no-cache, max-age=0" cp build/version.json gs://$BUCKET_NAME/$VERSION/version.json
echo "    ✅ version.json uploaded"

echo ""
echo "✅ Upload completed!"
echo ""
echo "📁 Cloud Storage structure:"
gsutil ls -r gs://$BUCKET_NAME/$VERSION/

echo ""
echo "📊 File sizes:"
gsutil du -sh gs://$BUCKET_NAME/$VERSION/macos/*
gsutil du -sh gs://$BUCKET_NAME/$VERSION/macos/resources/*

echo ""
echo "🎉 All files uploaded successfully!"
echo ""
echo "💡 Note: Old files have been removed to keep costs low"
echo "💡 GitHub Actions: Ensure sips/iconutil are available for .icns generation"
