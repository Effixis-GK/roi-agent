package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows API declarations
var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	kernel32                     = windows.NewLazySystemDLL("kernel32.dll")
	psapi                        = windows.NewLazySystemDLL("psapi.dll")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	procGetModuleBaseNameW       = psapi.NewProc("GetModuleBaseNameW")
	procEnumProcesses            = psapi.NewProc("EnumProcesses")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

// AppUsage represents application usage data
type AppUsage struct {
	Name           string    `json:"name"`
	DisplayName    string    `json:"display_name"`
	Category       string    `json:"category"`
	Vendor         string    `json:"vendor"`
	IconPath       string    `json:"icon_path"`
	WindowTitle    string    `json:"window_title"`
	ForegroundTime int64     `json:"foreground_time"` // seconds
	BackgroundTime int64     `json:"background_time"` // seconds
	FocusTime      int64     `json:"focus_time"`      // seconds
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

// getProcessName gets the process name from process ID
func getProcessName(pid uint32) (string, error) {
	handle, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		uintptr(0),
		uintptr(pid),
	)
	if handle == 0 {
		return "", fmt.Errorf("failed to open process %d", pid)
	}
	defer procCloseHandle.Call(handle)

	buf := make([]uint16, 260)
	ret, _, _ := procGetModuleBaseNameW.Call(
		handle,
		uintptr(0),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret == 0 {
		return "", fmt.Errorf("failed to get module name for process %d", pid)
	}

	return windows.UTF16ToString(buf), nil
}

// getForegroundWindow gets the currently focused window and its process name
func (a *Agent) getForegroundWindow() (string, error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return "", fmt.Errorf("no foreground window")
	}

	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return "", fmt.Errorf("failed to get process ID")
	}

	processName, err := getProcessName(pid)
	if err != nil {
		return "", err
	}

	// Remove .exe extension if present
	if strings.HasSuffix(processName, ".exe") {
		processName = processName[:len(processName)-4]
	}

	return processName, nil
}

// getRunningProcesses gets list of all running processes
func (a *Agent) getRunningProcesses() (map[string]bool, error) {
	processes := make([]uint32, 1024)
	var bytesReturned uint32

	ret, _, _ := procEnumProcesses.Call(
		uintptr(unsafe.Pointer(&processes[0])),
		uintptr(len(processes)*4),
		uintptr(unsafe.Pointer(&bytesReturned)),
	)
	if ret == 0 {
		return nil, fmt.Errorf("failed to enumerate processes")
	}

	processCount := int(bytesReturned) / 4
	runningApps := make(map[string]bool)

	for i := 0; i < processCount; i++ {
		if processes[i] == 0 {
			continue
		}

		processName, err := getProcessName(processes[i])
		if err != nil {
			continue
		}

		// Remove .exe extension if present
		if strings.HasSuffix(processName, ".exe") {
			processName = processName[:len(processName)-4]
		}

		// Filter out system processes
		if !isSystemProcess(processName) {
			runningApps[processName] = true
		}
	}

	return runningApps, nil
}

// isSystemProcess checks if a process is a system process that should be ignored
func isSystemProcess(processName string) bool {
	systemProcesses := []string{
		"System", "Registry", "smss", "csrss", "wininit", "winlogon",
		"services", "lsass", "svchost", "spoolsv", "explorer",
		"dwm", "conhost", "audiodg", "dllhost", "rundll32",
		"taskhostw", "SearchIndexer", "WmiPrvSE", "msdtc",
	}

	processNameLower := strings.ToLower(processName)
	for _, sysProc := range systemProcesses {
		if strings.ToLower(sysProc) == processNameLower {
			return true
		}
	}
	return false
}

// AppInfo represents application metadata
type AppInfo struct {
	DisplayName string
	Category    string
	Vendor      string
	IconPath    string
}

// getAppInfo returns detailed information about an application
func getAppInfo(processName string) AppInfo {
	appDatabase := map[string]AppInfo{
		// Communication Apps
		"slack": {
			DisplayName: "Slack",
			Category:    "communication",
			Vendor:      "Slack Technologies",
			IconPath:    "slack.ico",
		},
		"teams": {
			DisplayName: "Microsoft Teams",
			Category:    "communication",
			Vendor:      "Microsoft Corporation",
			IconPath:    "teams.ico",
		},
		"discord": {
			DisplayName: "Discord",
			Category:    "communication",
			Vendor:      "Discord Inc.",
			IconPath:    "discord.ico",
		},
		"zoom": {
			DisplayName: "Zoom",
			Category:    "communication",
			Vendor:      "Zoom Video Communications",
			IconPath:    "zoom.ico",
		},
		"skype": {
			DisplayName: "Skype",
			Category:    "communication",
			Vendor:      "Microsoft Corporation",
			IconPath:    "skype.ico",
		},

		// Microsoft Office Apps
		"winword": {
			DisplayName: "Microsoft Word",
			Category:    "productivity",
			Vendor:      "Microsoft Corporation",
			IconPath:    "word.ico",
		},
		"excel": {
			DisplayName: "Microsoft Excel",
			Category:    "productivity",
			Vendor:      "Microsoft Corporation",
			IconPath:    "excel.ico",
		},
		"powerpnt": {
			DisplayName: "Microsoft PowerPoint",
			Category:    "productivity",
			Vendor:      "Microsoft Corporation",
			IconPath:    "powerpoint.ico",
		},
		"outlook": {
			DisplayName: "Microsoft Outlook",
			Category:    "communication",
			Vendor:      "Microsoft Corporation",
			IconPath:    "outlook.ico",
		},
		"onenote": {
			DisplayName: "Microsoft OneNote",
			Category:    "productivity",
			Vendor:      "Microsoft Corporation",
			IconPath:    "onenote.ico",
		},

		// Development Tools
		"code": {
			DisplayName: "Visual Studio Code",
			Category:    "development",
			Vendor:      "Microsoft Corporation",
			IconPath:    "vscode.ico",
		},
		"devenv": {
			DisplayName: "Visual Studio",
			Category:    "development",
			Vendor:      "Microsoft Corporation",
			IconPath:    "visualstudio.ico",
		},
		"notepad++": {
			DisplayName: "Notepad++",
			Category:    "development",
			Vendor:      "Notepad++ Team",
			IconPath:    "notepadpp.ico",
		},
		"git": {
			DisplayName: "Git",
			Category:    "development",
			Vendor:      "Git SCM",
			IconPath:    "git.ico",
		},

		// Web Browsers
		"chrome": {
			DisplayName: "Google Chrome",
			Category:    "browser",
			Vendor:      "Google LLC",
			IconPath:    "chrome.ico",
		},
		"firefox": {
			DisplayName: "Mozilla Firefox",
			Category:    "browser",
			Vendor:      "Mozilla Foundation",
			IconPath:    "firefox.ico",
		},
		"msedge": {
			DisplayName: "Microsoft Edge",
			Category:    "browser",
			Vendor:      "Microsoft Corporation",
			IconPath:    "edge.ico",
		},
		"iexplore": {
			DisplayName: "Internet Explorer",
			Category:    "browser",
			Vendor:      "Microsoft Corporation",
			IconPath:    "ie.ico",
		},

		// Media & Entertainment
		"spotify": {
			DisplayName: "Spotify",
			Category:    "entertainment",
			Vendor:      "Spotify AB",
			IconPath:    "spotify.ico",
		},
		"vlc": {
			DisplayName: "VLC Media Player",
			Category:    "entertainment",
			Vendor:      "VideoLAN",
			IconPath:    "vlc.ico",
		},
		"wmplayer": {
			DisplayName: "Windows Media Player",
			Category:    "entertainment",
			Vendor:      "Microsoft Corporation",
			IconPath:    "wmp.ico",
		},

		// System & Utilities
		"notepad": {
			DisplayName: "Notepad",
			Category:    "utility",
			Vendor:      "Microsoft Corporation",
			IconPath:    "notepad.ico",
		},
		"calc": {
			DisplayName: "Calculator",
			Category:    "utility",
			Vendor:      "Microsoft Corporation",
			IconPath:    "calc.ico",
		},
		"cmd": {
			DisplayName: "Command Prompt",
			Category:    "utility",
			Vendor:      "Microsoft Corporation",
			IconPath:    "cmd.ico",
		},
		"powershell": {
			DisplayName: "PowerShell",
			Category:    "utility",
			Vendor:      "Microsoft Corporation",
			IconPath:    "powershell.ico",
		},
	}

	processNameLower := strings.ToLower(processName)
	if info, exists := appDatabase[processNameLower]; exists {
		return info
	}

	// Default info for unknown applications
	return AppInfo{
		DisplayName: strings.Title(processName),
		Category:    "other",
		Vendor:      "Unknown",
		IconPath:    "default.ico",
	}
}

// getWindowTitle gets the window title for the foreground window
func getWindowTitle(hwnd uintptr) string {
	buf := make([]uint16, 256)
	ret, _, _ := procGetWindowTextW.Call(
		hwnd,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret == 0 {
		return ""
	}
	return windows.UTF16ToString(buf)
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

	// Get foreground application
	foregroundApp, err := a.getForegroundWindow()
	if err != nil {
		log.Printf("Error getting foreground window: %v", err)
		foregroundApp = ""
	}

	// Get running processes
	runningApps, err := a.getRunningProcesses()
	if err != nil {
		log.Printf("Error getting running processes: %v", err)
		return
	}

	// Update existing apps
	for appName, appData := range a.dailyData.Apps {
		wasActive := appData.IsActive
		wasFocused := appData.IsFocused

		// Check if app is still running
		isRunning := runningApps[appName]
		isFocused := (appName == foregroundApp)

		if isRunning {
			// App is running (foreground or background)
			if wasActive {
				if isFocused {
					appData.ForegroundTime += interval
					a.dailyData.Total.ForegroundTime += interval
				} else {
					appData.BackgroundTime += interval
					a.dailyData.Total.BackgroundTime += interval
				}
			}
		} else {
			appData.IsActive = false
		}

		// Update focus time
		if wasFocused && isFocused {
			appData.FocusTime += interval
			a.dailyData.Total.FocusTime += interval
		}

		appData.IsActive = isRunning
		appData.IsFocused = isFocused
		appData.LastSeen = now
	}

	// Add new apps
	for appName := range runningApps {
		if _, exists := a.dailyData.Apps[appName]; !exists {
			isFocused := (appName == foregroundApp)
			appInfo := getAppInfo(appName)

			// Get window title if this is the focused app
			windowTitle := ""
			if isFocused {
				hwnd, _, _ := procGetForegroundWindow.Call()
				if hwnd != 0 {
					windowTitle = getWindowTitle(hwnd)
				}
			}

			newApp := &AppUsage{
				Name:           appName,
				DisplayName:    appInfo.DisplayName,
				Category:       appInfo.Category,
				Vendor:         appInfo.Vendor,
				IconPath:       appInfo.IconPath,
				WindowTitle:    windowTitle,
				ForegroundTime: 0,
				BackgroundTime: 0,
				FocusTime:      0,
				LastSeen:       now,
				IsActive:       true,
				IsFocused:      isFocused,
			}

			if isFocused {
				newApp.ForegroundTime = interval
				newApp.FocusTime = interval
				a.dailyData.Total.ForegroundTime += interval
				a.dailyData.Total.FocusTime += interval
			} else {
				newApp.BackgroundTime = interval
				a.dailyData.Total.BackgroundTime += interval
			}

			a.dailyData.Apps[appName] = newApp
		} else {
			// Update window title for existing focused app
			if appName == foregroundApp {
				hwnd, _, _ := procGetForegroundWindow.Call()
				if hwnd != 0 {
					a.dailyData.Apps[appName].WindowTitle = getWindowTitle(hwnd)
				}
			}
		}
	}

	log.Printf("Updated usage data for %d apps. Foreground: %s", len(runningApps), foregroundApp)
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
	log.Println("Starting Windows Application Monitor Agent")

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
		"running":          true,
		"platform":         "windows",
		"current_date":     a.dailyData.Date,
		"total_apps":       len(a.dailyData.Apps),
		"total_foreground": a.dailyData.Total.ForegroundTime,
		"total_background": a.dailyData.Total.BackgroundTime,
		"total_focus":      a.dailyData.Total.FocusTime,
		"last_update":      a.lastUpdate,
	}
}

func main() {
	// Get user's home directory
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal("Failed to get current user:", err)
	}

	baseDir := filepath.Join(currentUser.HomeDir, ".roiagent")

	agent := NewAgent(baseDir)

	// Handle command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "status":
			status := agent.Status()
			data, _ := json.MarshalIndent(status, "", "  ")
			fmt.Println(string(data))
			return
		case "start":
			fmt.Println("Starting ROI Agent...")
			agent.Start()
			return
		case "stop":
			fmt.Println("ROI Agent stop command received")
			return
		}
	}

	// Default: start the agent
	agent.Start()
}
