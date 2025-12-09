#!/bin/bash
# ROI Agent Release Script
# Usage: ./scripts/release.sh [version]
# Example: ./scripts/release.sh 1.4.0

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
GCS_BUCKET="roi-agent-releases"
CLOUDSQL_INSTANCE="roi-production"
CLOUDSQL_USER="admin"
CLOUDSQL_DB="production"  
GCP_PROJECT="teak-frame-465410-a0"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from argument or VERSION file
VERSION="${1:-$(cat "$PROJECT_ROOT/VERSION" | tr -d '\n')}"

if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Version not specified${NC}"
    echo "Usage: $0 [version]"
    exit 1
fi

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  ROI Agent Release Script${NC}"
echo -e "${GREEN}  Version: ${VERSION}${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""

# Confirm release
read -p "Release version $VERSION? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release cancelled."
    exit 0
fi

# Step 1: Build PKG for macOS
echo -e "\n${YELLOW}Step 1: Building macOS PKG...${NC}"
cd "$PROJECT_ROOT"

# Build agent
echo "Building agent..."
cd agent
GOOS=darwin GOARCH=arm64 go build -o "$PROJECT_ROOT/build/macos-arm64/roi-agent" .
GOOS=darwin GOARCH=amd64 go build -o "$PROJECT_ROOT/build/macos-amd64/roi-agent" .

# Build data-sender
echo "Building data-sender..."
cd ../data-sender
GOOS=darwin GOARCH=arm64 go build -o "$PROJECT_ROOT/build/macos-arm64/data-sender" .
GOOS=darwin GOARCH=amd64 go build -o "$PROJECT_ROOT/build/macos-amd64/data-sender" .

cd "$PROJECT_ROOT"
echo -e "${GREEN}✅ Build complete${NC}"

# Step 2: Create PKG (if pkgbuild is available)
PKG_ARM64="$PROJECT_ROOT/build/ROI-Agent-macOS-arm64.pkg"
PKG_AMD64="$PROJECT_ROOT/build/ROI-Agent-macOS-x64.pkg"

if command -v pkgbuild &> /dev/null; then
    echo -e "\n${YELLOW}Step 2: Creating PKG files...${NC}"
    
    # Create payload directories
    mkdir -p "$PROJECT_ROOT/build/payload-arm64/Applications/ROI Agent/bin"
    mkdir -p "$PROJECT_ROOT/build/payload-arm64/Applications/ROI Agent/Resources"
    mkdir -p "$PROJECT_ROOT/build/payload-amd64/Applications/ROI Agent/bin"
    mkdir -p "$PROJECT_ROOT/build/payload-amd64/Applications/ROI Agent/Resources"
    
    # Copy files
    cp "$PROJECT_ROOT/build/macos-arm64/roi-agent" "$PROJECT_ROOT/build/payload-arm64/Applications/ROI Agent/bin/"
    cp "$PROJECT_ROOT/build/macos-arm64/data-sender" "$PROJECT_ROOT/build/payload-arm64/Applications/ROI Agent/bin/"
    echo "$VERSION" > "$PROJECT_ROOT/build/payload-arm64/Applications/ROI Agent/Resources/VERSION"
    
    cp "$PROJECT_ROOT/build/macos-amd64/roi-agent" "$PROJECT_ROOT/build/payload-amd64/Applications/ROI Agent/bin/"
    cp "$PROJECT_ROOT/build/macos-amd64/data-sender" "$PROJECT_ROOT/build/payload-amd64/Applications/ROI Agent/bin/"
    echo "$VERSION" > "$PROJECT_ROOT/build/payload-amd64/Applications/ROI Agent/Resources/VERSION"
    
    # Build PKGs
    pkgbuild --root "$PROJECT_ROOT/build/payload-arm64" \
             --identifier "com.roiagent.monitor" \
             --version "$VERSION" \
             --install-location "/" \
             "$PKG_ARM64"
    
    pkgbuild --root "$PROJECT_ROOT/build/payload-amd64" \
             --identifier "com.roiagent.monitor" \
             --version "$VERSION" \
             --install-location "/" \
             "$PKG_AMD64"
    
    echo -e "${GREEN}✅ PKG files created${NC}"
else
    echo -e "${YELLOW}⚠️  pkgbuild not available, skipping PKG creation${NC}"
fi

# Step 3: Upload to GCS
echo -e "\n${YELLOW}Step 3: Uploading to GCS...${NC}"

if [ -f "$PKG_ARM64" ]; then
    gsutil cp "$PKG_ARM64" "gs://${GCS_BUCKET}/latest/macos/ROI-Agent-macOS-arm64.pkg"
    gsutil cp "$PKG_ARM64" "gs://${GCS_BUCKET}/versions/${VERSION}/macos/ROI-Agent-macOS-arm64.pkg"
    echo -e "${GREEN}✅ Uploaded arm64 PKG${NC}"
fi

if [ -f "$PKG_AMD64" ]; then
    gsutil cp "$PKG_AMD64" "gs://${GCS_BUCKET}/latest/macos/ROI-Agent-macOS-x64.pkg"
    gsutil cp "$PKG_AMD64" "gs://${GCS_BUCKET}/versions/${VERSION}/macos/ROI-Agent-macOS-x64.pkg"
    echo -e "${GREEN}✅ Uploaded amd64 PKG${NC}"
fi

# Step 4: Update CloudSQL
echo -e "\n${YELLOW}Step 4: Updating CloudSQL...${NC}"

SQL_QUERY="
INSERT INTO agent_releases (
    version,
    macos_arm64_url,
    macos_amd64_url,
    is_latest,
    release_notes,
    published_at
) VALUES (
    '${VERSION}',
    'https://storage.googleapis.com/${GCS_BUCKET}/latest/macos/ROI-Agent-macOS-arm64.pkg',
    'https://storage.googleapis.com/${GCS_BUCKET}/latest/macos/ROI-Agent-macOS-x64.pkg',
    TRUE,
    'Release ${VERSION}',
    CURRENT_TIMESTAMP
) ON CONFLICT (version) DO UPDATE SET 
    is_latest = TRUE,
    macos_arm64_url = EXCLUDED.macos_arm64_url,
    macos_amd64_url = EXCLUDED.macos_amd64_url,
    published_at = CURRENT_TIMESTAMP;

UPDATE agent_releases SET is_latest = false WHERE version != '${VERSION}';
"

# Check if psql is available
if command -v psql &> /dev/null; then
    echo "Using psql with Cloud SQL Proxy..."
    
    # Download Cloud SQL Proxy if not exists
    if [ ! -f "/tmp/cloud-sql-proxy" ]; then
        echo "Downloading Cloud SQL Proxy..."
        curl -o /tmp/cloud-sql-proxy https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.8.0/cloud-sql-proxy.darwin.arm64
        chmod +x /tmp/cloud-sql-proxy
    fi
    
    # Kill existing Cloud SQL Proxy if running
    pkill -f "cloud-sql-proxy.*${CLOUDSQL_INSTANCE}" 2>/dev/null || true
    sleep 2
    
    # Start Cloud SQL Proxy in background
    /tmp/cloud-sql-proxy --port 5433 "${GCP_PROJECT}:asia-northeast1:${CLOUDSQL_INSTANCE}" &
    PROXY_PID=$!
    
    # Wait for proxy to be ready
    sleep 5
    
    # Execute SQL
    echo "Executing SQL..."
    PGPASSWORD="${CLOUDSQL_PASSWORD}" psql \
        -h localhost \
        -p 5433 \
        -U ${CLOUDSQL_USER} \
        -d ${CLOUDSQL_DB} \
        -c "${SQL_QUERY}"
    
    # Stop proxy
    kill $PROXY_PID
    
    echo -e "${GREEN}✅ CloudSQL updated${NC}"
else
    echo -e "${YELLOW}⚠️  psql not found. Install PostgreSQL client:${NC}"
    echo "  brew install postgresql@15"
    echo ""
    echo "Or manually run this SQL in CloudSQL:"
    echo "${SQL_QUERY}"
    echo ""
    read -p "Press Enter to continue without CloudSQL update, or Ctrl+C to cancel..."
fi

# Step 5: Update VERSION file
echo -e "\n${YELLOW}Step 5: Updating VERSION file...${NC}"
echo "$VERSION" > "$PROJECT_ROOT/VERSION"

# Done
echo -e "\n${GREEN}=========================================${NC}"
echo -e "${GREEN}  ✅ Release ${VERSION} complete!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Next steps:"
echo "  1. Commit and push changes"
echo "  2. Create GitHub release tag: git tag v${VERSION} && git push origin v${VERSION}"
echo ""

