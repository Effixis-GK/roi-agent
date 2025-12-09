//go:build darwin
// +build darwin

// Package collectors provides system metrics collection functions for macOS
package collectors

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// UserActivityMetrics represents user activity indicators
type UserActivityMetrics struct {
	IdleTimeSec       int64   `json:"idle_time_sec"`
	ScreenLocked      bool    `json:"screen_locked"`
	ActiveWindowTitle string  `json:"active_window_title,omitempty"`
	ActiveAppName     string  `json:"active_app_name,omitempty"`
	MouseMoved        bool    `json:"mouse_moved"`        // Did mouse move recently
	KeyboardActive    bool    `json:"keyboard_active"`    // Was there keyboard input
	MeetingActive     bool    `json:"meeting_active"`     // Is user in a meeting
	AudioPlaying      bool    `json:"audio_playing"`      // Is audio output active
	CameraInUse       bool    `json:"camera_in_use"`      // Is camera being used
	MicrophoneInUse   bool    `json:"microphone_in_use"`  // Is microphone being used
}

// CollectUserActivityMetrics collects user activity indicators
func CollectUserActivityMetrics() (*UserActivityMetrics, error) {
	metrics := &UserActivityMetrics{}

	// Get idle time
	idleTime, screenLocked, err := GetIdleTime()
	if err == nil {
		metrics.IdleTimeSec = idleTime
		metrics.ScreenLocked = screenLocked
	}

	// Get active window info
	activeApp, windowTitle, err := GetActiveWindow()
	if err == nil {
		metrics.ActiveAppName = activeApp
		metrics.ActiveWindowTitle = windowTitle
	}

	// Check if in a meeting (Zoom, Teams, Meet, etc.)
	metrics.MeetingActive = IsMeetingActive()

	// Check camera usage
	metrics.CameraInUse = IsCameraInUse()

	// Check microphone usage
	metrics.MicrophoneInUse = IsMicrophoneInUse()

	// Infer activity from idle time
	metrics.MouseMoved = idleTime < 5
	metrics.KeyboardActive = idleTime < 10

	return metrics, nil
}

// GetIdleTime returns system idle time in seconds and screen lock status
func GetIdleTime() (int64, bool, error) {
	cmd := exec.Command("ioreg", "-c", "IOHIDSystem")
	output, err := cmd.Output()
	if err != nil {
		return 0, false, err
	}

	var idleTime int64
	outputStr := string(output)

	// Look for HIDIdleTime
	for _, line := range strings.Split(outputStr, "\n") {
		if strings.Contains(line, "HIDIdleTime") {
			// Extract the number
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				numStr := strings.TrimSpace(parts[1])
				if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
					idleTime = num / 1000000000 // Convert nanoseconds to seconds
					break
				}
			}
		}
	}

	// Check screen lock status
	screenLocked := false
	if idleTime > 300 { // More than 5 minutes idle
		screenLocked = true
	}

	// More accurate screen lock detection
	cmd = exec.Command("pmset", "-g", "assertions")
	output, err = cmd.Output()
	if err == nil {
		if !strings.Contains(string(output), "PreventUserIdleDisplaySleep") {
			screenLocked = true
		}
	}

	return idleTime, screenLocked, nil
}

// GetActiveWindow returns the currently active application and window title
func GetActiveWindow() (string, string, error) {
	script := `
		tell application "System Events"
			set frontApp to name of first application process whose frontmost is true
			try
				tell application process frontApp
					set windowTitle to name of window 1
				end tell
			on error
				set windowTitle to ""
			end try
			return frontApp & "|" & windowTitle
		end tell
	`
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	result := strings.TrimSpace(string(output))
	parts := strings.SplitN(result, "|", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return result, "", nil
}

// IsMeetingActive checks if user is in a video/audio meeting
func IsMeetingActive() bool {
	meetingApps := []string{
		"zoom.us", "Zoom", "Microsoft Teams", "Slack", 
		"Google Meet", "FaceTime", "Webex", "Discord",
		"Skype", "GoToMeeting",
	}

	// Get list of running applications
	cmd := exec.Command("osascript", "-e", `
		tell application "System Events"
			set appList to name of every application process whose background only is false
			set AppleScript's text item delimiters to "|"
			return appList as string
		end tell
	`)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	apps := strings.Split(strings.TrimSpace(string(output)), "|")
	for _, app := range apps {
		app = strings.TrimSpace(app)
		for _, meetingApp := range meetingApps {
			if strings.Contains(strings.ToLower(app), strings.ToLower(meetingApp)) {
				return true
			}
		}
	}

	return false
}

// IsCameraInUse checks if the camera is being used
func IsCameraInUse() bool {
	// Check if any process is using the camera
	cmd := exec.Command("log", "show", "--predicate", 
		`subsystem == "com.apple.camera" AND eventMessage CONTAINS "start"`,
		"--last", "1m", "--style", "compact")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If there's recent camera activity, it's likely in use
	return len(strings.TrimSpace(string(output))) > 0
}

// IsMicrophoneInUse checks if the microphone is being used
func IsMicrophoneInUse() bool {
	// Check audio input device usage
	cmd := exec.Command("log", "show", "--predicate",
		`subsystem == "com.apple.audio" AND eventMessage CONTAINS "input"`,
		"--last", "1m", "--style", "compact")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return len(strings.TrimSpace(string(output))) > 0
}

// BrowserTabMetrics represents browser tab statistics
type BrowserTabMetrics struct {
	TotalTabs    int            `json:"total_tabs"`
	BrowserTabs  map[string]int `json:"browser_tabs"` // browser name -> tab count
	TotalWindows int            `json:"total_windows"`
}

// CollectBrowserTabMetrics collects browser tab statistics
func CollectBrowserTabMetrics() (*BrowserTabMetrics, error) {
	metrics := &BrowserTabMetrics{
		BrowserTabs: make(map[string]int),
	}

	browsers := map[string]string{
		"Safari":        `tell application "Safari" to count of tabs of every window`,
		"Google Chrome": `tell application "Google Chrome" to count of tabs of every window`,
		"Firefox":       `tell application "Firefox" to count of tabs of every window`,
		"Arc":           `tell application "Arc" to count of tabs of every window`,
		"Brave Browser": `tell application "Brave Browser" to count of tabs of every window`,
		"Microsoft Edge": `tell application "Microsoft Edge" to count of tabs of every window`,
	}

	for browser, script := range browsers {
		// Check if browser is running first
		checkCmd := exec.Command("osascript", "-e", 
			fmt.Sprintf(`tell application "System Events" to (name of processes) contains "%s"`, browser))
		checkOutput, err := checkCmd.Output()
		if err != nil || !strings.Contains(string(checkOutput), "true") {
			continue
		}

		// Get tab count
		cmd := exec.Command("osascript", "-e", script)
		cmd.Env = append(cmd.Env, "HOME="+getHomeDir())
		
		// Set a timeout
		done := make(chan error, 1)
		var output []byte
		go func() {
			var err error
			output, err = cmd.Output()
			done <- err
		}()

		select {
		case err := <-done:
			if err == nil {
				tabCountStr := strings.TrimSpace(string(output))
				// Parse the result (might be like "5, 3, 2" for multiple windows)
				tabCounts := strings.Split(tabCountStr, ", ")
				totalTabs := 0
				for _, countStr := range tabCounts {
					if count, err := strconv.Atoi(strings.TrimSpace(countStr)); err == nil {
						totalTabs += count
					}
				}
				if totalTabs > 0 {
					metrics.BrowserTabs[browser] = totalTabs
					metrics.TotalTabs += totalTabs
					metrics.TotalWindows += len(tabCounts)
				}
			}
		case <-time.After(2 * time.Second):
			// Timeout - browser might be hanging
			continue
		}
	}

	return metrics, nil
}

// getHomeDir returns the user's home directory
func getHomeDir() string {
	cmd := exec.Command("sh", "-c", "echo $HOME")
	output, err := cmd.Output()
	if err != nil {
		return "/Users"
	}
	return strings.TrimSpace(string(output))
}

// ProductivityMetrics represents productivity indicators
type ProductivityMetrics struct {
	IsActive              bool    `json:"is_active"`               // User is actively working
	InMeeting             bool    `json:"in_meeting"`              // In a video/audio meeting
	ProductiveAppTime     int64   `json:"productive_app_time_sec"` // Time in productive apps
	DistractionAppTime    int64   `json:"distraction_app_time_sec"` // Time in distraction apps
	FocusScore            float64 `json:"focus_score"`             // 0-100 focus score
	MultitaskingLevel     string  `json:"multitasking_level"`      // low/medium/high
}

// CalculateProductivityMetrics calculates productivity indicators
func CalculateProductivityMetrics(activity *UserActivityMetrics, tabs *BrowserTabMetrics) *ProductivityMetrics {
	metrics := &ProductivityMetrics{
		IsActive:  activity.IdleTimeSec < 60,
		InMeeting: activity.MeetingActive,
	}

	// Calculate focus score based on various factors
	score := 100.0

	// Deduct for high tab count
	if tabs != nil && tabs.TotalTabs > 20 {
		score -= float64(tabs.TotalTabs - 20) * 0.5
	}

	// Deduct for long idle time
	if activity.IdleTimeSec > 300 {
		score -= 20
	}

	// Add for being in a meeting (structured time)
	if activity.MeetingActive {
		score += 10
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	metrics.FocusScore = score

	// Determine multitasking level
	if tabs != nil {
		if tabs.TotalTabs > 30 {
			metrics.MultitaskingLevel = "high"
		} else if tabs.TotalTabs > 15 {
			metrics.MultitaskingLevel = "medium"
		} else {
			metrics.MultitaskingLevel = "low"
		}
	} else {
		metrics.MultitaskingLevel = "unknown"
	}

	return metrics
}

