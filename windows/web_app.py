#!/usr/bin/env python3
"""
ROI Agent - Windows Web UI
Flask-based web interface for viewing application usage statistics on Windows
"""

import json
import os
import subprocess
import sys
from datetime import datetime, timedelta
from pathlib import Path

from flask import Flask, render_template, jsonify, request

app = Flask(__name__)

# Configuration
BASE_DIR = os.path.expanduser("~/.roiagent")
DATA_DIR = os.path.join(BASE_DIR, "data")
AGENT_BINARY = "roi-agent-windows.exe"  # Windows executable

class AppMonitorUI:
    def __init__(self):
        self.base_dir = BASE_DIR
        self.data_dir = DATA_DIR
        
    def load_daily_data(self, date=None):
        """Load usage data for a specific date"""
        if date is None:
            date = datetime.now().strftime("%Y-%m-%d")
        
        data_file = os.path.join(self.data_dir, f"usage_{date}.json")
        
        if not os.path.exists(data_file):
            return {
                "date": date,
                "apps": {},
                "total": {
                    "foreground_time": 0,
                    "background_time": 0,
                    "focus_time": 0
                }
            }
        
        try:
            with open(data_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            print(f"Error loading data file {data_file}: {e}")
            return None
    
    def get_agent_status(self):
        """Get current agent status"""
        try:
            # Try to find the agent binary in current directory or PATH
            agent_path = None
            if os.path.exists(AGENT_BINARY):
                agent_path = AGENT_BINARY
            else:
                # Try to find in PATH
                import shutil
                agent_path = shutil.which(AGENT_BINARY)
            
            if not agent_path:
                return {"error": "Agent binary not found", "running": False}
            
            result = subprocess.run(
                [agent_path, "status"], 
                capture_output=True, 
                text=True, 
                timeout=10
            )
            if result.returncode == 0:
                return json.loads(result.stdout)
            else:
                return {"error": "Agent not responding", "running": False}
        except FileNotFoundError:
            return {"error": "Agent binary not found", "running": False}
        except subprocess.TimeoutExpired:
            return {"error": "Agent status check timeout", "running": False}
        except Exception as e:
            return {"error": str(e), "running": False}
    
    def format_time(self, seconds):
        """Format seconds into human readable time"""
        if seconds < 60:
            return f"{seconds}s"
        elif seconds < 3600:
            minutes = seconds // 60
            secs = seconds % 60
            return f"{minutes}m {secs}s"
        else:
            hours = seconds // 3600
            minutes = (seconds % 3600) // 60
            return f"{hours}h {minutes}m"
    
    def get_ranking_data(self, data, category="foreground_time"):
        """Get ranking data for specified category"""
        if not data or not data.get("apps"):
            return []
        
        apps = []
        for app_name, app_data in data["apps"].items():
            usage_time = app_data.get(category, 0)
            if usage_time > 0:
                apps.append({
                    "name": app_name,
                    "time": usage_time,
                    "formatted_time": self.format_time(usage_time),
                    "foreground_time": app_data.get("foreground_time", 0),
                    "background_time": app_data.get("background_time", 0),
                    "focus_time": app_data.get("focus_time", 0),
                    "is_active": app_data.get("is_active", False),
                    "is_focused": app_data.get("is_focused", False),
                    "last_seen": app_data.get("last_seen", "")
                })
        
        # Sort by usage time (descending)
        apps.sort(key=lambda x: x["time"], reverse=True)
        return apps

# Initialize UI handler
ui = AppMonitorUI()

@app.route('/')
def index():
    """Main dashboard page"""
    return render_template('index.html')

@app.route('/api/status')
def api_status():
    """Get agent status"""
    return jsonify(ui.get_agent_status())

@app.route('/api/data')
def api_data():
    """Get usage data"""
    date = request.args.get('date')
    category = request.args.get('category', 'foreground_time')
    
    data = ui.load_daily_data(date)
    if data is None:
        return jsonify({"error": "Failed to load data"}), 500
    
    ranking = ui.get_ranking_data(data, category)
    
    return jsonify({
        "date": data["date"],
        "total": data["total"],
        "ranking": ranking,
        "category": category
    })

@app.route('/api/dates')
def api_dates():
    """Get available data dates"""
    if not os.path.exists(DATA_DIR):
        return jsonify([])
    
    dates = []
    for filename in os.listdir(DATA_DIR):
        if filename.startswith("usage_") and filename.endswith(".json"):
            date_str = filename[6:-5]  # Remove "usage_" and ".json"
            dates.append(date_str)
    
    dates.sort(reverse=True)
    return jsonify(dates)

# Create templates directory and save template
def create_template():
    template_dir = os.path.join(os.path.dirname(__file__), 'templates')
    os.makedirs(template_dir, exist_ok=True)

if __name__ == '__main__':
    # Ensure directories exist
    os.makedirs(DATA_DIR, exist_ok=True)
    
    # Create template
    create_template()
    
    print("=== ROI Agent Windows Web UI ===")
    print(f"Data directory: {DATA_DIR}")
    print(f"Agent binary: {AGENT_BINARY}")
    print("Starting web server on http://localhost:5002")
    print("Press Ctrl+C to stop")
    
    app.run(host='127.0.0.1', port=5002, debug=True)
