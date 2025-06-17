#!/usr/bin/env python3
"""
ROI Agent - Debug Tools
Debugging and troubleshooting utilities for development and testing
"""

import json
import os
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

BASE_DIR = "/Users/taktakeu/Local/GitHub/roi-agent"

class DebugTools:
    def __init__(self):
        self.base_dir = BASE_DIR
        self.agent_binary = os.path.join(BASE_DIR, "agent", "monitor")
        self.data_dir = os.path.join(BASE_DIR, "data")
        self.logs_dir = os.path.join(BASE_DIR, "logs")
    
    def check_system_requirements(self):
        """Check system requirements and permissions"""
        print("=== System Requirements Check ===")
        
        # Check macOS version
        try:
            result = subprocess.run(["sw_vers", "-productVersion"], 
                                  capture_output=True, text=True)
            print(f"✓ macOS Version: {result.stdout.strip()}")
        except:
            print("✗ Unable to detect macOS version")
        
        # Check Go installation
        try:
            result = subprocess.run(["go", "version"], 
                                  capture_output=True, text=True)
            print(f"✓ Go: {result.stdout.strip()}")
        except FileNotFoundError:
            print("✗ Go not installed - Required for building agent")
        
        # Check Python
        try:
            result = subprocess.run([sys.executable, "--version"], 
                                  capture_output=True, text=True)
            print(f"✓ Python: {result.stdout.strip()}")
        except:
            print("✗ Python check failed")
        
        # Check directories
        dirs_to_check = ["agent", "web", "config"]
        for dir_name in dirs_to_check:
            dir_path = os.path.join(self.base_dir, dir_name)
            if os.path.exists(dir_path):
                print(f"✓ Directory exists: {dir_name}")
            else:
                print(f"✗ Missing directory: {dir_name}")
        
        # Check agent binary
        if os.path.exists(self.agent_binary):
            print(f"✓ Agent binary exists: {self.agent_binary}")
        else:
            print(f"✗ Agent binary not found: {self.agent_binary}")
            print("  Run: cd agent && go build -o monitor main.go")
    
    def test_accessibility_permissions(self):
        """Test macOS accessibility permissions"""
        print("=== Accessibility Permissions Test ===")
        
        # Test using AppleScript
        test_script = '''
        tell application "System Events"
            try
                set frontApp to name of first application process whose frontmost is true
                return "SUCCESS: " & frontApp
            on error errMsg
                return "ERROR: " & errMsg
            end try
        end tell
        '''
        
        try:
            result = subprocess.run(
                ["osascript", "-e", test_script],
                capture_output=True, text=True, timeout=10
            )
            
            output = result.stdout.strip()
            if output.startswith("SUCCESS:"):
                app_name = output.replace("SUCCESS: ", "")
                print(f"✓ Accessibility permissions granted")
                print(f"  Current frontmost app: {app_name}")
            else:
                print("✗ Accessibility permissions required")
                print(f"  Error: {output}")
                print("  Please grant accessibility permissions in System Preferences")
        except subprocess.TimeoutExpired:
            print("✗ Accessibility test timed out")
        except Exception as e:
            print(f"✗ Accessibility test failed: {e}")
    
    def test_app_detection(self):
        """Test application detection functionality"""
        print("=== Application Detection Test ===")
        
        # Test running apps detection
        test_script = '''
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
        '''
        
        try:
            result = subprocess.run(
                ["osascript", "-e", test_script],
                capture_output=True, text=True, timeout=15
            )
            
            output = result.stdout.strip()
            if ":::" in output:
                parts = output.split(":::")
                frontmost = parts[0]
                apps = parts[1].split("|") if parts[1] else []
                
                print(f"✓ Detected {len(apps)} running applications")
                print(f"  Frontmost app: {frontmost}")
                print(f"  Sample apps: {', '.join(apps[:5])}")
                if len(apps) > 5:
                    print(f"  ... and {len(apps) - 5} more")
            else:
                print(f"✗ Unexpected output format: {output}")
        except Exception as e:
            print(f"✗ App detection test failed: {e}")
    
    def test_agent_communication(self):
        """Test agent binary communication"""
        print("=== Agent Communication Test ===")
        
        if not os.path.exists(self.agent_binary):
            print("✗ Agent binary not found - build it first")
            print("  To build: cd agent && go build -o monitor main.go")
            return
        
        # Test status command
        try:
            result = subprocess.run(
                [self.agent_binary, "status"],
                capture_output=True, text=True, timeout=10
            )
            
            if result.returncode == 0:
                try:
                    status_data = json.loads(result.stdout)
                    print("✓ Agent status command works")
                    print(f"  Running: {status_data.get('running', 'unknown')}")
                    print(f"  Accessibility: {status_data.get('accessibility_ok', 'unknown')}")
                    print(f"  Total apps: {status_data.get('total_apps', 'unknown')}")
                except json.JSONDecodeError:
                    print(f"✗ Invalid JSON response: {result.stdout}")
            else:
                print(f"✗ Agent status failed (exit code: {result.returncode})")
                print(f"  Error: {result.stderr}")
        except subprocess.TimeoutExpired:
            print("✗ Agent status command timed out")
        except Exception as e:
            print(f"✗ Agent communication failed: {e}")
        
        # Test permissions check
        try:
            result = subprocess.run(
                [self.agent_binary, "check-permissions"],
                capture_output=True, text=True, timeout=10
            )
            print(f"  Permission check: {result.stdout.strip()}")
        except Exception as e:
            print(f"  Permission check failed: {e}")
    
    def analyze_data_files(self):
        """Analyze collected data files"""
        print("=== Data Files Analysis ===")
        
        if not os.path.exists(self.data_dir):
            print("✗ Data directory not found")
            return
        
        data_files = [f for f in os.listdir(self.data_dir) if f.startswith("usage_") and f.endswith(".json")]
        
        if not data_files:
            print("✗ No data files found")
            print("  Start the agent to begin collecting data")
            return
        
        print(f"✓ Found {len(data_files)} data files")
        
        for filename in sorted(data_files)[-3:]:  # Show last 3 files
            filepath = os.path.join(self.data_dir, filename)
            try:
                with open(filepath, 'r') as f:
                    data = json.load(f)
                
                date = data.get('date', 'unknown')
                apps_count = len(data.get('apps', {}))
                total = data.get('total', {})
                
                print(f"  {filename}:")
                print(f"    Date: {date}")
                print(f"    Apps tracked: {apps_count}")
                print(f"    Total foreground: {total.get('foreground_time', 0)}s")
                print(f"    Total focus: {total.get('focus_time', 0)}s")
                
                if apps_count > 0:
                    # Show top 3 apps by foreground time
                    apps = data.get('apps', {})
                    sorted_apps = sorted(apps.items(), 
                                       key=lambda x: x[1].get('foreground_time', 0), 
                                       reverse=True)
                    print(f"    Top apps: {', '.join([name for name, _ in sorted_apps[:3]])}")
                
            except Exception as e:
                print(f"  {filename}: Error reading file - {e}")
    
    def generate_test_data(self):
        """Generate test data for development"""
        print("=== Generating Test Data ===")
        
        os.makedirs(self.data_dir, exist_ok=True)
        
        # Generate sample data for today
        today = datetime.now().strftime("%Y-%m-%d")
        test_data = {
            "date": today,
            "apps": {
                "Finder": {
                    "name": "Finder",
                    "foreground_time": 3600,  # 1 hour
                    "background_time": 1800,  # 30 minutes
                    "focus_time": 2700,       # 45 minutes
                    "last_seen": datetime.now().isoformat(),
                    "is_active": True,
                    "is_focused": False
                },
                "Safari": {
                    "name": "Safari",
                    "foreground_time": 7200,  # 2 hours
                    "background_time": 900,   # 15 minutes
                    "focus_time": 6300,       # 1 hour 45 minutes
                    "last_seen": datetime.now().isoformat(),
                    "is_active": True,
                    "is_focused": True
                },
                "Cursor": {
                    "name": "Cursor",
                    "foreground_time": 5400,  # 1.5 hours
                    "background_time": 600,   # 10 minutes
                    "focus_time": 4800,       # 1 hour 20 minutes
                    "last_seen": datetime.now().isoformat(),
                    "is_active": True,
                    "is_focused": False
                },
                "Terminal": {
                    "name": "Terminal",
                    "foreground_time": 2400,  # 40 minutes
                    "background_time": 300,   # 5 minutes
                    "focus_time": 2100,       # 35 minutes
                    "last_seen": datetime.now().isoformat(),
                    "is_active": True,
                    "is_focused": False
                }
            },
            "total": {
                "foreground_time": 18600,  # Total foreground
                "background_time": 3600,   # Total background
                "focus_time": 15900        # Total focus
            }
        }
        
        filename = f"usage_{today}.json"
        filepath = os.path.join(self.data_dir, filename)
        
        with open(filepath, 'w') as f:
            json.dump(test_data, f, indent=2)
        
        print(f"✓ Generated test data: {filename}")
        print(f"  Apps: {len(test_data['apps'])}")
        print(f"  Total usage: {test_data['total']['foreground_time']} seconds")
    
    def run_all_tests(self):
        """Run all diagnostic tests"""
        print("=== ROI Agent - Full Diagnostic ===")
        print(f"Timestamp: {datetime.now()}")
        print(f"Base directory: {self.base_dir}")
        print()
        
        self.check_system_requirements()
        print()
        self.test_accessibility_permissions()
        print()
        self.test_app_detection()
        print()
        self.test_agent_communication()
        print()
        self.analyze_data_files()
        print()
        
        print("=== Diagnostic Complete ===")

def main():
    tools = DebugTools()
    
    if len(sys.argv) < 2:
        print("ROI Agent - Debug Tools")
        print()
        print("Usage: python debug_tools.py <command>")
        print()
        print("Commands:")
        print("  full       - Run all diagnostic tests")
        print("  system     - Check system requirements")
        print("  permissions - Test accessibility permissions")
        print("  apps       - Test application detection")
        print("  agent      - Test agent communication")
        print("  data       - Analyze data files")
        print("  testdata   - Generate test data")
        return
    
    command = sys.argv[1]
    
    if command == "full":
        tools.run_all_tests()
    elif command == "system":
        tools.check_system_requirements()
    elif command == "permissions":
        tools.test_accessibility_permissions()
    elif command == "apps":
        tools.test_app_detection()
    elif command == "agent":
        tools.test_agent_communication()
    elif command == "data":
        tools.analyze_data_files()
    elif command == "testdata":
        tools.generate_test_data()
    else:
        print(f"Unknown command: {command}")
        print("Use 'python debug_tools.py' without arguments to see available commands")

if __name__ == "__main__":
    main()
