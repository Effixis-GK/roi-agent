package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RemoteConfigCache represents the cached remote configuration
type RemoteConfigCache struct {
	ConfigVersion   int       `json:"config_version"`
	IntervalMinutes int       `json:"interval_minutes"`
	Enabled         bool      `json:"enabled"`
	LastFetched     time.Time `json:"last_fetched"`
}

// NetworkConnection represents a simplified network connection with FQDN
type NetworkConnection struct {
	Domain          string    `json:"domain"`
	Port            int       `json:"port"`
	Protocol        string    `json:"protocol"`
	Duration        int64     `json:"duration"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	IsActive        bool      `json:"is_active"`
	AppName         string    `json:"app_name"`
	ConnectionState string    `json:"connection_state"`
}

// AppUsage represents application usage data
type AppUsage struct {
	Name           string    `json:"name"`
	ForegroundTime int64     `json:"foreground_time"`
	BackgroundTime int64     `json:"background_time"`
	FocusTime      int64     `json:"focus_time"`
	LastSeen       time.Time `json:"last_seen"`
	IsActive       bool      `json:"is_active"`
	IsFocused      bool      `json:"is_focused"`
}

// SystemMetrics represents system-wide performance metrics
type SystemMetrics struct {
	Timestamp            time.Time `json:"timestamp"`
	CPUPercent           float64   `json:"cpu_percent"`
	MemoryUsedMB         int64     `json:"memory_used_mb"`
	MemoryTotalMB        int64     `json:"memory_total_mb"`
	MemoryPercent        float64   `json:"memory_percent"`
	DiskReadMBps         float64   `json:"disk_read_mbps"`
	DiskWriteMBps        float64   `json:"disk_write_mbps"`
	DiskReadOps          int64     `json:"disk_read_ops"`
	DiskWriteOps         int64     `json:"disk_write_ops"`
	BatteryLevel         int       `json:"battery_level,omitempty"`         // 0-100
	BatteryCharging      bool      `json:"battery_charging,omitempty"`
	BatteryTimeRemaining int       `json:"battery_time_remaining,omitempty"` // minutes
	IdleTimeSec          int64     `json:"idle_time_sec"`
	ScreenLocked         bool      `json:"screen_locked"`
	ProcessCount         int       `json:"process_count"`
	SystemUptimeSec      int64     `json:"system_uptime_sec"`
	// New metrics from datadog-agent
	Load1              float64 `json:"load_1,omitempty"`
	Load5              float64 `json:"load_5,omitempty"`
	Load15             float64 `json:"load_15,omitempty"`
	LoadNorm1          float64 `json:"load_norm_1,omitempty"`
	LoadNorm5          float64 `json:"load_norm_5,omitempty"`
	LoadNorm15         float64 `json:"load_norm_15,omitempty"`
	SwapUsedMB         int64   `json:"swap_used_mb,omitempty"`
	SwapTotalMB        int64   `json:"swap_total_mb,omitempty"`
	SwapPercent        float64 `json:"swap_percent,omitempty"`
	MemoryPressure     string  `json:"memory_pressure,omitempty"`
	NetBytesRecv       uint64  `json:"net_bytes_recv,omitempty"`
	NetBytesSent       uint64  `json:"net_bytes_sent,omitempty"`
	NetPacketsRecv     uint64  `json:"net_packets_recv,omitempty"`
	NetPacketsSent     uint64  `json:"net_packets_sent,omitempty"`
	NetErrorsIn        uint64  `json:"net_errors_in,omitempty"`
	NetErrorsOut       uint64  `json:"net_errors_out,omitempty"`
	WiFiSSID           string  `json:"wifi_ssid,omitempty"`
	WiFiRSSI           int     `json:"wifi_rssi,omitempty"`
	WiFiNoise          int     `json:"wifi_noise,omitempty"`
	WiFiChannel        int     `json:"wifi_channel,omitempty"`
	WiFiTransmitRate   float64 `json:"wifi_transmit_rate,omitempty"`
	WiFiPHYMode        string  `json:"wifi_phy_mode,omitempty"`
	WiFiSignalQuality  string  `json:"wifi_signal_quality,omitempty"`
	TCPEstablished     int     `json:"tcp_established,omitempty"`
	TCPTimeWait        int     `json:"tcp_time_wait,omitempty"`
	TCPCloseWait       int     `json:"tcp_close_wait,omitempty"`
	// Additional metrics - Phase 2
	DiskUsedPercent    float64 `json:"disk_used_percent,omitempty"`
	DiskFreeGB         float64 `json:"disk_free_gb,omitempty"`
	DiskTotalGB        float64 `json:"disk_total_gb,omitempty"`
	DiskHealth         string  `json:"disk_health,omitempty"`
	OpenFileHandles    int64   `json:"open_file_handles,omitempty"`
	MaxFileHandles     int64   `json:"max_file_handles,omitempty"`
	ExternalDisplays   int     `json:"external_displays,omitempty"`
	TotalDisplays      int     `json:"total_displays,omitempty"`
	BluetoothDevices   int     `json:"bluetooth_devices,omitempty"`
	MeetingActive      bool    `json:"meeting_active,omitempty"`
	CameraInUse        bool    `json:"camera_in_use,omitempty"`
	MicrophoneInUse    bool    `json:"microphone_in_use,omitempty"`
	BrowserTabs        int     `json:"browser_tabs,omitempty"`
	FocusScore         float64 `json:"focus_score,omitempty"`
	// New fields for DB compatibility
	MeetingStatus    string `json:"meeting_status,omitempty"`    // "zoom", "teams", "slack_huddle", "google_meet", "idle"
	UserActive       bool   `json:"user_active,omitempty"`       // true if idle_time_sec < 300
	BrowserTabCount  int    `json:"browser_tab_count,omitempty"` // Total browser tabs
}

// ProcessMetrics represents per-process CPU and memory metrics
type ProcessMetrics struct {
	PID        int       `json:"pid"`
	Name       string    `json:"name"`
	CPUPercent float64   `json:"cpu_percent"`
	MemoryMB   int64     `json:"memory_mb"`
	Timestamp  time.Time `json:"timestamp"`
}

// CombinedData represents combined application and network usage data
type CombinedData struct {
	Date           string                        `json:"date"`
	Apps           map[string]*AppUsage          `json:"apps"`
	Network        map[string]*NetworkConnection `json:"network"`
	SystemMetrics  []*SystemMetrics              `json:"system_metrics"`
	ProcessMetrics []*ProcessMetrics             `json:"process_metrics"`
	AppTotal       struct {
		ForegroundTime int64 `json:"foreground_time"`
		BackgroundTime int64 `json:"background_time"`
		FocusTime      int64 `json:"focus_time"`
	} `json:"app_total"`
	NetworkTotal struct {
		TotalDuration     int64 `json:"total_duration"`
		UniqueConnections int   `json:"unique_connections"`
		UniqueDomains     int   `json:"unique_domains"`
	} `json:"network_total"`
}

// getInstallDir returns the installation directory
func getInstallDir() string {
	// Check if running from PKG installation
	if _, err := os.Stat("/Applications/ROI Agent"); err == nil {
		return "/Applications/ROI Agent"
	}

	// Development mode: use current directory
	wd, _ := os.Getwd()
	return wd
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if os.Getenv(key) == "" {
				os.Setenv(key, value)
				log.Printf("Loaded from .env: %s=%s", key, value)
			}
		}
	}

	return scanner.Err()
}

// Agent represents the main monitoring agent
type Agent struct {
	dataDir              string
	installDir           string
	combinedData         *CombinedData
	lastUpdate           time.Time
	activeDomains        map[string]*NetworkConnection
	domainMutex          sync.RWMutex
	tcpdumpCmd           *exec.Cmd
	tcpdumpCtx           context.Context
	tcpdumpCancel        context.CancelFunc
	lastTransmission     time.Time
	transmissionInterval time.Duration
	dataSenderPath       string
}

// NewAgent creates a new monitoring agent
func NewAgent() *Agent {
	homeDir, _ := os.UserHomeDir()
	userDataDir := filepath.Join(homeDir, ".roiagent")
	dataDir := filepath.Join(userDataDir, "data")

	installDir := getInstallDir()
	log.Printf("Install directory: %s", installDir)

	// Load .env file from PKG Resources or development directory
	envPaths := []string{
		filepath.Join(installDir, "Resources", ".env"),
		filepath.Join(installDir, ".env"),
		filepath.Join(installDir, "data-sender", ".env"),
	}

	envLoaded := false
	for _, envPath := range envPaths {
		if _, err := os.Stat(envPath); err == nil {
			log.Printf("Loading .env file from: %s", envPath)
			if err := loadEnvFile(envPath); err != nil {
				log.Printf("Warning: Failed to load .env file %s: %v", envPath, err)
			} else {
				envLoaded = true
				break
			}
		}
	}

	if !envLoaded {
		log.Printf("Warning: No .env file found")
	}

	intervalMinutes := 10
	if intervalStr := os.Getenv("ROI_AGENT_INTERVAL_MINUTES"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			intervalMinutes = interval
			log.Printf("Using custom transmission interval: %d minutes", interval)
		}
	}

	dataSenderPath := filepath.Join(installDir, "bin", "data-sender")
	if _, err := os.Stat(dataSenderPath); err != nil {
		dataSenderPath = filepath.Join(installDir, "data-sender", "data-sender")
	}
	log.Printf("Data sender path: %s", dataSenderPath)

	agent := &Agent{
		dataDir:              dataDir,
		installDir:           installDir,
		activeDomains:        make(map[string]*NetworkConnection),
		transmissionInterval: time.Duration(intervalMinutes) * time.Minute,
		lastTransmission:     time.Now(),
		dataSenderPath:       dataSenderPath,
	}

	os.MkdirAll(agent.dataDir, 0755)
	agent.initCombinedData()

	return agent
}

func (a *Agent) initCombinedData() {
	today := time.Now().Format("2006-01-02")
	a.combinedData = &CombinedData{
		Date:           today,
		Apps:           make(map[string]*AppUsage),
		Network:        make(map[string]*NetworkConnection),
		SystemMetrics:  make([]*SystemMetrics, 0),
		ProcessMetrics: make([]*ProcessMetrics, 0),
	}
	log.Printf("Initialized fresh agent data for %s", today)
}

func (a *Agent) checkAccessibilityPermissions() bool {
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
		return false
	}

	return !strings.Contains(strings.TrimSpace(string(output)), "ERROR")
}

func (a *Agent) getRunningApps() (map[string]bool, string, error) {
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
		return nil, "", fmt.Errorf("unexpected output format")
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

func (a *Agent) startTcpdumpDNSMonitoring() error {
	if a.tcpdumpCmd != nil {
		return nil
	}

	a.tcpdumpCtx, a.tcpdumpCancel = context.WithCancel(context.Background())

	isRoot := os.Geteuid() == 0

	var cmd *exec.Cmd
	if isRoot {
		// Running as root (LaunchDaemon) - no sudo needed
		log.Println("Running as root - starting tcpdump directly")
		cmd = exec.CommandContext(a.tcpdumpCtx, "tcpdump", "-i", "any", "port", "53", "-l", "-n", "-t")
	} else {
		// Running as normal user - use sudo
		log.Println("Running as user - starting tcpdump with sudo")
		cmd = exec.CommandContext(a.tcpdumpCtx, "sudo", "tcpdump", "-i", "any", "port", "53", "-l", "-n", "-t")
	}

	a.tcpdumpCmd = cmd

	stdout, err := a.tcpdumpCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := a.tcpdumpCmd.Start(); err != nil {
		return fmt.Errorf("failed to start tcpdump: %v", err)
	}

	log.Println("Started DNS monitoring")

	go func() {
		defer stdout.Close()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			a.processDNSLine(scanner.Text())
		}
	}()

	return nil
}

func (a *Agent) stopTcpdumpDNSMonitoring() {
	if a.tcpdumpCancel != nil {
		a.tcpdumpCancel()
	}
	log.Println("Stopped DNS monitoring")
}

func (a *Agent) processDNSLine(line string) {
	fqdn, port := a.extractFQDNAndPortFromDNSQuery(line)
	if fqdn == "" {
		return
	}

	protocol := "HTTP"
	if port == 443 {
		protocol = "HTTPS"
	}

	key := fmt.Sprintf("%s:%d", fqdn, port)
	currentTime := time.Now()

	a.domainMutex.Lock()
	defer a.domainMutex.Unlock()

	if conn, exists := a.activeDomains[key]; exists {
		conn.LastSeen = currentTime
		conn.IsActive = true
	} else {
		a.activeDomains[key] = &NetworkConnection{
			Domain:          fqdn,
			Port:            port,
			Protocol:        protocol,
			Duration:        0,
			FirstSeen:       currentTime,
			LastSeen:        currentTime,
			IsActive:        true,
			AppName:         "Unknown",
			ConnectionState: "DNS_QUERY",
		}
		log.Printf("DNS Query: %s:%d (%s)", fqdn, port, protocol)
	}
}

func (a *Agent) extractFQDNAndPortFromDNSQuery(line string) (string, int) {
	if !strings.Contains(line, "?") || strings.Contains(line, "CNAME?") {
		return "", 0
	}

	patterns := []string{
		` A\? ([a-zA-Z0-9.-]+)\.`,
		` AAAA\? ([a-zA-Z0-9.-]+)\.`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			fqdn := matches[1]
			if a.isValidFQDN(fqdn) && !a.isCNAMEPattern(fqdn) {
				return fqdn, a.inferPortFromFQDN(fqdn)
			}
		}
	}

	return "", 0
}

func (a *Agent) inferPortFromFQDN(fqdn string) int {
	fqdnLower := strings.ToLower(fqdn)

	if strings.Contains(fqdnLower, "http.") || strings.Contains(fqdnLower, "insecure.") {
		return 80
	}

	if strings.Contains(fqdnLower, "localhost") || strings.Contains(fqdnLower, "local") {
		return 80
	}

	return 443
}

func (a *Agent) isCNAMEPattern(domain string) bool {
	cnamePatterns := []string{
		".akamai.net", ".cloudfront.net", ".fastly.com", ".cdn",
	}

	domainLower := strings.ToLower(domain)
	for _, pattern := range cnamePatterns {
		if strings.Contains(domainLower, pattern) {
			return true
		}
	}

	return strings.Count(domain, ".") > 3
}

func (a *Agent) isValidFQDN(domain string) bool {
	if len(domain) < 4 || len(domain) > 253 || !strings.Contains(domain, ".") {
		return false
	}

	excludePatterns := []string{
		"localhost", ".local", "apple.com", "icloud.com",
		"doubleclick.", "analytics.", "tracking.", "cdn.",
		"ads.", "metrics.", "telemetry.",
	}

	domainLower := strings.ToLower(domain)
	for _, exclude := range excludePatterns {
		if strings.Contains(domainLower, exclude) {
			return false
		}
	}

	return true
}

func (a *Agent) updateNetworkUsage() {
	interval := int64(15)
	currentTime := time.Now()

	a.domainMutex.Lock()
	defer a.domainMutex.Unlock()

	activeConnections := make(map[string]*NetworkConnection)
	domainSet := make(map[string]bool)

	for key, conn := range a.combinedData.Network {
		if currentTime.Sub(conn.LastSeen) <= 2*time.Minute {
			activeConnections[key] = conn
			domainSet[conn.Domain] = true

			if currentTime.Sub(conn.LastSeen) > 30*time.Second {
				conn.IsActive = false
			} else {
				conn.IsActive = true
				conn.Duration += interval
			}
		}
	}

	for key, conn := range a.activeDomains {
		if existing, exists := activeConnections[key]; exists {
			existing.LastSeen = conn.LastSeen
			existing.IsActive = true
			existing.Duration += interval
		} else {
			conn.Duration = interval
			activeConnections[key] = conn
		}
		domainSet[conn.Domain] = true
	}

	a.combinedData.Network = activeConnections
	a.combinedData.NetworkTotal.UniqueConnections = len(activeConnections)
	a.combinedData.NetworkTotal.UniqueDomains = len(domainSet)
}

// collectSystemMetrics collects system-wide performance metrics
func (a *Agent) collectSystemMetrics() *SystemMetrics {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// Get CPU usage
	if cpuPercent, err := a.getCPUUsage(); err == nil {
		metrics.CPUPercent = cpuPercent
	}

	// Get memory usage
	if memUsed, memTotal, memPercent, err := a.getMemoryUsage(); err == nil {
		metrics.MemoryUsedMB = memUsed
		metrics.MemoryTotalMB = memTotal
		metrics.MemoryPercent = memPercent
	}

	// Get disk I/O
	if diskReadMBps, diskWriteMBps, diskReadOps, diskWriteOps, err := a.getDiskIO(); err == nil {
		metrics.DiskReadMBps = diskReadMBps
		metrics.DiskWriteMBps = diskWriteMBps
		metrics.DiskReadOps = diskReadOps
		metrics.DiskWriteOps = diskWriteOps
	}

	// Get battery status
	if batteryLevel, batteryCharging, batteryTimeRemaining, err := a.getBatteryStatus(); err == nil {
		metrics.BatteryLevel = batteryLevel
		metrics.BatteryCharging = batteryCharging
		metrics.BatteryTimeRemaining = batteryTimeRemaining
	}

	// Get idle time and screen lock status
	if idleTimeSec, screenLocked, err := a.getSystemIdleTime(); err == nil {
		metrics.IdleTimeSec = idleTimeSec
		metrics.ScreenLocked = screenLocked
	}

	// Get process count
	if processCount, err := a.getProcessCount(); err == nil {
		metrics.ProcessCount = processCount
	}

	// Get system uptime
	if uptimeSec, err := a.getSystemUptime(); err == nil {
		metrics.SystemUptimeSec = uptimeSec
	}

	// === New metrics from datadog-agent ===

	// Get load average
	if load, err := a.getLoadAverage(); err == nil {
		metrics.Load1 = load.Load1
		metrics.Load5 = load.Load5
		metrics.Load15 = load.Load15
		metrics.LoadNorm1 = load.LoadNorm1
		metrics.LoadNorm5 = load.LoadNorm5
		metrics.LoadNorm15 = load.LoadNorm15
	}

	// Get swap and memory pressure
	if swap, err := a.getSwapUsage(); err == nil {
		metrics.SwapUsedMB = swap.UsedMB
		metrics.SwapTotalMB = swap.TotalMB
		metrics.SwapPercent = swap.Percent
		metrics.MemoryPressure = swap.Pressure
	}

	// Get network I/O
	if netIO, err := a.getNetworkIO(); err == nil {
		metrics.NetBytesRecv = netIO.BytesRecv
		metrics.NetBytesSent = netIO.BytesSent
		metrics.NetPacketsRecv = netIO.PacketsRecv
		metrics.NetPacketsSent = netIO.PacketsSent
		metrics.NetErrorsIn = netIO.ErrorsIn
		metrics.NetErrorsOut = netIO.ErrorsOut
	}

	// Get WiFi info
	if wifi, err := a.getWiFiInfo(); err == nil {
		metrics.WiFiSSID = wifi.SSID
		metrics.WiFiRSSI = wifi.RSSI
		metrics.WiFiNoise = wifi.Noise
		metrics.WiFiChannel = wifi.Channel
		metrics.WiFiTransmitRate = wifi.TransmitRate
		metrics.WiFiPHYMode = wifi.PHYMode
		metrics.WiFiSignalQuality = wifi.SignalQuality
	}

	// Get TCP connection stats
	if tcp, err := a.getTCPStats(); err == nil {
		metrics.TCPEstablished = tcp.Established
		metrics.TCPTimeWait = tcp.TimeWait
		metrics.TCPCloseWait = tcp.CloseWait
	}

	// === Additional metrics - Phase 2 ===

	// Get disk usage
	if disk, err := a.getDiskUsage(); err == nil {
		metrics.DiskUsedPercent = disk.UsedPercent
		metrics.DiskFreeGB = disk.FreeGB
		metrics.DiskTotalGB = disk.TotalGB
		metrics.DiskHealth = disk.Health
	}

	// Get file handle usage
	if fileHandles, err := a.getFileHandles(); err == nil {
		metrics.OpenFileHandles = fileHandles.OpenFiles
		metrics.MaxFileHandles = fileHandles.MaxFiles
	}

	// Get display info
	if displays, err := a.getDisplayInfo(); err == nil {
		metrics.TotalDisplays = displays.TotalDisplays
		metrics.ExternalDisplays = displays.ExternalDisplays
	}

	// Get Bluetooth device count
	if btDevices, err := a.getBluetoothDevices(); err == nil {
		metrics.BluetoothDevices = btDevices
	}

	// Get user activity metrics
	if activity, err := a.getUserActivity(); err == nil {
		metrics.MeetingActive = activity.MeetingActive
		metrics.CameraInUse = activity.CameraInUse
		metrics.MicrophoneInUse = activity.MicrophoneInUse
		metrics.FocusScore = activity.FocusScore
		// Set meeting status based on which app is active
		metrics.MeetingStatus = activity.MeetingApp
	}

	// Get browser tabs count
	if tabs, err := a.getBrowserTabs(); err == nil {
		metrics.BrowserTabs = tabs
		metrics.BrowserTabCount = tabs // For DB compatibility
	}

	// Set UserActive based on idle time (active if idle < 5 minutes)
	metrics.UserActive = metrics.IdleTimeSec < 300

	return metrics
}

// getCPUUsage gets system-wide CPU usage percentage
func (a *Agent) getCPUUsage() (float64, error) {
	cmd := exec.Command("top", "-l", "1", "-n", "0")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU usage:") {
			// Parse: "CPU usage: 5.23% user, 2.10% sys, 92.66% idle"
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "idle" && i > 0 {
					idleStr := strings.TrimSuffix(parts[i-1], "%")
					if idle, err := strconv.ParseFloat(idleStr, 64); err == nil {
						return 100.0 - idle, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("could not parse CPU usage")
}

// getMemoryUsage gets system memory usage
func (a *Agent) getMemoryUsage() (int64, int64, float64, error) {
	// Get page size
	cmd := exec.Command("sysctl", "-n", "hw.pagesize")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, err
	}
	pageSize, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	// Get vm_stat
	cmd = exec.Command("vm_stat")
	output, err = cmd.Output()
	if err != nil {
		return 0, 0, 0, err
	}

	var freePages, inactivePages, speculativePages, wiredPages, activePages int64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Pages free:") {
			fmt.Sscanf(line, "Pages free: %d", &freePages)
		} else if strings.Contains(line, "Pages inactive:") {
			fmt.Sscanf(line, "Pages inactive: %d", &inactivePages)
		} else if strings.Contains(line, "Pages speculative:") {
			fmt.Sscanf(line, "Pages speculative: %d", &speculativePages)
		} else if strings.Contains(line, "Pages wired down:") {
			fmt.Sscanf(line, "Pages wired down: %d", &wiredPages)
		} else if strings.Contains(line, "Pages active:") {
			fmt.Sscanf(line, "Pages active: %d", &activePages)
		}
	}

	// Get total memory
	cmd = exec.Command("sysctl", "-n", "hw.memsize")
	output, err = cmd.Output()
	if err != nil {
		return 0, 0, 0, err
	}
	totalBytes, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	totalMB := totalBytes / (1024 * 1024)
	freeMB := (freePages + inactivePages + speculativePages) * pageSize / (1024 * 1024)
	usedMB := totalMB - freeMB
	usedPercent := float64(usedMB) / float64(totalMB) * 100.0

	return usedMB, totalMB, usedPercent, nil
}

// getDiskIO gets disk I/O statistics
func (a *Agent) getDiskIO() (float64, float64, int64, int64, error) {
	// Use iostat to get disk I/O - format: iostat -d -w 1 -c 2
	// This gives us two samples, we'll use the second one for current activity
	cmd := exec.Command("iostat", "-d", "-w", "1", "-c", "2")
	output, err := cmd.Output()
	if err != nil {
		// iostat might not be available, return zeros
		return 0, 0, 0, 0, nil
	}

	lines := strings.Split(string(output), "\n")
	var readKBps, writeKBps float64
	var readOps, writeOps int64

	// Parse iostat output - look for disk device lines
	// Format: "disk0           0.00    0.00    0.00    0.00"
	// or:     "disk0           123.45  456.78  1234    5678"
	foundData := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "disk") && !strings.Contains(line, "device") {
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				// parts[0] = device name, parts[1] = KB_read/s, parts[2] = KB_wrtn/s, parts[3] = reads, parts[4] = writes
				if r, err := strconv.ParseFloat(parts[1], 64); err == nil {
					readKBps += r
				}
				if w, err := strconv.ParseFloat(parts[2], 64); err == nil {
					writeKBps += w
				}
				if r, err := strconv.ParseInt(parts[3], 10, 64); err == nil {
					readOps += r
				}
				if w, err := strconv.ParseInt(parts[4], 10, 64); err == nil {
					writeOps += w
				}
				foundData = true
			}
		}
	}

	if !foundData {
		return 0, 0, 0, 0, nil
	}

	readMBps := readKBps / 1024.0
	writeMBps := writeKBps / 1024.0

	return readMBps, writeMBps, readOps, writeOps, nil
}

// getBatteryStatus gets battery information
func (a *Agent) getBatteryStatus() (int, bool, int, error) {
	cmd := exec.Command("pmset", "-g", "batt")
	output, err := cmd.Output()
	if err != nil {
		return 0, false, 0, err
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Battery") {
		// No battery (desktop Mac)
		return 0, false, 0, fmt.Errorf("no battery")
	}

	// Parse: " -InternalBattery-0 (id=123456)	100%; charged; 0:00 remaining"
	// or " -InternalBattery-0 (id=123456)	85%; AC attached; not charging"
	var level int
	var charging bool
	var timeRemaining int

	// Extract battery level
	if strings.Contains(outputStr, "%") {
		re := regexp.MustCompile(`(\d+)%`)
		matches := re.FindStringSubmatch(outputStr)
		if len(matches) >= 2 {
			level, _ = strconv.Atoi(matches[1])
		}
	}

	// Check charging status
	charging = strings.Contains(outputStr, "charging") || strings.Contains(outputStr, "AC attached")

	// Extract time remaining
	if strings.Contains(outputStr, "remaining") {
		re := regexp.MustCompile(`(\d+):(\d+) remaining`)
		matches := re.FindStringSubmatch(outputStr)
		if len(matches) >= 3 {
			hours, _ := strconv.Atoi(matches[1])
			minutes, _ := strconv.Atoi(matches[2])
			timeRemaining = hours*60 + minutes
		}
	}

	return level, charging, timeRemaining, nil
}

// getSystemIdleTime gets system idle time and screen lock status
func (a *Agent) getSystemIdleTime() (int64, bool, error) {
	// Get idle time using ioreg
	cmd := exec.Command("ioreg", "-c", "IOHIDSystem")
	output, err := cmd.Output()
	if err != nil {
		return 0, false, err
	}

	var idleTime int64
	outputStr := string(output)
	
	// Look for HIDIdleTime
	re := regexp.MustCompile(`"HIDIdleTime"=(\d+)`)
	matches := re.FindStringSubmatch(outputStr)
	if len(matches) >= 2 {
		idleTime, _ = strconv.ParseInt(matches[1], 10, 64)
		// Convert from nanoseconds to seconds
		idleTime = idleTime / 1000000000
	}

	// Check screen lock status - use a simple heuristic based on idle time
	// If idle time is very high (more than 5 minutes), likely screen is locked
	// Also check pmset assertions for more accurate detection
	screenLocked := false
	if idleTime > 300 { // 5 minutes
		screenLocked = true
	}

	// Try to get more accurate screen lock status using pmset
	cmd = exec.Command("pmset", "-g", "assertions")
	output, err = cmd.Output()
	if err == nil {
		outputStr = string(output)
		// If there's no active assertion preventing display sleep, screen might be locked
		// Or if there's an assertion preventing system sleep but not display sleep, screen might be locked
		if !strings.Contains(outputStr, "PreventUserIdleDisplaySleep") {
			screenLocked = true
		}
	}

	return idleTime, screenLocked, nil
}

// getProcessCount gets the total number of running processes
func (a *Agent) getProcessCount() (int, error) {
	cmd := exec.Command("ps", "-ax")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	// Subtract 1 for the header line
	count := len(lines) - 1
	if count < 0 {
		count = 0
	}

	return count, nil
}

// getSystemUptime gets system uptime in seconds
func (a *Agent) getSystemUptime() (int64, error) {
	cmd := exec.Command("sysctl", "-n", "kern.boottime")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Output format: { sec = 1234567890, usec = 123456 }
	outputStr := strings.TrimSpace(string(output))
	re := regexp.MustCompile(`sec = (\d+)`)
	matches := re.FindStringSubmatch(outputStr)
	if len(matches) >= 2 {
		bootTime, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return 0, err
		}
		uptime := time.Now().Unix() - bootTime
		return uptime, nil
	}

	return 0, fmt.Errorf("could not parse boot time")
}

// LoadMetrics represents load average metrics
type LoadMetrics struct {
	Load1      float64
	Load5      float64
	Load15     float64
	LoadNorm1  float64
	LoadNorm5  float64
	LoadNorm15 float64
}

// getLoadAverage gets system load average
func (a *Agent) getLoadAverage() (*LoadMetrics, error) {
	cmd := exec.Command("sysctl", "-n", "vm.loadavg")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Output format: "{ 1.23 4.56 7.89 }"
	outputStr := strings.TrimSpace(string(output))
	outputStr = strings.Trim(outputStr, "{ }")
	parts := strings.Fields(outputStr)

	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected load average format")
	}

	load1, _ := strconv.ParseFloat(parts[0], 64)
	load5, _ := strconv.ParseFloat(parts[1], 64)
	load15, _ := strconv.ParseFloat(parts[2], 64)

	// Get CPU count for normalization
	cpuCmd := exec.Command("sysctl", "-n", "hw.logicalcpu")
	cpuOutput, err := cpuCmd.Output()
	numCPU := 1.0
	if err == nil {
		if count, err := strconv.ParseFloat(strings.TrimSpace(string(cpuOutput)), 64); err == nil && count > 0 {
			numCPU = count
		}
	}

	return &LoadMetrics{
		Load1:      load1,
		Load5:      load5,
		Load15:     load15,
		LoadNorm1:  load1 / numCPU,
		LoadNorm5:  load5 / numCPU,
		LoadNorm15: load15 / numCPU,
	}, nil
}

// SwapMetrics represents swap usage metrics
type SwapMetrics struct {
	UsedMB   int64
	TotalMB  int64
	Percent  float64
	Pressure string
}

// getSwapUsage gets swap usage and memory pressure
func (a *Agent) getSwapUsage() (*SwapMetrics, error) {
	cmd := exec.Command("sysctl", "-n", "vm.swapusage")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Output format: "total = 2048.00M  used = 512.00M  free = 1536.00M"
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Fields(outputStr)

	metrics := &SwapMetrics{}

	for i := 0; i < len(parts); i++ {
		if parts[i] == "total" && i+2 < len(parts) && parts[i+1] == "=" {
			metrics.TotalMB = parseMemoryMB(parts[i+2])
		} else if parts[i] == "used" && i+2 < len(parts) && parts[i+1] == "=" {
			metrics.UsedMB = parseMemoryMB(parts[i+2])
		}
	}

	if metrics.TotalMB > 0 {
		metrics.Percent = float64(metrics.UsedMB) / float64(metrics.TotalMB) * 100
	}

	// Get memory pressure using sysctl (more reliable than memory_pressure command)
	pressureCmd := exec.Command("sysctl", "-n", "kern.memorystatus_vm_pressure_level")
	pressureOutput, err := pressureCmd.Output()
	if err == nil {
		level := strings.TrimSpace(string(pressureOutput))
		switch level {
		case "1":
			metrics.Pressure = "normal"
		case "2":
			metrics.Pressure = "warning"
		case "4":
			metrics.Pressure = "critical"
		default:
			// Fallback: try memory_pressure command
			mpCmd := exec.Command("memory_pressure")
			mpOutput, mpErr := mpCmd.Output()
			if mpErr == nil {
				mpStr := strings.ToLower(string(mpOutput))
				if strings.Contains(mpStr, "critical") {
					metrics.Pressure = "critical"
				} else if strings.Contains(mpStr, "warn") {
					metrics.Pressure = "warning"
				} else if strings.Contains(mpStr, "normal") {
					metrics.Pressure = "normal"
				} else {
					metrics.Pressure = "normal" // Default to normal if parsing fails
				}
			} else {
				metrics.Pressure = "normal"
			}
		}
	} else {
		// Fallback to memory_pressure command
		mpCmd := exec.Command("memory_pressure")
		mpOutput, mpErr := mpCmd.Output()
		if mpErr == nil {
			mpStr := strings.ToLower(string(mpOutput))
			if strings.Contains(mpStr, "critical") {
				metrics.Pressure = "critical"
			} else if strings.Contains(mpStr, "warn") {
				metrics.Pressure = "warning"
			} else {
				metrics.Pressure = "normal"
			}
		} else {
			metrics.Pressure = "normal"
		}
	}

	return metrics, nil
}

// parseMemoryMB parses memory string like "2048.00M" to MB int64
func parseMemoryMB(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if len(sizeStr) == 0 {
		return 0
	}

	unit := sizeStr[len(sizeStr)-1]
	valueStr := sizeStr[:len(sizeStr)-1]
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0
	}

	switch unit {
	case 'K', 'k':
		return int64(value / 1024)
	case 'M', 'm':
		return int64(value)
	case 'G', 'g':
		return int64(value * 1024)
	case 'T', 't':
		return int64(value * 1024 * 1024)
	default:
		return int64(value / (1024 * 1024))
	}
}

// NetworkIOMetrics represents network I/O metrics
type NetworkIOMetrics struct {
	BytesRecv   uint64
	BytesSent   uint64
	PacketsRecv uint64
	PacketsSent uint64
	ErrorsIn    uint64
	ErrorsOut   uint64
}

// getNetworkIO gets network I/O statistics
func (a *Agent) getNetworkIO() (*NetworkIOMetrics, error) {
	cmd := exec.Command("netstat", "-ib")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	metrics := &NetworkIOMetrics{}
	lines := strings.Split(string(output), "\n")

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 10 {
			continue
		}

		interfaceName := parts[0]
		// Skip loopback and virtual interfaces
		if strings.HasPrefix(interfaceName, "lo") ||
			strings.HasPrefix(interfaceName, "gif") ||
			strings.HasPrefix(interfaceName, "utun") ||
			strings.HasPrefix(interfaceName, "awdl") {
			continue
		}

		// Check if this is a Link layer line
		if len(parts) >= 3 && parts[2] == "Link" && len(parts) >= 11 {
			packetsRecv, _ := strconv.ParseUint(parts[4], 10, 64)
			errorsIn, _ := strconv.ParseUint(parts[5], 10, 64)
			bytesRecv, _ := strconv.ParseUint(parts[6], 10, 64)
			packetsSent, _ := strconv.ParseUint(parts[7], 10, 64)
			errorsOut, _ := strconv.ParseUint(parts[8], 10, 64)
			bytesSent, _ := strconv.ParseUint(parts[9], 10, 64)

			metrics.BytesRecv += bytesRecv
			metrics.BytesSent += bytesSent
			metrics.PacketsRecv += packetsRecv
			metrics.PacketsSent += packetsSent
			metrics.ErrorsIn += errorsIn
			metrics.ErrorsOut += errorsOut
		}
	}

	return metrics, nil
}

// WiFiInfo represents WiFi connection info
type WiFiInfo struct {
	SSID          string
	RSSI          int
	Noise         int
	Channel       int
	TransmitRate  float64
	PHYMode       string
	SignalQuality string
}

// getWiFiInfo gets WiFi connection information
func (a *Agent) getWiFiInfo() (*WiFiInfo, error) {
	info := &WiFiInfo{}

	// Try using system_profiler with text output (more reliable parsing)
	cmd := exec.Command("system_profiler", "SPAirPortDataType")
	output, err := cmd.Output()
	if err != nil {
		return a.getWiFiInfoFallback()
	}

	lines := strings.Split(string(output), "\n")
	inCurrentNetwork := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for "Current Network Information:" section
		if strings.Contains(trimmed, "Current Network Information:") {
			inCurrentNetwork = true
			continue
		}

		if inCurrentNetwork {
			// Parse key-value pairs
			if strings.Contains(trimmed, ":") && !strings.HasSuffix(trimmed, ":") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])

					switch key {
					case "PHY Mode":
						info.PHYMode = value
					case "Channel":
						// Parse "36 (5GHz, 80MHz)"
						chanParts := strings.Split(value, " ")
						if len(chanParts) > 0 {
							info.Channel, _ = strconv.Atoi(chanParts[0])
						}
					case "Signal / Noise":
						// Parse "-55 dBm / -88 dBm"
						snParts := strings.Split(value, "/")
						if len(snParts) >= 1 {
							rssiStr := strings.TrimSpace(snParts[0])
							rssiStr = strings.TrimSuffix(rssiStr, " dBm")
							info.RSSI, _ = strconv.Atoi(rssiStr)
						}
						if len(snParts) >= 2 {
							noiseStr := strings.TrimSpace(snParts[1])
							noiseStr = strings.TrimSuffix(noiseStr, " dBm")
							info.Noise, _ = strconv.Atoi(noiseStr)
						}
					case "Transmit Rate":
						// Parse "866 Mbps"
						rateStr := strings.TrimSuffix(value, " Mbps")
						info.TransmitRate, _ = strconv.ParseFloat(rateStr, 64)
					}
				}
			} else if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " ") {
				// This is a network name (SSID)
				if info.SSID == "" {
					info.SSID = strings.TrimSuffix(trimmed, ":")
				}
			}

			// Exit current network section when we hit a blank line or new section
			if trimmed == "" && info.SSID != "" {
				break
			}
		}
	}

	// If we couldn't get data from system_profiler, try fallback
	if info.SSID == "" {
		return a.getWiFiInfoFallback()
	}

	// Determine signal quality based on RSSI
	switch {
	case info.RSSI >= -50:
		info.SignalQuality = "Excellent"
	case info.RSSI >= -60:
		info.SignalQuality = "Good"
	case info.RSSI >= -70:
		info.SignalQuality = "Fair"
	case info.RSSI >= -80:
		info.SignalQuality = "Poor"
	default:
		if info.RSSI == 0 {
			info.SignalQuality = "Excellent" // Default if RSSI not available
		} else {
			info.SignalQuality = "Very Poor"
		}
	}

	return info, nil
}

// getWiFiInfoFallback uses networksetup as a fallback method
func (a *Agent) getWiFiInfoFallback() (*WiFiInfo, error) {
	info := &WiFiInfo{}

	// Get current WiFi network name
	cmd := exec.Command("networksetup", "-getairportnetwork", "en0")
	output, err := cmd.Output()
	if err != nil {
		// Try en1 as well
		cmd = exec.Command("networksetup", "-getairportnetwork", "en1")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("WiFi not available")
		}
	}

	// Parse output like "Current Wi-Fi Network: NetworkName"
	line := strings.TrimSpace(string(output))
	if strings.HasPrefix(line, "Current Wi-Fi Network:") {
		info.SSID = strings.TrimSpace(strings.TrimPrefix(line, "Current Wi-Fi Network:"))
	} else if strings.Contains(line, "not associated") {
		return nil, fmt.Errorf("WiFi not connected")
	}

	if info.SSID == "" {
		return nil, fmt.Errorf("could not determine WiFi SSID")
	}

	// Use wdutil for more details (may require privileges)
	// For now, set reasonable defaults
	info.SignalQuality = "Unknown"
	info.PHYMode = "Unknown"

	return info, nil
}

// TCPStats represents TCP connection statistics
type TCPStats struct {
	Established int
	TimeWait    int
	CloseWait   int
}

// getTCPStats gets TCP connection statistics
func (a *Agent) getTCPStats() (*TCPStats, error) {
	cmd := exec.Command("netstat", "-an", "-p", "tcp")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	stats := &TCPStats{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Active") || strings.HasPrefix(line, "Proto") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 6 {
			continue
		}

		state := strings.ToUpper(parts[len(parts)-1])

		switch state {
		case "ESTABLISHED":
			stats.Established++
		case "TIME_WAIT":
			stats.TimeWait++
		case "CLOSE_WAIT":
			stats.CloseWait++
		}
	}

	return stats, nil
}

// DiskUsage represents disk usage metrics
type DiskUsage struct {
	UsedPercent float64
	FreeGB      float64
	TotalGB     float64
	Health      string
}

// getDiskUsage gets root disk usage
func (a *Agent) getDiskUsage() (*DiskUsage, error) {
	cmd := exec.Command("df", "-k", "/")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected df output")
	}

	parts := strings.Fields(lines[1])
	if len(parts) < 6 {
		return nil, fmt.Errorf("unexpected df output format")
	}

	totalKB, _ := strconv.ParseFloat(parts[1], 64)
	usedKB, _ := strconv.ParseFloat(parts[2], 64)
	freeKB, _ := strconv.ParseFloat(parts[3], 64)

	disk := &DiskUsage{
		TotalGB: totalKB / (1024 * 1024),
		FreeGB:  freeKB / (1024 * 1024),
	}

	if totalKB > 0 {
		disk.UsedPercent = (usedKB / totalKB) * 100
	}

	// Determine health
	if disk.UsedPercent >= 90 {
		disk.Health = "critical"
	} else if disk.UsedPercent >= 80 {
		disk.Health = "warning"
	} else {
		disk.Health = "healthy"
	}

	return disk, nil
}

// FileHandles represents file handle metrics
type FileHandles struct {
	OpenFiles int64
	MaxFiles  int64
}

// getFileHandles gets file handle usage
func (a *Agent) getFileHandles() (*FileHandles, error) {
	handles := &FileHandles{}

	cmd := exec.Command("sysctl", "-n", "kern.num_files")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	handles.OpenFiles, _ = strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)

	cmd = exec.Command("sysctl", "-n", "kern.maxfiles")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}
	handles.MaxFiles, _ = strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)

	return handles, nil
}

// DisplayMetrics represents display configuration
type DisplayMetrics struct {
	TotalDisplays    int
	ExternalDisplays int
}

// getDisplayInfo gets connected display information
func (a *Agent) getDisplayInfo() (*DisplayMetrics, error) {
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	metrics := &DisplayMetrics{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Count displays by looking for Resolution lines
		if strings.HasPrefix(line, "Resolution:") {
			metrics.TotalDisplays++
		}
		// Count external displays
		if strings.Contains(line, "Built-In") || strings.Contains(line, "Internal") {
			// Built-in display found, don't count as external
		} else if strings.HasPrefix(line, "Resolution:") {
			metrics.ExternalDisplays++
		}
	}

	// Adjust external count (first display is usually built-in on laptops)
	if metrics.ExternalDisplays > 0 {
		metrics.ExternalDisplays--
	}

	return metrics, nil
}

// getBluetoothDevices returns the number of connected Bluetooth devices
func (a *Agent) getBluetoothDevices() (int, error) {
	cmd := exec.Command("system_profiler", "SPBluetoothDataType")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	count := 0
	lines := strings.Split(string(output), "\n")
	inConnectedSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Connected:") {
			inConnectedSection = true
			continue
		}
		if inConnectedSection && strings.HasSuffix(line, ":") &&
			!strings.Contains(line, "Address") && !strings.Contains(line, "Type") {
			count++
		}
	}

	return count, nil
}

// UserActivity represents user activity metrics
type UserActivity struct {
	MeetingActive   bool
	MeetingApp      string // "zoom", "teams", "slack", "google_meet", "facetime", "webex", "idle"
	CameraInUse     bool
	MicrophoneInUse bool
	FocusScore      float64
}

// getUserActivity gets user activity metrics
func (a *Agent) getUserActivity() (*UserActivity, error) {
	activity := &UserActivity{
		FocusScore: 80.0,    // Default focus score
		MeetingApp: "idle",  // Default status
	}

	// Meeting apps with their status names
	meetingApps := map[string]string{
		"zoom.us":         "zoom",
		"zoom":            "zoom",
		"microsoft teams": "teams",
		"teams":           "teams",
		"slack":           "slack",
		"google meet":     "google_meet",
		"facetime":        "facetime",
		"webex":           "webex",
		"cisco webex":     "webex",
	}

	cmd := exec.Command("osascript", "-e", `
		tell application "System Events"
			set appList to name of every application process whose background only is false
			set AppleScript's text item delimiters to "|"
			return appList as string
		end tell
	`)
	output, err := cmd.Output()
	if err == nil {
		apps := strings.Split(strings.TrimSpace(string(output)), "|")
		for _, app := range apps {
			appLower := strings.ToLower(strings.TrimSpace(app))
			for meetingApp, status := range meetingApps {
				if strings.Contains(appLower, meetingApp) {
					activity.MeetingActive = true
					activity.MeetingApp = status
					break
				}
			}
			if activity.MeetingActive {
				break
			}
		}
	}

	// If in a meeting, assume camera/mic might be in use and increase focus score
	if activity.MeetingActive {
		activity.CameraInUse = true
		activity.MicrophoneInUse = true
		activity.FocusScore = 90.0
	}

	return activity, nil
}

// getBrowserTabs returns total browser tabs count
func (a *Agent) getBrowserTabs() (int, error) {
	totalTabs := 0

	browsers := map[string]string{
		"Safari":        `tell application "Safari" to count of tabs of every window`,
		"Google Chrome": `tell application "Google Chrome" to count of tabs of every window`,
	}

	for browser, script := range browsers {
		// Check if browser is running
		checkCmd := exec.Command("osascript", "-e",
			fmt.Sprintf(`tell application "System Events" to (name of processes) contains "%s"`, browser))
		checkOutput, err := checkCmd.Output()
		if err != nil || !strings.Contains(string(checkOutput), "true") {
			continue
		}

		// Get tab count
		cmd := exec.Command("osascript", "-e", script)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		tabCountStr := strings.TrimSpace(string(output))
		tabCounts := strings.Split(tabCountStr, ", ")
		for _, countStr := range tabCounts {
			if count, err := strconv.Atoi(strings.TrimSpace(countStr)); err == nil {
				totalTabs += count
			}
		}
	}

	return totalTabs, nil
}

// collectProcessMetrics collects per-process CPU and memory metrics
func (a *Agent) collectProcessMetrics() []*ProcessMetrics {
	var metrics []*ProcessMetrics

	cmd := exec.Command("ps", "-ax", "-o", "pid,pcpu,pmem,rss,comm")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting process metrics: %v", err)
		return metrics
	}

	lines := strings.Split(string(output), "\n")
	// Skip header line
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		var pid int
		var cpuPercent float64
		var rssKB int64

		// Parse: "1234  5.2  2.1  123456  ProcessName"
		parts := strings.Fields(line)
		if len(parts) >= 5 {
			pid, _ = strconv.Atoi(parts[0])
			cpuPercent, _ = strconv.ParseFloat(parts[1], 64)
			// parts[2] is memPercent (not used, we use RSS instead)
			rssKB, _ = strconv.ParseInt(parts[3], 10, 64)

			// Get process name (may contain spaces, so join remaining parts)
			processName := strings.Join(parts[4:], " ")

			// Convert RSS from KB to MB
			memoryMB := rssKB / 1024

			metrics = append(metrics, &ProcessMetrics{
				PID:        pid,
				Name:       processName,
				CPUPercent: cpuPercent,
				MemoryMB:   memoryMB,
				Timestamp:  time.Now(),
			})
		}
	}

	return metrics
}

// updateSystemMetrics collects and stores system and process metrics
func (a *Agent) updateSystemMetrics() {
	// Collect system metrics
	systemMetrics := a.collectSystemMetrics()
	a.combinedData.SystemMetrics = append(a.combinedData.SystemMetrics, systemMetrics)

	// Collect process metrics
	processMetrics := a.collectProcessMetrics()
	a.combinedData.ProcessMetrics = append(a.combinedData.ProcessMetrics, processMetrics...)

	// Limit the size of metrics arrays to prevent unbounded growth
	// Keep last 576 entries (15 seconds * 576 = 144 minutes = ~2.4 hours)
	maxEntries := 576
	if len(a.combinedData.SystemMetrics) > maxEntries {
		a.combinedData.SystemMetrics = a.combinedData.SystemMetrics[len(a.combinedData.SystemMetrics)-maxEntries:]
	}

	// For process metrics, keep last 10000 entries (roughly 2-3 hours depending on process count)
	maxProcessEntries := 10000
	if len(a.combinedData.ProcessMetrics) > maxProcessEntries {
		a.combinedData.ProcessMetrics = a.combinedData.ProcessMetrics[len(a.combinedData.ProcessMetrics)-maxProcessEntries:]
	}
}

func (a *Agent) updateAppUsage() {
	currentTime := time.Now()
	interval := int64(15)

	runningApps, frontmostApp, err := a.getRunningApps()
	if err != nil {
		return
	}

	for _, appData := range a.combinedData.Apps {
		appData.IsActive = false
		appData.IsFocused = false
	}

	for appName := range runningApps {
		isFocused := (appName == frontmostApp)

		if appData, exists := a.combinedData.Apps[appName]; exists {
			appData.IsActive = true
			appData.IsFocused = isFocused
			appData.LastSeen = currentTime
			appData.ForegroundTime += interval

			if isFocused {
				appData.FocusTime += interval
			}
		} else {
			focusTime := int64(0)
			if isFocused {
				focusTime = interval
			}

			a.combinedData.Apps[appName] = &AppUsage{
				Name:           appName,
				ForegroundTime: interval,
				FocusTime:      focusTime,
				LastSeen:       currentTime,
				IsActive:       true,
				IsFocused:      isFocused,
			}
		}
	}

	for appName, appData := range a.combinedData.Apps {
		if !appData.IsActive && currentTime.Sub(appData.LastSeen) > 2*time.Minute {
			delete(a.combinedData.Apps, appName)
		}
	}
}

func (a *Agent) triggerDataTransmission() {
	// Check for remote config changes and update interval if needed
	a.updateTransmissionIntervalFromRemoteConfig()

	if time.Since(a.lastTransmission) >= a.transmissionInterval {
		log.Println("Triggering data transmission...")
		a.saveCombinedData()

		go func() {
			if _, err := os.Stat(a.dataSenderPath); err != nil {
				log.Printf("Data sender not found: %s", a.dataSenderPath)
				return
			}

			cmd := exec.Command(a.dataSenderPath, "process")
			cmd.Dir = filepath.Dir(a.dataSenderPath)
			cmd.Env = os.Environ()

			if output, err := cmd.CombinedOutput(); err != nil {
				log.Printf("Data transmission error: %v, output: %s", err, string(output))
			} else {
				log.Printf("Data transmission completed: %s", string(output))
			}
		}()

		a.lastTransmission = time.Now()
	}
}

func (a *Agent) saveCombinedData() {
	dataFile := filepath.Join(a.dataDir, fmt.Sprintf("combined_%s.json", a.combinedData.Date))
	data, err := json.MarshalIndent(a.combinedData, "", "  ")
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		return
	}

	if err := ioutil.WriteFile(dataFile, data, 0644); err != nil {
		log.Printf("Error saving data: %v", err)
		return
	}

	a.lastUpdate = time.Now()
}

func (a *Agent) Start() {
	log.Println("Starting ROI Agent")
	log.Printf("Install directory: %s", a.installDir)
	log.Printf("Data directory: %s", a.dataDir)

	if !a.checkAccessibilityPermissions() {
		fmt.Println("=== Accessibility Permissions Required ===")
		fmt.Println("Please grant permissions in System Settings")
		return
	}

	if err := a.startTcpdumpDNSMonitoring(); err != nil {
		fmt.Println("=== sudo Permissions Required ===")
		fmt.Println("Please run with sudo")
		return
	}

	defer a.stopTcpdumpDNSMonitoring()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	a.updateAppUsage()
	a.updateNetworkUsage()
	a.updateSystemMetrics()
	a.saveCombinedData()

	for {
		select {
		case <-ticker.C:
			today := time.Now().Format("2006-01-02")
			if a.combinedData.Date != today {
				a.saveCombinedData()
				a.initCombinedData()
			}

			a.updateAppUsage()
			a.updateNetworkUsage()
			a.updateSystemMetrics()
			a.saveCombinedData()
			a.triggerDataTransmission()
		}
	}
}

func (a *Agent) Status() map[string]interface{} {
	return map[string]interface{}{
		"running":            true,
		"accessibility_ok":   a.checkAccessibilityPermissions(),
		"dns_monitoring":     a.tcpdumpCmd != nil,
		"current_date":       a.combinedData.Date,
		"total_apps":         len(a.combinedData.Apps),
		"active_connections": len(a.combinedData.Network),
		"unique_domains":     a.combinedData.NetworkTotal.UniqueDomains,
		"last_update":        a.lastUpdate,
	}
}

// loadRemoteConfigCache loads the cached remote configuration to check for interval changes
func (a *Agent) loadRemoteConfigCache() *RemoteConfigCache {
	// Try multiple possible locations for the cache file
	possiblePaths := []string{
		"/var/root/.roiagent/remote_config.json",
		filepath.Join(os.Getenv("HOME"), ".roiagent", "remote_config.json"),
	}

	// Also scan /Users/* for user home directories (macOS specific)
	// This is needed because data-sender runs as the logged-in user
	// while roi-agent runs as root via LaunchDaemon
	userDirs, err := filepath.Glob("/Users/*/.roiagent/remote_config.json")
	if err == nil {
		possiblePaths = append(possiblePaths, userDirs...)
	}

	// Find the most recently modified cache file
	var bestCache *RemoteConfigCache
	var bestModTime time.Time

	for _, cachePath := range possiblePaths {
		info, err := os.Stat(cachePath)
		if err != nil {
			continue
		}

		data, err := ioutil.ReadFile(cachePath)
		if err != nil {
			continue
		}

		var cache RemoteConfigCache
		if err := json.Unmarshal(data, &cache); err != nil {
			log.Printf("Error parsing remote config cache at %s: %v", cachePath, err)
			continue
		}

		// Use the most recently modified file
		if bestCache == nil || info.ModTime().After(bestModTime) {
			bestCache = &cache
			bestModTime = info.ModTime()
		}
	}

	return bestCache
}

// updateTransmissionIntervalFromRemoteConfig checks and updates the transmission interval
func (a *Agent) updateTransmissionIntervalFromRemoteConfig() {
	cache := a.loadRemoteConfigCache()
	if cache == nil {
		return
	}

	if cache.IntervalMinutes > 0 {
		newInterval := time.Duration(cache.IntervalMinutes) * time.Minute
		if newInterval != a.transmissionInterval {
			log.Printf("Updating transmission interval from %v to %v (from remote config)",
				a.transmissionInterval, newInterval)
			a.transmissionInterval = newInterval
		}
	}
}

func main() {
	// ログを標準出力に設定（LaunchDaemonで/var/log/roiagent.logに出力するため）
	log.SetOutput(os.Stdout)
	
	agent := NewAgent()

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
			}
			return
		}
	}

	agent.Start()
}
