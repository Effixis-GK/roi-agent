#!/bin/bash

# ROI Agent Data Transmission Debug Script

echo "🔧 ROI Agent Data Transmission Debug"
echo "===================================="

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Project root: $PROJECT_ROOT"

# Navigate to debug directory
cd "$PROJECT_ROOT/debug"

# Check if .env file exists in data-sender directory
ENV_FILE="$PROJECT_ROOT/data-sender/.env"
if [ ! -f "$ENV_FILE" ]; then
    echo "❌ Error: .env file not found at $ENV_FILE"
    echo "Please create the .env file with:"
    echo "ROI_AGENT_BASE_URL=https://test-bjdnhp7xna-an.a.run.app/api/v1/device"
    echo "ROI_AGENT_API_KEY=VvxFHdH4KKoux6n7"
    exit 1
fi

echo "✅ Found .env file at $ENV_FILE"
echo ""

# Show current .env contents (hide API key)
echo "📄 Current .env configuration:"
echo "=============================="
while IFS= read -r line; do
    if [[ $line == ROI_AGENT_API_KEY* ]]; then
        echo "ROI_AGENT_API_KEY=***hidden***"
    else
        echo "$line"
    fi
done < "$ENV_FILE"
echo ""

# Download dependencies if needed
echo "📦 Checking Go dependencies..."
if [ ! -f "go.sum" ]; then
    echo "Downloading dependencies..."
    go mod download
    go mod tidy
fi

# Build and run the debug tool
echo "🔨 Building debug tool..."
go build -o debug-tool test_data_transmission.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "🚀 Running debug test..."
    echo "======================="
    ./debug-tool
else
    echo "❌ Build failed"
    exit 1
fi
