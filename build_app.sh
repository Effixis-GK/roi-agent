#!/bin/bash

# ROI Agent - macOS App Builder
# This script creates a native macOS application bundle

set -e

BASE_DIR="/Users/taktakeu/Local/GitHub/roi-agent"
APP_NAME="ROI Agent"
BUNDLE_NAME="ROI Agent.app"
BUILD_DIR="$BASE_DIR/build"
APP_DIR="$BUILD_DIR/$BUNDLE_NAME"

echo "=== Building ROI Agent macOS Application ==="

# Clean previous build
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Create app bundle structure
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"
mkdir -p "$APP_DIR/Contents/Frameworks"

# Create Info.plist
cat > "$APP_DIR/Contents/Info.plist" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>roi-agent</string>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
    <key>CFBundleIdentifier</key>
    <string>com.productivity.roiagent</string>
    <key>CFBundleName</key>
    <string>ROI Agent</string>
    <key>CFBundleDisplayName</key>
    <string>ROI Agent</string>
    <key>CFBundleVersion</key>
    <string>1.0.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleSignature</key>
    <string>ROIA</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.14</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSAppleEventsUsageDescription</key>
    <string>ROI Agent needs to monitor application usage for productivity analysis.</string>
    <key>NSSystemAdministrationUsageDescription</key>
    <string>ROI Agent needs system access to monitor running applications.</string>
</dict>
</plist>
EOF

# Build Go agent
echo "Building Go agent..."
cd "$BASE_DIR/agent"
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "$APP_DIR/Contents/MacOS/monitor" main.go

# Copy Python web components
echo "Copying Python web components..."
cp -r "$BASE_DIR/web" "$APP_DIR/Contents/Resources/"
cp -r "$BASE_DIR/config" "$APP_DIR/Contents/Resources/"
cp "$BASE_DIR/debug_tools.py" "$APP_DIR/Contents/Resources/"

# Create embedded Python environment
echo "Setting up embedded Python environment..."
cd "$APP_DIR/Contents/Resources/web"

# Remove existing venv and create a clean one
rm -rf venv
python3 -m venv venv --system-site-packages
source venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt

# Create launcher script
cat > "$APP_DIR/Contents/MacOS/roi-agent" << 'EOF'
#!/bin/bash

# ROI Agent Launcher Script
BUNDLE_DIR="$(dirname "$(dirname "$0")")"
RESOURCES_DIR="$BUNDLE_DIR/Resources"
CONFIG_DIR="$RESOURCES_DIR/config"
WEB_DIR="$RESOURCES_DIR/web"
MONITOR_BIN="$BUNDLE_DIR/MacOS/monitor"

# Create data and logs directories in user's home
USER_DATA_DIR="$HOME/.roiagent"
DATA_DIR="$USER_DATA_DIR/data"
LOGS_DIR="$USER_DATA_DIR/logs"

mkdir -p "$DATA_DIR"
mkdir -p "$LOGS_DIR"

# Function to check if processes are running
is_monitor_running() {
    pgrep -f "$MONITOR_BIN" > /dev/null
}

is_web_running() {
    pgrep -f "python.*app.py" > /dev/null
}

# Function to start monitor
start_monitor() {
    if ! is_monitor_running; then
        echo "Starting ROI Agent monitor..."
        cd "$USER_DATA_DIR"
        nohup "$MONITOR_BIN" > "$LOGS_DIR/monitor.log" 2>&1 &
        echo "Monitor started with PID $!"
    else
        echo "Monitor is already running"
    fi
}

# Function to start web UI
start_web() {
    if ! is_web_running; then
        echo "Starting ROI Agent web interface..."
        cd "$WEB_DIR"
        source venv/bin/activate
        export PYTHONPATH="$WEB_DIR:$PYTHONPATH"
        nohup python app.py > "$LOGS_DIR/web.log" 2>&1 &
        echo "Web interface started on http://localhost:5002"
    else
        echo "Web interface is already running"
    fi
}

# Function to stop all processes
stop_all() {
    echo "Stopping ROI Agent..."
    pkill -f "$MONITOR_BIN" || echo "Monitor was not running"
    pkill -f "python.*app.py" || echo "Web interface was not running"
    echo "ROI Agent stopped"
}

# Function to show status
show_status() {
    echo "=== ROI Agent Status ==="
    if is_monitor_running; then
        echo "✓ Monitor: Running"
    else
        echo "✗ Monitor: Not running"
    fi
    
    if is_web_running; then
        echo "✓ Web Interface: Running (http://localhost:5002)"
    else
        echo "✗ Web Interface: Not running"
    fi
    
    echo "Data directory: $DATA_DIR"
    echo "Logs directory: $LOGS_DIR"
}

# Function to open dashboard
open_dashboard() {
    if is_web_running; then
        open "http://localhost:5002"
    else
        echo "Web interface is not running. Starting it now..."
        start_web
        sleep 3
        open "http://localhost:5002"
    fi
}

# Main execution
case "$1" in
    "start")
        start_monitor
        start_web
        ;;
    "stop")
        stop_all
        ;;
    "restart")
        stop_all
        sleep 2
        start_monitor
        start_web
        ;;
    "status")
        show_status
        ;;
    "dashboard")
        open_dashboard
        ;;
    "")
        # Default behavior - start services and open dashboard
        start_monitor
        start_web
        sleep 2
        open_dashboard
        ;;
    *)
        echo "Usage: $0 [start|stop|restart|status|dashboard]"
        echo "  start     - Start monitoring services"
        echo "  stop      - Stop all services"
        echo "  restart   - Restart all services"
        echo "  status    - Show service status"
        echo "  dashboard - Open web dashboard"
        echo "  (no args) - Start services and open dashboard"
        ;;
esac
EOF

chmod +x "$APP_DIR/Contents/MacOS/roi-agent"

# Update app.py to use user data directory
cat > "$APP_DIR/Contents/Resources/web/app.py" << 'EOF'
#!/usr/bin/env python3
"""
ROI Agent - Web UI
Flask-based web interface for viewing application usage statistics
"""

import json
import os
import subprocess
import sys
from datetime import datetime, timedelta
from pathlib import Path

from flask import Flask, render_template, jsonify, request

app = Flask(__name__)

# Configuration - Use user's home directory for data
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

class AppMonitorUI:
    def __init__(self):
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
            with open(data_file, 'r') as f:
                return json.load(f)
        except Exception as e:
            print(f"Error loading data file {data_file}: {e}")
            return None
    
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

if __name__ == '__main__':
    print("=== ROI Agent Web UI ===")
    print(f"Data directory: {DATA_DIR}")
    print(f"Agent binary: {AGENT_BINARY}")
    print("Starting web server on http://localhost:5002")
    print("Press Ctrl+C to stop")
    
    app.run(host='127.0.0.1', port=5002, debug=False)
EOF

# Update monitor to use user data directory
cd "$BASE_DIR/agent"
cat > temp_main.go << 'EOF'
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// AppUsage represents application usage data
type AppUsage struct {
	Name           string    `json:"name"`
	ForegroundTime int64     `json:"foreground_time"`    // seconds
	BackgroundTime int64     `json:"background_time"`    // seconds
	FocusTime      int64     `json:"focus_time"`         // seconds
	LastSeen       time.Time `json:"last_seen"`
	IsActive       bool      `json:"is_active"`
	IsFocused      bool      `json:"is_focused"`
}

// DailyData represents a day's worth of application usage data
type DailyData struct {
	Date  string               `json:"date"`
	Apps  map[string]*AppUsage `json:"apps"`
	Total struct {
		ForegroundTime int64 `json:"foreground_time"`
		BackgroundTime int64 `json:"background_time"`
		FocusTime      int64 `json:"focus_time"`
	} `json:"total"`
}

// Agent represents the main monitoring agent
type Agent struct {
	dataDir    string
	dailyData  *DailyData
	lastUpdate time.Time
}

// NewAgent creates a new monitoring agent
func NewAgent() *Agent {
	// Use user's home directory
	homeDir, _ := os.UserHomeDir()
	userDataDir := filepath.Join(homeDir, ".roiagent")
	dataDir := filepath.Join(userDataDir, "data")
	
	agent := &Agent{
		dataDir: dataDir,
	}

	// Create directories if they don't exist
	os.MkdirAll(agent.dataDir, 0755)

	// Initialize daily data
	agent.initDailyData()

	return agent
}

// initDailyData initializes or loads today's data
func (a *Agent) initDailyData() {
	today := time.Now().Format("2006-01-02")
	dataFile := filepath.Join(a.dataDir, fmt.Sprintf("usage_%s.json", today))

	// Try to load existing data
	if data, err := ioutil.ReadFile(dataFile); err == nil {
		if err := json.Unmarshal(data, &a.dailyData); err == nil {
			log.Printf("Loaded existing data for %s", today)
			return
		}
	}

	// Create new daily data
	a.dailyData = &DailyData{
		Date: today,
		Apps: make(map[string]*AppUsage),
	}
	log.Printf("Created new daily data for %s", today)
}

// checkAccessibilityPermissions checks if the app has accessibility permissions
func (a *Agent) checkAccessibilityPermissions() bool {
	// Test by trying to get window information
	cmd := exec.Command("osascript", "-e", `
		tell application "System Events"
			try
				set frontApp to name of first application process whose frontmost is true
				return frontApp
			on error
				return "ERROR: No accessibility permissions"
			end try
		end tell
	`)

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Accessibility check failed: %v", err)
		return false
	}

	result := strings.TrimSpace(string(output))
	if strings.Contains(result, "ERROR") {
		log.Printf("Accessibility permissions required: %s", result)
		return false
	}

	return true
}

// requestAccessibilityPermissions prompts user to grant accessibility permissions
func (a *Agent) requestAccessibilityPermissions() {
	fmt.Println("=== macOS Accessibility Permissions Required ===")
	fmt.Println("ROI Agent needs accessibility permissions to monitor app usage.")
	fmt.Println("Please follow these steps:")
	fmt.Println("1. Go to System Preferences > Security & Privacy > Privacy")
	fmt.Println("2. Select 'Accessibility' from the left panel")
	fmt.Println("3. Click the lock icon and enter your password")
	fmt.Println("4. Add 'ROI Agent' to the list")
	fmt.Println("5. Restart ROI Agent")
	fmt.Println("================================================")

	// Try to open System Preferences
	exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Run()
}

// getRunningApps gets list of running applications with their status
func (a *Agent) getRunningApps() (map[string]bool, string, error) {
	// Get all running applications
	cmd := exec.Command("osascript", "-e", `
		tell application "System Events"
			set appList to {}
			set frontAppName to ""
			
			try
				set frontAppName to name of first application process whose frontmost is true
			end try
			
			repeat with theProcess in application processes
				if background only of theProcess is false then
					set end of appList to name of theProcess
				end if
			end repeat
			
			set AppleScript's text item delimiters to "|"
			set appListString to appList as string
			set AppleScript's text item delimiters to ""
			
			return frontAppName & ":::" & appListString
		end tell
	`)

	output, err := cmd.Output()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get running apps: %v", err)
	}

	result := strings.TrimSpace(string(output))
	parts := strings.Split(result, ":::")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("unexpected output format: %s", result)
	}

	frontmostApp := parts[0]
	appNames := strings.Split(parts[1], "|")

	apps := make(map[string]bool)
	for _, name := range appNames {
		name = strings.TrimSpace(name)
		if name != "" {
			apps[name] = true
		}
	}

	return apps, frontmostApp, nil
}

// updateAppUsage updates usage data for all apps
func (a *Agent) updateAppUsage() {
	now := time.Now()
	interval := int64(15) // 15 seconds

	// Check if it's a new day
	today := now.Format("2006-01-02")
	if a.dailyData.Date != today {
		a.saveDailyData()
		a.initDailyData()
	}

	runningApps, frontmostApp, err := a.getRunningApps()
	if err != nil {
		log.Printf("Error getting running apps: %v", err)
		return
	}

	// Update existing apps
	for appName, appData := range a.dailyData.Apps {
		wasActive := appData.IsActive
		wasFocused := appData.IsFocused

		// Check if app is still running
		isRunning := runningApps[appName]
		isFocused := (appName == frontmostApp)

		if isRunning {
			// App is running (foreground)
			if wasActive {
				appData.ForegroundTime += interval
				a.dailyData.Total.ForegroundTime += interval
			}
		} else {
			// Check if app might be running in background
			// For now, we'll assume it's not running if not in the list
			appData.IsActive = false
		}

		// Update focus time
		if wasFocused && isFocused {
			appData.FocusTime += interval
			a.dailyData.Total.FocusTime += interval
		}

		appData.IsFocused = isFocused
		appData.LastSeen = now
	}

	// Add new apps
	for appName := range runningApps {
		if _, exists := a.dailyData.Apps[appName]; !exists {
			a.dailyData.Apps[appName] = &AppUsage{
				Name:           appName,
				ForegroundTime: interval,
				BackgroundTime: 0,
				FocusTime:      0,
				LastSeen:       now,
				IsActive:       true,
				IsFocused:      (appName == frontmostApp),
			}
			a.dailyData.Total.ForegroundTime += interval

			if appName == frontmostApp {
				a.dailyData.Apps[appName].FocusTime = interval
				a.dailyData.Total.FocusTime += interval
			}
		}
	}

	log.Printf("Updated usage data for %d apps. Frontmost: %s", len(runningApps), frontmostApp)
}

// saveDailyData saves the current daily data to file
func (a *Agent) saveDailyData() {
	dataFile := filepath.Join(a.dataDir, fmt.Sprintf("usage_%s.json", a.dailyData.Date))

	data, err := json.MarshalIndent(a.dailyData, "", "  ")
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		return
	}

	if err := ioutil.WriteFile(dataFile, data, 0644); err != nil {
		log.Printf("Error saving data: %v", err)
		return
	}

	log.Printf("Saved daily data to %s", dataFile)
}

// Start begins the monitoring process
func (a *Agent) Start() {
	log.Println("Starting ROI Agent Monitor")

	// Check accessibility permissions
	if !a.checkAccessibilityPermissions() {
		a.requestAccessibilityPermissions()
		log.Println("Waiting for accessibility permissions...")
		for !a.checkAccessibilityPermissions() {
			time.Sleep(5 * time.Second)
		}
	}

	log.Println("Accessibility permissions granted. Starting monitoring...")

	// Start monitoring loop
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Initial update
	a.updateAppUsage()

	for {
		select {
		case <-ticker.C:
			a.updateAppUsage()
			a.saveDailyData()
		}
	}
}

// Status returns current agent status
func (a *Agent) Status() map[string]interface{} {
	return map[string]interface{}{
		"running":           true,
		"accessibility_ok":  a.checkAccessibilityPermissions(),
		"current_date":      a.dailyData.Date,
		"total_apps":        len(a.dailyData.Apps),
		"total_foreground":  a.dailyData.Total.ForegroundTime,
		"total_background":  a.dailyData.Total.BackgroundTime,
		"total_focus":       a.dailyData.Total.FocusTime,
		"last_update":       a.lastUpdate,
	}
}

func main() {
	agent := NewAgent()

	// Handle command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "status":
			status := agent.Status()
			data, _ := json.MarshalIndent(status, "", "  ")
			fmt.Println(string(data))
			return
		case "check-permissions":
			if agent.checkAccessibilityPermissions() {
				fmt.Println("Accessibility permissions: OK")
			} else {
				fmt.Println("Accessibility permissions: REQUIRED")
				agent.requestAccessibilityPermissions()
			}
			return
		}
	}

	// Start the agent
	agent.Start()
}
EOF

# Rebuild with updated code
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "$APP_DIR/Contents/MacOS/monitor" temp_main.go
rm temp_main.go

# Create simple icon (text-based for now)
echo "Creating application icon..."
mkdir -p icon.iconset
for size in 16 32 64 128 256 512 1024; do
    # Create a simple colored square as placeholder icon
    sips -s format png --resampleHeightWidth $size $size /System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/ExecutableBinaryIcon.icns --out icon.iconset/icon_${size}x${size}.png 2>/dev/null || true
done

# Convert to icns if iconutil is available
if command -v iconutil >/dev/null 2>&1; then
    iconutil -c icns icon.iconset -o "$APP_DIR/Contents/Resources/icon.icns" 2>/dev/null || true
fi
rm -rf icon.iconset

echo ""
echo "✅ ROI Agent.app created successfully!"
echo ""
echo "Location: $APP_DIR"
echo ""
echo "To install:"
echo "1. Copy '$BUNDLE_NAME' to /Applications/"
echo "2. Double-click to launch"
echo "3. Grant accessibility permissions when prompted"
echo ""
echo "The app will run in the background and can be controlled via:"
echo "- Double-click: Start services and open dashboard"
echo "- Terminal: /Applications/ROI\ Agent.app/Contents/MacOS/roi-agent [command]"
echo ""
