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
	configDir  string
	dailyData  *DailyData
	lastUpdate time.Time
}

// NewAgent creates a new monitoring agent
func NewAgent(baseDir string) *Agent {
	agent := &Agent{
		dataDir:   filepath.Join(baseDir, "data"),
		configDir: filepath.Join(baseDir, "config"),
	}

	// Create directories if they don't exist
	os.MkdirAll(agent.dataDir, 0755)
	os.MkdirAll(agent.configDir, 0755)

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
	fmt.Println("This application needs accessibility permissions to monitor app usage.")
	fmt.Println("Please follow these steps:")
	fmt.Println("1. Go to System Preferences > Security & Privacy > Privacy")
	fmt.Println("2. Select 'Accessibility' from the left panel")
	fmt.Println("3. Click the lock icon and enter your password")
	fmt.Println("4. Add this application to the list")
	fmt.Println("5. Restart this application")
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
	log.Println("Starting macOS Application Monitor Agent")

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
	baseDir := "/Users/taktakeu/Local/GitHub/roi-agent"

	agent := NewAgent(baseDir)

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
