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

// CombinedData represents combined application and network usage data
type CombinedData struct {
	Date     string                        `json:"date"`
	Apps     map[string]*AppUsage          `json:"apps"`
	Network  map[string]*NetworkConnection `json:"network"`
	AppTotal struct {
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
		Date:    today,
		Apps:    make(map[string]*AppUsage),
		Network: make(map[string]*NetworkConnection),
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

func main() {
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
