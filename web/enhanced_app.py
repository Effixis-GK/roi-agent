#!/usr/bin/env python3
"""
ROI Agent - Enhanced Web UI with Network Monitoring
Flask-based web interface for viewing application and network usage statistics
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
HOME_DIR = os.path.expanduser("~")
USER_DATA_DIR = os.path.join(HOME_DIR, ".roiagent")
DATA_DIR = os.path.join(USER_DATA_DIR, "data")
LOGS_DIR = os.path.join(USER_DATA_DIR, "logs")

# Ensure directories exist
os.makedirs(DATA_DIR, exist_ok=True)
os.makedirs(LOGS_DIR, exist_ok=True)

# Try to find monitor binary
POSSIBLE_MONITOR_PATHS = [
    os.path.join(os.path.dirname(os.path.dirname(__file__)), "MacOS", "monitor"),
    "/Applications/ROI Agent.app/Contents/MacOS/monitor",
    os.path.join(USER_DATA_DIR, "monitor")
]

AGENT_BINARY = None
for path in POSSIBLE_MONITOR_PATHS:
    if os.path.exists(path):
        AGENT_BINARY = path
        break

class EnhancedMonitorUI:
    def __init__(self):
        self.data_dir = DATA_DIR
        
    def load_combined_data(self, date=None):
        """Load combined usage data for a specific date"""
        if date is None:
            date = datetime.now().strftime("%Y-%m-%d")
        
        # Try new combined format first
        combined_file = os.path.join(self.data_dir, f"combined_{date}.json")
        if os.path.exists(combined_file):
            try:
                with open(combined_file, 'r') as f:
                    return json.load(f)
            except Exception as e:
                print(f"Error loading combined data file {combined_file}: {e}")
        
        # Fallback to legacy format
        legacy_file = os.path.join(self.data_dir, f"usage_{date}.json")
        if os.path.exists(legacy_file):
            try:
                with open(legacy_file, 'r') as f:
                    legacy_data = json.load(f)
                    # Convert to new format
                    return {
                        "date": date,
                        "apps": legacy_data.get("apps", {}),
                        "network": {},
                        "app_total": legacy_data.get("total", {}),
                        "network_total": {
                            "total_duration": 0,
                            "total_bytes_sent": 0,
                            "total_bytes_received": 0,
                            "unique_connections": 0
                        }
                    }
            except Exception as e:
                print(f"Error loading legacy data file {legacy_file}: {e}")
        
        # Return empty data structure
        return {
            "date": date,
            "apps": {},
            "network": {},
            "app_total": {
                "foreground_time": 0,
                "background_time": 0,
                "focus_time": 0
            },
            "network_total": {
                "total_duration": 0,
                "total_bytes_sent": 0,
                "total_bytes_received": 0,
                "unique_connections": 0
            }
        }
    
    def get_agent_status(self):
        """Get current agent status"""
        if not AGENT_BINARY:
            return {"error": "Agent binary not found", "running": False}
            
        try:
            result = subprocess.run(
                [AGENT_BINARY, "status"], 
                capture_output=True, 
                text=True, 
                timeout=10,
                cwd=USER_DATA_DIR
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
    
    def format_bytes(self, bytes_count):
        """Format bytes into human readable size"""
        if bytes_count < 1024:
            return f"{bytes_count}B"
        elif bytes_count < 1024 * 1024:
            return f"{bytes_count / 1024:.1f}KB"
        elif bytes_count < 1024 * 1024 * 1024:
            return f"{bytes_count / (1024 * 1024):.1f}MB"
        else:
            return f"{bytes_count / (1024 * 1024 * 1024):.1f}GB"
    
    def get_app_ranking_data(self, data, category="foreground_time"):
        """Get app ranking data for specified category - only active apps"""
        if not data or not data.get("apps"):
            return []
        
        apps = []
        for app_name, app_data in data["apps"].items():
            # Only include currently active apps
            if not app_data.get("is_active", False):
                continue
                
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
        
        apps.sort(key=lambda x: x["time"], reverse=True)
        return apps
    
    def get_network_ranking_data(self, data, category="duration"):
        """Get network ranking data for specified category - only active connections"""
        if not data or not data.get("network"):
            return []
        
        connections = []
        for conn_key, conn_data in data["network"].items():
            # Only include currently active connections
            if not conn_data.get("is_active", False):
                continue
                
            if category == "duration":
                value = conn_data.get("duration", 0)
                formatted_value = self.format_time(value)
            elif category == "bytes_sent":
                value = conn_data.get("bytes_sent", 0)
                formatted_value = self.format_bytes(value)
            elif category == "bytes_received":
                value = conn_data.get("bytes_received", 0)
                formatted_value = self.format_bytes(value)
            else:
                value = conn_data.get("duration", 0)
                formatted_value = self.format_time(value)
            
            if value > 0:
                connections.append({
                    "domain": conn_data.get("domain", "Unknown"),
                    "port": conn_data.get("port", 0),
                    "protocol": conn_data.get("protocol", "Unknown"),
                    "app_name": conn_data.get("app_name", "Unknown"),
                    "value": value,
                    "formatted_value": formatted_value,
                    "duration": conn_data.get("duration", 0),
                    "bytes_sent": conn_data.get("bytes_sent", 0),
                    "bytes_received": conn_data.get("bytes_received", 0),
                    "is_active": conn_data.get("is_active", False),
                    "first_seen": conn_data.get("first_seen", ""),
                    "last_seen": conn_data.get("last_seen", "")
                })
        
        connections.sort(key=lambda x: x["value"], reverse=True)
        return connections

# Initialize UI handler
ui = EnhancedMonitorUI()

@app.route('/')
def index():
    """Main dashboard page"""
    return render_template('enhanced_index.html')

@app.route('/api/status')
def api_status():
    """Get agent status"""
    return jsonify(ui.get_agent_status())

@app.route('/api/data')
def api_data():
    """Get combined usage data"""
    date = request.args.get('date')
    app_category = request.args.get('app_category', 'foreground_time')
    network_category = request.args.get('network_category', 'duration')
    data_type = request.args.get('type', 'both')  # both, apps, network
    
    data = ui.load_combined_data(date)
    if data is None:
        return jsonify({"error": "Failed to load data"}), 500
    
    response = {
        "date": data["date"],
        "app_total": data["app_total"],
        "network_total": data["network_total"]
    }
    
    if data_type in ['both', 'apps']:
        response["app_ranking"] = ui.get_app_ranking_data(data, app_category)
        response["app_category"] = app_category
    
    if data_type in ['both', 'network']:
        response["network_ranking"] = ui.get_network_ranking_data(data, network_category)
        response["network_category"] = network_category
    
    return jsonify(response)

@app.route('/api/dates')
def api_dates():
    """Get available data dates"""
    if not os.path.exists(DATA_DIR):
        return jsonify([])
    
    dates = set()
    for filename in os.listdir(DATA_DIR):
        if filename.startswith("combined_") and filename.endswith(".json"):
            date_str = filename[9:-5]  # Remove "combined_" and ".json"
            dates.add(date_str)
        elif filename.startswith("usage_") and filename.endswith(".json"):
            date_str = filename[6:-5]  # Remove "usage_" and ".json"
            dates.add(date_str)
    
    sorted_dates = sorted(list(dates), reverse=True)
    return jsonify(sorted_dates)

@app.route('/api/network/domains')
def api_network_domains():
    """Get network usage by domain"""
    date = request.args.get('date')
    data = ui.load_combined_data(date)
    
    if not data or not data.get("network"):
        return jsonify([])
    
    # Group by domain
    domain_stats = {}
    for conn_key, conn_data in data["network"].items():
        domain = conn_data.get("domain", "Unknown")
        
        if domain not in domain_stats:
            domain_stats[domain] = {
                "domain": domain,
                "total_duration": 0,
                "total_bytes_sent": 0,
                "total_bytes_received": 0,
                "connections": 0,
                "protocols": set(),
                "apps": set()
            }
        
        stats = domain_stats[domain]
        stats["total_duration"] += conn_data.get("duration", 0)
        stats["total_bytes_sent"] += conn_data.get("bytes_sent", 0)
        stats["total_bytes_received"] += conn_data.get("bytes_received", 0)
        stats["connections"] += 1
        stats["protocols"].add(conn_data.get("protocol", "Unknown"))
        stats["apps"].add(conn_data.get("app_name", "Unknown"))
    
    # Convert sets to lists and format
    result = []
    for domain, stats in domain_stats.items():
        result.append({
            "domain": domain,
            "total_duration": stats["total_duration"],
            "formatted_duration": ui.format_time(stats["total_duration"]),
            "total_bytes_sent": stats["total_bytes_sent"],
            "formatted_bytes_sent": ui.format_bytes(stats["total_bytes_sent"]),
            "total_bytes_received": stats["total_bytes_received"],
            "formatted_bytes_received": ui.format_bytes(stats["total_bytes_received"]),
            "connections": stats["connections"],
            "protocols": list(stats["protocols"]),
            "apps": list(stats["apps"])
        })
    
    # Sort by duration
    result.sort(key=lambda x: x["total_duration"], reverse=True)
    return jsonify(result)

if __name__ == '__main__':
    print("=== ROI Agent Enhanced Web UI (Apps + Network) ===")
    print(f"Data directory: {DATA_DIR}")
    print(f"Agent binary: {AGENT_BINARY}")
    print("Starting enhanced web server on http://localhost:5002")
    print("Features: Application monitoring + Network monitoring")
    print("Press Ctrl+C to stop")
    
    app.run(host='127.0.0.1', port=5002, debug=False)
