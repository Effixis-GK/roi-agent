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
	Date         string                        `json:"date"`
	Apps         map[string]*AppUsage          `json:"apps"`
	Network      map[string]*NetworkConnection `json:"network"`
	AppTotal     struct {
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
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Only set if not already set (environment variables take precedence)
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
	dataDir          string
	combinedData     *CombinedData
	lastUpdate       time.Time
	activeDomains    map[string]*NetworkConnection
	domainMutex      sync.RWMutex
	tcpdumpCmd       *exec.Cmd
	tcpdumpCtx       context.Context
	tcpdumpCancel    context.CancelFunc
	lastTransmission time.Time
	transmissionInterval time.Duration
}

// NewAgent creates a new monitoring agent
func NewAgent() *Agent {
	homeDir, _ := os.UserHomeDir()
	userDataDir := filepath.Join(homeDir, ".roiagent")
	dataDir := filepath.Join(userDataDir, "data")

	// Try to load .env file first (for interval settings)
	envPaths := []string{
		".env",               // Current directory
		"./data-sender/.env", // From project root
		"../.env",            // Parent directory
		"../data-sender/.env", // Parent/data-sender
	}
	
	for _, envPath := range envPaths {
		if _, err := os.Stat(envPath); err == nil {
			log.Printf("Loading .env file from: %s", envPath)
			if err := loadEnvFile(envPath); err != nil {
				log.Printf("Warning: Failed to load .env file %s: %v", envPath, err)
			}
			break
		}
	}

	// Load transmission interval from environment variable
	intervalMinutes := 10 // default
	if intervalStr := os.Getenv("ROI_AGENT_INTERVAL_MINUTES"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			intervalMinutes = interval
			log.Printf("Using custom transmission interval: %d minutes", interval)
		} else {
			log.Printf("Warning: Invalid interval value '%s', using default %d minutes", intervalStr, intervalMinutes)
		}
	} else {
		log.Printf("Using default transmission interval: %d minutes", intervalMinutes)
	}

	agent := &Agent{
		dataDir:       dataDir,
		activeDomains: make(map[string]*NetworkConnection),
		transmissionInterval: time.Duration(intervalMinutes) * time.Minute,
		lastTransmission: time.Now(),
	}

	os.MkdirAll(agent.dataDir, 0755)
	agent.initCombinedData()

	return agent
}

// initCombinedData initializes today's data (clean start but allow Web UI to read existing data)
func (a *Agent) initCombinedData() {
	today := time.Now().Format("2006-01-02")

	// Create fresh data structure for the agent - don't load past data
	// But allow Web UI to read the saved files
	a.combinedData = &CombinedData{
		Date:    today,
		Apps:    make(map[string]*AppUsage),
		Network: make(map[string]*NetworkConnection),
	}
	log.Printf("Initialized fresh agent data for %s (Web UI can read saved files)", today)
}

// checkAccessibilityPermissions checks if the app has accessibility permissions
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
		log.Printf("Accessibility check failed: %v", err)
		return false
	}

	result := strings.TrimSpace(string(output))
	return !strings.Contains(result, "ERROR")
}

// getRunningApps gets list of running applications with their status
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

// startTcpdumpDNSMonitoring starts tcpdump-based DNS monitoring
func (a *Agent) startTcpdumpDNSMonitoring() error {
	if a.tcpdumpCmd != nil {
		return nil // Already running
	}

	a.tcpdumpCtx, a.tcpdumpCancel = context.WithCancel(context.Background())
	
	// Start tcpdump to capture DNS queries
	a.tcpdumpCmd = exec.CommandContext(a.tcpdumpCtx, "sudo", "tcpdump", "-i", "any", "port", "53", "-l", "-n", "-t")
	
	stdout, err := a.tcpdumpCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := a.tcpdumpCmd.Start(); err != nil {
		return fmt.Errorf("failed to start tcpdump: %v", err)
	}

	log.Println("Started DNS monitoring with tcpdump")

	// Start goroutine to process tcpdump output
	go func() {
		defer stdout.Close()
		scanner := bufio.NewScanner(stdout)
		
		for scanner.Scan() {
			line := scanner.Text()
			a.processDNSLine(line)
		}
		
		if err := scanner.Err(); err != nil {
			log.Printf("DNS monitoring scanner error: %v", err)
		}
	}()

	return nil
}

// stopTcpdumpDNSMonitoring stops tcpdump monitoring
func (a *Agent) stopTcpdumpDNSMonitoring() {
	if a.tcpdumpCancel != nil {
		a.tcpdumpCancel()
		a.tcpdumpCancel = nil
	}
	
	if a.tcpdumpCmd != nil {
		a.tcpdumpCmd = nil
	}
	
	log.Println("Stopped DNS monitoring")
}

// processDNSLine processes a single line from tcpdump DNS output
func (a *Agent) processDNSLine(line string) {
	// Parse tcpdump DNS query lines like:
	// "14:38:50.724810 IP 192.168.0.14.49457 > cache2.itscom.jp.domain: 52389+ A? www.yahoo.co.jp. (33)"
	// "14:38:50.723800 IP 192.168.0.14.62960 > cache2.itscom.jp.domain: 53501+ AAAA? www.yahoo.co.jp. (33)"
	
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
		// Update existing connection
		conn.LastSeen = currentTime
		conn.IsActive = true
	} else {
		// Add new connection
		a.activeDomains[key] = &NetworkConnection{
			Domain:          fqdn,
			Port:            port,
			Protocol:        protocol,
			Duration:        0,
			FirstSeen:       currentTime,
			LastSeen:        currentTime,
			IsActive:        true,
			AppName:         "Unknown", // Will be determined by association with app activity
			ConnectionState: "DNS_QUERY",
		}
		log.Printf("DNS Query detected: %s:%d (%s)", fqdn, port, protocol)
	}
}

// extractFQDNAndPortFromDNSQuery extracts FQDN and inferred port from tcpdump DNS query output
// Excludes CNAME records and only returns final A/AAAA record FQDNs
func (a *Agent) extractFQDNAndPortFromDNSQuery(line string) (string, int) {
	// Look for DNS query patterns
	if !strings.Contains(line, "?") {
		return "", 0
	}
	
	// Skip CNAME queries - we only want A and AAAA records
	if strings.Contains(line, "CNAME?") {
		return "", 0
	}
	
	// Match A or AAAA queries only (excludes CNAME)
	patterns := []string{
		` A\? ([a-zA-Z0-9.-]+)\.`,     // A record query
		` AAAA\? ([a-zA-Z0-9.-]+)\.`,  // AAAA record query
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			fqdn := matches[1]
			
			// Validate FQDN and exclude CNAME-like patterns
			if a.isValidFQDN(fqdn) && !a.isCNAMEPattern(fqdn) {
				// Infer port based on common patterns
				port := a.inferPortFromFQDN(fqdn)
				return fqdn, port
			}
		}
	}
	
	return "", 0
}

// inferPortFromFQDN infers the likely port based on FQDN patterns
// Prioritizes main user-facing ports (80/443)
func (a *Agent) inferPortFromFQDN(fqdn string) int {
	fqdnLower := strings.ToLower(fqdn)
	
	// For main websites, default to HTTPS (443) as most modern sites use it
	// HTTP (80) is less common for main sites now
	
	// Explicit HTTP patterns (rare nowadays)
	if strings.Contains(fqdnLower, "http.") ||
	   strings.Contains(fqdnLower, "insecure.") ||
	   strings.HasPrefix(fqdnLower, "demo.") ||
	   strings.HasPrefix(fqdnLower, "example.") {
		return 80 // HTTP
	}
	
	// Development/local patterns - usually HTTP
	if strings.Contains(fqdnLower, "localhost") ||
	   strings.Contains(fqdnLower, "local") ||
	   strings.Contains(fqdnLower, "dev.") ||
	   strings.Contains(fqdnLower, "staging.") ||
	   strings.Contains(fqdnLower, "test.") {
		// Check for common development ports in domain name
		if strings.Contains(fqdnLower, "3000") {
			return 3000
		} else if strings.Contains(fqdnLower, "8080") {
			return 8080
		} else if strings.Contains(fqdnLower, "5000") {
			return 5000
		} else if strings.Contains(fqdnLower, "8000") {
			return 8000
		}
		return 80 // Default HTTP for development
	}
	
	// For all main websites (yahoo.co.jp, google.com, etc.), default to HTTPS
	// This covers the vast majority of user-facing websites
	return 443
}

// isCNAMEPattern checks if a domain looks like a CNAME record
func (a *Agent) isCNAMEPattern(domain string) bool {
	// Common CNAME patterns to exclude
	cnamePatterns := []string{
		".edgekey.net",
		".edgesuite.net",
		".akamai.net",
		".akamaiedge.net",
		".cloudfront.net",
		".elb.amazonaws.com",
		".azureedge.net",
		".trafficmanager.net",
		".fastly.com",
		".cdn",
		"dualstack.",
		"edge-",
		"prod-",
		"cache-",
	}
	
	domainLower := strings.ToLower(domain)
	for _, pattern := range cnamePatterns {
		if strings.Contains(domainLower, pattern) {
			return true
		}
	}
	
	// Skip domains that look like CDN or load balancer endpoints
	if strings.Count(domain, ".") > 3 {
		// Too many subdomains, likely a CNAME
		return true
	}
	
	return false
}

// isValidFQDN checks if a string is a valid FQDN for display (user-accessed sites only)
func (a *Agent) isValidFQDN(domain string) bool {
	// Basic domain validation
	if len(domain) < 4 || len(domain) > 253 {
		return false
	}
	
	// Must contain at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}
	
	// Must not start or end with dot or hyphen
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") ||
	   strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
	}
	
	// Skip local domains and common non-internet domains
	skipDomains := []string{
		"localhost", ".local", ".lan", ".home", ".corp", ".internal",
		"_tcp", "_udp", "_tls", "_service", "_dns", "in-addr.arpa",
	}
	
	domainLower := strings.ToLower(domain)
	for _, skip := range skipDomains {
		if strings.Contains(domainLower, skip) {
			return false
		}
	}
	
	// Skip Apple/system specific domains
	applePatterns := []string{
		"apple.com", "icloud.com", "mzstatic.com", "itunes.com",
		"captive.apple.com", "dns.apple.com",
	}
	
	for _, pattern := range applePatterns {
		if strings.Contains(domainLower, pattern) {
			return false
		}
	}
	
	// Comprehensive exclusion list for non-user-facing domains
	excludePatterns := []string{
		// Advertising, tracking, and analytics
		"doubleclick.", "google-analytics.", "googletagmanager.",
		"facebook.com/tr", "connect.facebook.net", "googlesyndication.",
		"scorecardresearch.", "quantserve.", "omtrdc.net",
		"demdex.net", "adsystem.", "advertising.", "adnxs.com",
		"googleadservices.", "googletag.", "adsense.", "criteo.com",
		"adtrafficquality.google", "bat.bing.com",
		
		// Analytics and metrics
		"metrics.", "analytics.", "tracking.", "telemetry.",
		"beacons.", "stats.", "collect.", "pixel.",
		"insights.", "segment.", "mixpanel.", "hotjar.",
		"kinesis.", "amazonaws.com",
		
		// CDN and infrastructure
		"cdn.", "static.", "assets.", "media.",
		"s3.amazonaws.com", "cloudfront.net", "fastly.com",
		"akamai.", "edgecast.", "maxcdn.",
		
		// Social media widgets and embeds
		"widgets.", "embed.", "apis.google.com",
		"platform.twitter.com", "connect.facebook.net",
		
		// Security and certificates
		"ocsp.", "crl.", "certificates.",
		
		// API and technical subdomains
		"api.", "ajax.", "fonts.", "ssl.", "secure.",
		"www2.", "www3.", "m.", "mobile.",
		"admin.", "staging.", "dev.", "test.",
		"clients.", "client.", "config.", "settings.",
		
		// Image and content servers
		"images.", "img.", "thumbs.", "avatar.",
		"content.", "uploads.", "downloads.",
		
		// Version-specific or temporary subdomains
		"v1.", "v2.", "beta.", "alpha.", "preview.",
		"temp.", "tmp.", "cache.",
		
		// Yahoo specific tracking/internal domains
		"cksync.yahoo.co.jp", "clb.yahoo.co.jp", "dsb.yahoo.co.jp",
		"logql.yahoo.co.jp", "quriosity.yahoo.co.jp", "yeas.yahoo.co.jp",
		
		// Office/Microsoft internal services
		"mira-ssc.", "tmc-g2.", "tm-4.office.com",
		
		// Extension and app-specific domains
		"extension.", "grammarly.com", "walkme.com",
		
		// Cisco/Webex internal
		"wbx2.com", "apheleia-", "code42.com",
		"cloud-ec-asn.",
		
		// Spotify technical domains
		"spclient.",
		
		// Cursor/development tools
		"api2.", "api2direct.",
		
		// DeepL technical subdomains
		"ita-free.", "dict.", "s.deepl.com",
	}
	
	for _, exclude := range excludePatterns {
		if strings.Contains(domainLower, exclude) {
			return false
		}
	}
	
	// Additional check: exclude domains with too many subdomains
	// (likely infrastructure/CDN domains)
	if strings.Count(domain, ".") > 3 {
		return false
	}
	
	// Only allow domains that look like main user-facing websites
	// Must have reasonable structure: [subdomain.]domain.tld
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}
	
	// Check if it's a valid TLD pattern
	lastPart := parts[len(parts)-1]
	if len(lastPart) < 2 || len(lastPart) > 6 {
		return false
	}
	
	// Allow only main user-facing websites
	// Examples: www.yahoo.co.jp, google.com, chatgpt.com, claude.ai
	mainSitePatterns := []string{
		"www.", "mail.", "login.", "news.", "search.", "docs.",
	}
	
	// If it has a subdomain, it should be a common user-facing one
	if len(parts) > 2 {
		firstPart := parts[0]
		isMainSubdomain := false
		for _, pattern := range mainSitePatterns {
			if strings.HasPrefix(firstPart+".", pattern) {
				isMainSubdomain = true
				break
			}
		}
		
		// Also allow single-letter or short subdomains (like w.deepl.com)
		if len(firstPart) <= 2 {
			isMainSubdomain = true
		}
		
		if !isMainSubdomain {
			return false
		}
	}
	
	return true
}

// updateNetworkUsage updates network usage statistics
func (a *Agent) updateNetworkUsage() {
	interval := int64(15) // 15 seconds
	currentTime := time.Now()

	a.domainMutex.Lock()
	defer a.domainMutex.Unlock()

	// Update active connections and maintain them for longer visibility
	activeConnections := make(map[string]*NetworkConnection)
	domainSet := make(map[string]bool)

	// First, preserve existing connections from combined data
	for key, conn := range a.combinedData.Network {
		// Keep connections that were seen within the last 2 minutes
		if currentTime.Sub(conn.LastSeen) <= 2*time.Minute {
			activeConnections[key] = conn
			domainSet[conn.Domain] = true
			
			// Mark as inactive if not seen recently
			if currentTime.Sub(conn.LastSeen) > 30*time.Second {
				conn.IsActive = false
			} else {
				conn.IsActive = true
				conn.Duration += interval
				a.combinedData.NetworkTotal.TotalDuration += interval
			}
		}
	}

	// Add newly detected connections from DNS monitoring
	for key, conn := range a.activeDomains {
		if existing, exists := activeConnections[key]; exists {
			// Update existing connection
			existing.LastSeen = conn.LastSeen
			existing.IsActive = true
			existing.Duration += interval
			a.combinedData.NetworkTotal.TotalDuration += interval
		} else {
			// Add new connection
			conn.Duration = interval
			activeConnections[key] = conn
			a.combinedData.NetworkTotal.TotalDuration += interval
			a.combinedData.NetworkTotal.UniqueConnections++
		}
		domainSet[conn.Domain] = true
	}

	// Update combined data with persistent connections
	a.combinedData.Network = activeConnections
	
	// Update totals
	a.combinedData.NetworkTotal.UniqueConnections = len(activeConnections)
	a.combinedData.NetworkTotal.UniqueDomains = len(domainSet)

	log.Printf("Network update: %d total connections (%d active), %d unique domains", 
		len(activeConnections), a.countActiveConnections(activeConnections), len(domainSet))
}

// countActiveConnections counts currently active connections
func (a *Agent) countActiveConnections(connections map[string]*NetworkConnection) int {
	count := 0
	for _, conn := range connections {
		if conn.IsActive {
			count++
		}
	}
	return count
}

// updateAppUsage updates application usage data
func (a *Agent) updateAppUsage() {
	currentTime := time.Now()
	interval := int64(15)

	runningApps, frontmostApp, err := a.getRunningApps()
	if err != nil {
		log.Printf("Error getting running apps: %v", err)
		return
	}

	// Reset all apps to inactive first
	for _, appData := range a.combinedData.Apps {
		appData.IsActive = false
		appData.IsFocused = false
	}

	// Update existing apps and add new ones
	for appName := range runningApps {
		isFocused := (appName == frontmostApp)
		
		if appData, exists := a.combinedData.Apps[appName]; exists {
			// Update existing app
			appData.IsActive = true
			appData.IsFocused = isFocused
			appData.LastSeen = currentTime
			appData.ForegroundTime += interval
			a.combinedData.AppTotal.ForegroundTime += interval
			
			if isFocused {
				appData.FocusTime += interval
				a.combinedData.AppTotal.FocusTime += interval
			}
		} else {
			// Add new app
			focusTime := int64(0)
			if isFocused {
				focusTime = interval
				a.combinedData.AppTotal.FocusTime += interval
			}

			a.combinedData.Apps[appName] = &AppUsage{
				Name:           appName,
				ForegroundTime: interval,
				FocusTime:      focusTime,
				LastSeen:       currentTime,
				IsActive:       true,
				IsFocused:      isFocused,
			}
			a.combinedData.AppTotal.ForegroundTime += interval
		}
	}

	// Remove apps that haven't been active for more than 2 minutes
	for appName, appData := range a.combinedData.Apps {
		if !appData.IsActive && currentTime.Sub(appData.LastSeen) > 2*time.Minute {
			delete(a.combinedData.Apps, appName)
		}
	}

	activeApps := 0
	for _, appData := range a.combinedData.Apps {
		if appData.IsActive {
			activeApps++
		}
	}

	log.Printf("App update: %d active apps (total: %d), frontmost: %s", 
		activeApps, len(a.combinedData.Apps), frontmostApp)
}

// triggerDataTransmission triggers data transmission if interval has passed
func (a *Agent) triggerDataTransmission() {
	if time.Since(a.lastTransmission) >= a.transmissionInterval {
		log.Println("Triggering data transmission...")
		
		// Save current data before transmission
		a.saveCombinedData()
		
		// Execute data sender using built binary
		go func() {
			// Get the current working directory and find project root
			wd, err := os.Getwd()
			if err != nil {
				log.Printf("Error getting working directory: %v", err)
				return
			}
			
			// Look for project root (directory containing data-sender folder)
			projectRoot := wd
			for {
				dataSenderDir := filepath.Join(projectRoot, "data-sender")
				if _, err := os.Stat(dataSenderDir); err == nil {
					break // Found project root
				}
				parent := filepath.Dir(projectRoot)
				if parent == projectRoot {
					// Reached filesystem root without finding data-sender
					log.Printf("Error: Could not find project root with data-sender directory")
					return
				}
				projectRoot = parent
			}
			
			dataSenderDir := filepath.Join(projectRoot, "data-sender")
			dataSenderBinary := filepath.Join(dataSenderDir, "data-sender")
			
			// Build data-sender if binary doesn't exist or source is newer
			needsBuild := true
			if info, err := os.Stat(dataSenderBinary); err == nil {
				// Check if any .go file is newer than binary
				binaryTime := info.ModTime()
				needsBuild = false
				
				files, _ := filepath.Glob(filepath.Join(dataSenderDir, "*.go"))
				for _, file := range files {
					if fileInfo, err := os.Stat(file); err == nil {
						if fileInfo.ModTime().After(binaryTime) {
							needsBuild = true
							break
						}
					}
				}
			}
			
			if needsBuild {
				log.Printf("Building data-sender binary...")
				buildCmd := exec.Command("go", "build", "-o", "data-sender", ".")
				buildCmd.Dir = dataSenderDir
				if output, err := buildCmd.CombinedOutput(); err != nil {
					log.Printf("Data-sender build error: %v, output: %s", err, string(output))
					return
				}
				log.Printf("Data-sender built successfully")
			}
			
			log.Printf("Executing data transmission: %s process", dataSenderBinary)
			
			cmd := exec.Command(dataSenderBinary, "process")
			cmd.Dir = dataSenderDir
			
			if output, err := cmd.CombinedOutput(); err != nil {
				log.Printf("Data transmission error: %v, output: %s", err, string(output))
			} else {
				log.Printf("Data transmission completed: %s", string(output))
			}
		}()
		
		a.lastTransmission = time.Now()
	}
}

// saveCombinedData saves the current combined data to file
func (a *Agent) saveCombinedData() {
	dataFile := filepath.Join(a.dataDir, fmt.Sprintf("combined_%s.json", a.combinedData.Date))

	data, err := json.MarshalIndent(a.combinedData, "", "  ")
	if err != nil {
		log.Printf("Error marshaling combined data: %v", err)
		return
	}

	if err := ioutil.WriteFile(dataFile, data, 0644); err != nil {
		log.Printf("Error saving combined data: %v", err)
		return
	}

	a.lastUpdate = time.Now()
}

// Start begins the monitoring process
func (a *Agent) Start() {
	log.Println("Starting ROI Agent with tcpdump-based DNS Monitoring")

	if !a.checkAccessibilityPermissions() {
		fmt.Println("=== macOS Accessibility Permissions Required ===")
		fmt.Println("ROI Agent needs accessibility permissions to monitor app usage.")
		fmt.Println("Please grant permissions and restart the application.")
		fmt.Println("================================================")
		return
	}

	// Start DNS monitoring
	if err := a.startTcpdumpDNSMonitoring(); err != nil {
		log.Printf("Failed to start DNS monitoring: %v", err)
		fmt.Println("=== sudo Permissions Required ===")
		fmt.Println("DNS monitoring requires sudo permissions for tcpdump.")
		fmt.Println("Please run with sudo or use the start script.")
		fmt.Println("==================================")
		return
	}

	defer a.stopTcpdumpDNSMonitoring()

	log.Println("Starting comprehensive monitoring...")

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Initial updates
	a.updateAppUsage()
	a.updateNetworkUsage()
	a.saveCombinedData()

	for {
		select {
		case <-ticker.C:
			// Check if it's a new day
			today := time.Now().Format("2006-01-02")
			if a.combinedData.Date != today {
				a.saveCombinedData()
				a.initCombinedData()
			}

			a.updateAppUsage()
			a.updateNetworkUsage()
			a.saveCombinedData()
			
			// Check for data transmission
			a.triggerDataTransmission()
		}
	}
}

// Status returns current agent status
func (a *Agent) Status() map[string]interface{} {
	activeApps := 0
	focusedApp := ""
	for _, appData := range a.combinedData.Apps {
		if appData.IsActive {
			activeApps++
		}
		if appData.IsFocused {
			focusedApp = appData.Name
		}
	}

	return map[string]interface{}{
		"running":              true,
		"accessibility_ok":     a.checkAccessibilityPermissions(),
		"dns_monitoring":       a.tcpdumpCmd != nil,
		"current_date":         a.combinedData.Date,
		"active_apps":          activeApps,
		"total_apps":           len(a.combinedData.Apps),
		"focused_app":          focusedApp,
		"active_connections":   len(a.combinedData.Network),
		"unique_domains":       a.combinedData.NetworkTotal.UniqueDomains,
		"app_foreground_time":  a.combinedData.AppTotal.ForegroundTime,
		"app_focus_time":       a.combinedData.AppTotal.FocusTime,
		"network_duration":     a.combinedData.NetworkTotal.TotalDuration,
		"last_update":          a.lastUpdate,
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
		case "test-dns":
			// Test DNS monitoring for 30 seconds
			fmt.Println("Testing DNS monitoring for 30 seconds...")
			testAgent := NewAgent()
			if err := testAgent.startTcpdumpDNSMonitoring(); err != nil {
				fmt.Printf("Error starting DNS monitoring: %v\n", err)
				return
			}
			
			time.Sleep(30 * time.Second)
			testAgent.stopTcpdumpDNSMonitoring()
			
			fmt.Printf("Detected %d connections:\n", len(testAgent.activeDomains))
			for key, conn := range testAgent.activeDomains {
				fmt.Printf("%s: %s (%s)\n", key, conn.Domain, conn.Protocol)
			}
			return
		}
	}

	agent.Start()
}
