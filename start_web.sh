#!/bin/bash

cd /Users/taktakeu/Local/GitHub/roi-agent/web

# Create virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
source venv/bin/activate

# Install requirements
if [ -f "requirements.txt" ]; then
    echo "Installing Python dependencies..."
    pip install -r requirements.txt
fi

echo "Starting ROI Agent Web UI on http://localhost:5002"
python app.py
