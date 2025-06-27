package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NetworkConnection represents a network connection with FQDN
type NetworkConnection struct {
	Domain            string    `json:"domain"`             // Final FQDN after redirects
	OriginalDomain    string    `json:"original_domain"`    // Original requested domain
	Port              int       `json:"port"`
	Protocol          string    `json:"protocol"`
	BytesSent         int64     `json:"bytes_sent"`
	BytesReceived     int64     `json:"bytes_received"`
	Duration          int64     `json:"duration"`
	FirstSeen         time.Time `json:"first_seen"`
	LastSeen          time.Time `json:"last_seen"`
	IsActive          bool      `json:"is_active"`
	AppName           string    `json:"app_name"`
	LocalPort         int       `json:"local_port"`
	RemoteIP          string    `json:"remote_ip"`
	ConnectionState   string    `json:"connection_state"`
	RedirectChain     []string  `json:"redirect_chain,omitempty"`
}

// DNSQuery represents a DNS query
type DNSQuery struct {
	Domain    string    `json:"domain"`
	QueryType string    `json:"query_type"`
	Response  string    `json:"response"`
	Timestamp time.Time `json:"timestamp"`
	AppName   string    `json:"app_name"`
}

// HTTPTransaction represents an HTTP/HTTPS transaction
type HTTPTransaction struct {
	Method           string            `json:"method"`
	URL              string            `json:"url"`
	Host             string            `json:"host"`
	FinalURL         string            `json:"final_url"`         // After redirects
	FinalHost        string            `json:"final_host"`        // Final FQDN
	StatusCode       int               `json:"status_code"`
	ContentLength    int64             `json:"content_length"`
	UserAgent        string            `json:"user_agent"`
	Headers          map[string]string `json:"headers"`
	RedirectChain    []string          `json:"redirect_chain"`
	ResponseTime     time.Duration     `json:"response_time"`
	Timestamp        time.Time         `json:"timestamp"`
	AppName          string            `json:"app_name"`
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
	Date           string                        `json:"date"`
	Apps           map[string]*AppUsage          `json:"apps"`
	Network        map[string]*NetworkConnection `json:"network"`
	DNSQueries     []DNSQuery                    `json:"dns_queries"`
	HTTPTransactions []HTTPTransaction           `json:"http_transactions"`
	AppTotal       struct {
		ForegroundTime int64 `json:"foreground_time"`
		BackgroundTime int64 `json:"background_time"`
		FocusTime      int64 `json:"focus_time"`
	} `json:"app_total"`
	NetworkTotal struct {
		TotalDuration      int64 `json:"total_duration"`
		TotalBytesSent     int64 `json:"total_bytes_sent"`
		TotalBytesReceived int64 `json:"total_bytes_received"`
		UniqueConnections  int   `json:"unique_connections"`
		UniqueDomains      int   `json:"unique_domains"`
	} `json:"network_total"`
}

// Agent represents the main monitoring agent
type Agent struct {
	dataDir      string
	combinedData *CombinedData
	lastUpdate   time.Time
	dnsCache     map[string]string // IP -> FQDN cache
}

// NewAgent creates a new monitoring agent
func NewAgent() *Agent {
	homeDir, _ := os.UserHomeDir()
	userDataDir := filepath.Join(homeDir, ".roiagent")
	dataDir := filepath.Join(userDataDir, "data")

	agent := &Agent{
		dataDir:  dataDir,
		dnsCache: make(map[string]string),
	}

	os.MkdirAll(agent.dataDir, 0755)
	agent.initCombinedData()

	return agent
}

// initCombinedData initializes or loads today's data
func (a *Agent) initCombinedData() {
	today := time.Now().Format("2006-01-02")
	dataFile := filepath.Join(a.dataDir, fmt.Sprintf("combined_%s.json", today))

	if data, err := ioutil.ReadFile(dataFile); err == nil {
		if err := json.Unmarshal(data, &a.combinedData); err == nil {
			log.Printf("Loaded existing combined data for %s", today)
			return
		}
	}

	a.combinedData = &CombinedData{
		Date:             today,
		Apps:             make(map[string]*AppUsage),
		Network:          make(map[string]*NetworkConnection),
		DNSQueries:       make([]DNSQuery, 0),
		HTTPTransactions: make([]HTTPTransaction, 0),
	}
	log.Printf("Created new combined data for %s", today)
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

// captureNetworkConnections uses multiple methods to capture actual network connections
func (a *Agent) captureNetworkConnections() (map[string]*NetworkConnection, error) {
	connections := make(map[string]*NetworkConnection)

	// Method 1: Enhanced lsof for detailed connection info
	if err := a.captureLsofConnections(connections); err != nil {
		log.Printf("lsof capture failed: %v", err)
	}

	// Method 2: netstat for connection states
	if err := a.captureNetstatConnections(connections); err != nil {
		log.Printf("netstat capture failed: %v", err)
	}

	// Method 3: DNS monitoring
	if err := a.captureDNSQueries(); err != nil {
		log.Printf("DNS monitoring failed: %v", err)
	}

	return connections, nil
}

// captureLsofConnections captures detailed connection information using lsof
func (a *Agent) captureLsofConnections(connections map[string]*NetworkConnection) error {
	cmd := exec.Command("lsof", "-i", "-n", "-P", "-F", "pcn")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("lsof failed: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var currentPID, currentCommand, currentNode string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch line[0] {
		case 'p':
			currentPID = line[1:]
		case 'c':
			currentCommand = line[1:]
		case 'n':
			currentNode = line[1:]
			
			// Parse the network node information
			if currentCommand != "" && currentNode != "" {
				conn := a.parseNetworkNode(currentNode, currentCommand, currentPID)
				if conn != nil {
					key := fmt.Sprintf("%s:%d", conn.Domain, conn.Port)
					connections[key] = conn
				}
			}
		}
	}

	return nil
}

// parseNetworkNode parses network node information and resolves FQDN
func (a *Agent) parseNetworkNode(nodeInfo, appName, pid string) *NetworkConnection {
	// Parse various formats: 
	// TCP connection: *:80 (LISTEN)
	// TCP connection: 192.168.1.100:49152->140.82.112.4:443 (ESTABLISHED)
	// TCP connection: [::1]:8080 (LISTEN)

	if !strings.Contains(nodeInfo, "->") {
		return nil // Skip listening sockets
	}

	parts := strings.Split(nodeInfo, "->")
	if len(parts) != 2 {
		return nil
	}

	localPart := parts[0]
	remotePart := parts[1]

	// Extract remote IP and port
	var remoteIP string
	var port int

	// Handle IPv6 addresses
	if strings.Contains(remotePart, "[") && strings.Contains(remotePart, "]") {
		re := regexp.MustCompile(`\[([^\]]+)\]:(\d+)`)
		matches := re.FindStringSubmatch(remotePart)
		if len(matches) == 3 {
			remoteIP = matches[1]
			port, _ = strconv.Atoi(matches[2])
		}
	} else {
		// Handle IPv4 addresses
		lastColon := strings.LastIndex(remotePart, ":")
		if lastColon > 0 {
			remoteIP = remotePart[:lastColon]
			port, _ = strconv.Atoi(remotePart[lastColon+1:])
		}
	}

	if remoteIP == "" || port == 0 {
		return nil
	}

	// Skip localhost connections
	if remoteIP == "127.0.0.1" || remoteIP == "::1" || remoteIP == "localhost" {
		return nil
	}

	// Filter for HTTP/HTTPS and other interesting ports
	interestingPorts := map[int]bool{
		80: true, 443: true, 8080: true, 3000: true, 5000: true, 8000: true, 9000: true,
	}

	if !interestingPorts[port] {
		return nil
	}

	// Resolve IP to FQDN
	domain := a.resolveFQDN(remoteIP)
	if domain == "" {
		domain = remoteIP // Fallback to IP if resolution fails
	}

	// Extract local port
	var localPort int
	if strings.Contains(localPart, "[") && strings.Contains(localPart, "]") {
		re := regexp.MustCompile(`\[([^\]]+)\]:(\d+)`)
		matches := re.FindStringSubmatch(localPart)
		if len(matches) == 3 {
			localPort, _ = strconv.Atoi(matches[2])
		}
	} else {
		lastColon := strings.LastIndex(localPart, ":")
		if lastColon > 0 {
			localPort, _ = strconv.Atoi(localPart[lastColon+1:])
		}
	}

	protocol := "HTTP"
	if port == 443 {
		protocol = "HTTPS"
	}

	currentTime := time.Now()
	return &NetworkConnection{
		Domain:          domain,
		OriginalDomain:  domain, // Will be updated if we detect redirects
		Port:            port,
		Protocol:        protocol,
		FirstSeen:       currentTime,
		LastSeen:        currentTime,
		IsActive:        true,
		AppName:         appName,
		LocalPort:       localPort,
		RemoteIP:        remoteIP,
		ConnectionState: "ESTABLISHED",
	}
}

// resolveFQDN resolves IP address to FQDN with caching
func (a *Agent) resolveFQDN(ip string) string {
	// Check cache first
	if cached, exists := a.dnsCache[ip]; exists {
		return cached
	}

	// Perform reverse DNS lookup
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		// Cache the failure
		a.dnsCache[ip] = ""
		return ""
	}

	// Get the first (primary) name and remove trailing dot
	fqdn := strings.TrimSuffix(names[0], ".")
	
	// Cache the result
	a.dnsCache[ip] = fqdn
	
	log.Printf("Resolved %s -> %s", ip, fqdn)
	return fqdn
}

// captureNetstatConnections captures connection states using netstat
func (a *Agent) captureNetstatConnections(connections map[string]*NetworkConnection) error {
	cmd := exec.Command("netstat", "-an", "-p", "tcp")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("netstat failed: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if !strings.Contains(line, "ESTABLISHED") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			foreignAddr := fields[4]
			state := fields[5]

			// Parse foreign address to get IP and port
			lastColon := strings.LastIndex(foreignAddr, ":")
			if lastColon > 0 {
				ip := foreignAddr[:lastColon]
				portStr := foreignAddr[lastColon+1:]
				
				if port, err := strconv.Atoi(portStr); err == nil {
					// Find matching connection and update state
					domain := a.resolveFQDN(ip)
					if domain == "" {
						domain = ip
					}
					
					key := fmt.Sprintf("%s:%d", domain, port)
					if conn, exists := connections[key]; exists {
						conn.ConnectionState = state
					}
				}
			}
		}
	}

	return nil
}

// captureDNSQueries monitors DNS queries using system logs
func (a *Agent) captureDNSQueries() error {
	// On macOS, we can monitor DNS queries through system logs
	// This requires elevated permissions for full access
	cmd := exec.Command("log", "show", "--predicate", "subsystem == 'com.apple.network.dnsproxy'", "--style", "syslog", "--last", "1m")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to parsing DNS from established connections
		log.Printf("DNS log monitoring failed, using fallback method: %v", err)
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "query") && strings.Contains(line, "A") {
			// Extract domain from DNS query log
			// Log format varies, but typically contains the queried domain
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, ".com") || strings.Contains(part, ".org") || 
				   strings.Contains(part, ".net") || strings.Contains(part, ".io") {
					// Found a potential domain
					domain := strings.Trim(part, ".,")
					if a.isValidDomain(domain) {
						dnsQuery := DNSQuery{
							Domain:    domain,
							QueryType: "A",
							Timestamp: time.Now(),
							AppName:   "System", // Could be refined with process tracking
						}
						a.combinedData.DNSQueries = append(a.combinedData.DNSQueries, dnsQuery)
					}
					break
				}
			}
		}
	}

	return nil
}

// isValidDomain checks if a string is a valid domain name
func (a *Agent) isValidDomain(domain string) bool {
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
	
	return true
}

// followRedirects follows HTTP redirects to get final FQDN
func (a *Agent) followRedirects(originalURL string) *HTTPTransaction {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
		Timeout: 10 * time.Second,
	}

	startTime := time.Now()
	resp, err := client.Get(originalURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	responseTime := time.Since(startTime)

	// Parse original and final URLs
	originalParsed, _ := url.Parse(originalURL)
	finalURL := resp.Request.URL.String()
	finalParsed, _ := url.Parse(finalURL)

	// Build redirect chain
	redirectChain := make([]string, 0)
	if originalParsed.Host != finalParsed.Host {
		redirectChain = append(redirectChain, originalParsed.Host)
		redirectChain = append(redirectChain, finalParsed.Host)
	}

	return &HTTPTransaction{
		Method:        "GET",
		URL:           originalURL,
		Host:          originalParsed.Host,
		FinalURL:      finalURL,
		FinalHost:     finalParsed.Host,
		StatusCode:    resp.StatusCode,
		ContentLength: resp.ContentLength,
		RedirectChain: redirectChain,
		ResponseTime:  responseTime,
		Timestamp:     time.Now(),
		AppName:       "ROI-Agent-Test", // In practice, derive from connection
	}
}

// updateNetworkUsage updates network usage statistics with real data
func (a *Agent) updateNetworkUsage() {
	interval := int64(15) // 15 seconds

	connections, err := a.captureNetworkConnections()
	if err != nil {
		log.Printf("Error capturing network connections: %v", err)
		return
	}

	// Track unique domains
	domainSet := make(map[string]bool)

	// Update existing connections and add new ones
	for key, newConn := range connections {
		domainSet[newConn.Domain] = true

		if existingConn, exists := a.combinedData.Network[key]; exists {
			// Update existing connection
			existingConn.LastSeen = newConn.LastSeen
			existingConn.IsActive = newConn.IsActive
			existingConn.ConnectionState = newConn.ConnectionState
			
			if existingConn.IsActive {
				existingConn.Duration += interval
				a.combinedData.NetworkTotal.TotalDuration += interval
			}

			// Update original domain if we detected a redirect
			if newConn.OriginalDomain != newConn.Domain {
				existingConn.OriginalDomain = newConn.OriginalDomain
				existingConn.RedirectChain = newConn.RedirectChain
			}
		} else {
			// Add new connection
			newConn.Duration = interval
			a.combinedData.Network[key] = newConn
			a.combinedData.NetworkTotal.TotalDuration += interval
			a.combinedData.NetworkTotal.UniqueConnections++
		}
	}

	// Mark inactive connections
	for key, conn := range a.combinedData.Network {
		if _, stillActive := connections[key]; !stillActive {
			if time.Since(conn.LastSeen) > 30*time.Second {
				conn.IsActive = false
			}
		}
	}

	// Update domain count
	a.combinedData.NetworkTotal.UniqueDomains = len(domainSet)

	log.Printf("Updated network data: %d active connections, %d unique domains", 
		len(connections), len(domainSet))
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

	// Update existing apps
	for appName, appData := range a.combinedData.Apps {
		wasActive := appData.IsActive
		wasFocused := appData.IsFocused

		isRunning := runningApps[appName]
		isFocused := (appName == frontmostApp)

		if isRunning && wasActive {
			appData.ForegroundTime += interval
			a.combinedData.AppTotal.ForegroundTime += interval
		}

		if wasFocused && isFocused {
			appData.FocusTime += interval
			a.combinedData.AppTotal.FocusTime += interval
		}

		appData.IsFocused = isFocused
		appData.IsActive = isRunning
		appData.LastSeen = currentTime
	}

	// Add new apps
	for appName := range runningApps {
		if _, exists := a.combinedData.Apps[appName]; !exists {
			isFocused := (appName == frontmostApp)
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

	log.Printf("Updated app data: %d apps, frontmost: %s", len(runningApps), frontmostApp)
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

	log.Printf("Saved combined data to %s", dataFile)
}

// Start begins the monitoring process
func (a *Agent) Start() {
	log.Println("Starting ROI Agent with Enhanced Network Monitoring (FQDN + Real Connections)")

	if !a.checkAccessibilityPermissions() {
		fmt.Println("=== macOS Accessibility Permissions Required ===")
		fmt.Println("ROI Agent needs accessibility permissions to monitor app usage.")
		fmt.Println("Please grant permissions and restart the application.")
		fmt.Println("================================================")
		return
	}

	log.Println("Starting comprehensive monitoring with FQDN resolution...")

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Initial updates
	a.updateAppUsage()
	a.updateNetworkUsage()

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
		}
	}
}

// Status returns current agent status
func (a *Agent) Status() map[string]interface{} {
	return map[string]interface{}{
		"running":             true,
		"accessibility_ok":    a.checkAccessibilityPermissions(),
		"current_date":        a.combinedData.Date,
		"total_apps":          len(a.combinedData.Apps),
		"total_connections":   len(a.combinedData.Network),
		"unique_domains":      a.combinedData.NetworkTotal.UniqueDomains,
		"dns_cache_size":      len(a.dnsCache),
		"dns_queries":         len(a.combinedData.DNSQueries),
		"app_foreground_time": a.combinedData.AppTotal.ForegroundTime,
		"app_focus_time":      a.combinedData.AppTotal.FocusTime,
		"network_duration":    a.combinedData.NetworkTotal.TotalDuration,
		"last_update":         a.lastUpdate,
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
		case "test-fqdn":
			// Test FQDN resolution
			testAgent := NewAgent()
			fmt.Println("Testing FQDN resolution:")
			testIPs := []string{"140.82.112.4", "172.217.14.196", "13.107.42.14"}
			for _, ip := range testIPs {
				fqdn := testAgent.resolveFQDN(ip)
				fmt.Printf("%s -> %s\n", ip, fqdn)
			}
			return
		case "test-connections":
			// Test current network connections
			testAgent := NewAgent()
			fmt.Println("Testing current network connections:")
			connections, err := testAgent.captureNetworkConnections()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			for key, conn := range connections {
				fmt.Printf("%s: %s (%s) -> %s:%d\n", 
					key, conn.AppName, conn.Protocol, conn.RemoteIP, conn.Port)
			}
			return
		}
	}

	agent.Start()
}
