#!/bin/bash

# Build Go Agent for ROI Agent
cd /Users/taktakeu/Local/GitHub/roi-agent/agent
echo "Building Go agent..."
go build -o monitor main.go

if [ $? -eq 0 ]; then
    echo "✓ Agent built successfully: monitor"
else
    echo "✗ Agent build failed"
    exit 1
fi
