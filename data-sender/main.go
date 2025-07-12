package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Configuration for data transmission
type Config struct {
	BaseURL    string `json:"base_url"`
	APIKey     string `json:"api_key"`
	DeviceID   string `json:"device_id"`
	Enabled    bool   `json:"enabled"`
}

// AppData represents application usage data for transmission
type AppData struct {
	ActiveApp  string `json:"active_app"`
	FocusedApp string `json:"focused_app"`
	FocusTime  int64  `json:"focus_time_seconds"`
	Timestamp  string `json:"timestamp"`
}

// NetworkData represents network access data for transmission
type NetworkData struct {
	FQDN        string `json:"fqdn"`
	Port        int    `json:"port"`
	AccessCount int    `json:"access_count"`
	Protocol    string `json:"protocol"`
	Timestamp   string `json:"timestamp"`
}

// TransmissionPayload represents the complete data package to send
type TransmissionPayload struct {
	DeviceID     string        `json:"device_id"`
	Timestamp    string        `json:"timestamp"`
	IntervalMins int           `json:"interval_minutes"`
	Apps         []AppData     `json:"apps"`
	Networks     []NetworkData `json:"networks"`
	Metadata     struct {
		OSVersion    string `json:"os_version"`
		AgentVersion string `json:"agent_version"`
		TotalApps    int    `json:"total_apps"`
		TotalDomains int    `json:"total_domains"`
	} `json:"metadata"`
}

// CombinedData represents the local data structure (matching main.go)
type CombinedData struct {
	Date    string                     `json:"date"`
	Apps    map[string]*AppUsage       `json:"apps"`
	Network map[string]*NetworkConn    `json:"network"`
}

type AppUsage struct {
	Name           string    `json:"name"`
	ForegroundTime int64     `json:"foreground_time"`
	FocusTime      int64     `json:"focus_time"`
	LastSeen       time.Time `json:"last_seen"`
	IsActive       bool      `json:"is_active"`
	IsFocused      bool      `json:"is_focused"`
}

type NetworkConn struct {
	Domain   string    `json:"domain"`
	Port     int       `json:"port"`
	Protocol string    `json:"protocol"`
	Duration int64     `json:"duration"`
	LastSeen time.Time `json:"last_seen"`
	IsActive bool      `json:"is_active"`
}

// DataSender handles the transmission of monitoring data
type DataSender struct {
	config       Config
	dataDir      string
	transmissionDir string
	configPath   string
}

// NewDataSender creates a new data sender instance
func NewDataSender() *DataSender {
	homeDir, _ := os.UserHomeDir()
	userDataDir := filepath.Join(homeDir, ".roiagent")
	dataDir := filepath.Join(userDataDir, "data")
	transmissionDir := filepath.Join(userDataDir, "transmission")
	configPath := filepath.Join(userDataDir, "transmission_config.json")

	// Create directories
	os.MkdirAll(transmissionDir, 0755)

	sender := &DataSender{
		dataDir:         dataDir,
		transmissionDir: transmissionDir,
		configPath:      configPath,
	}

	sender.loadConfig()
	return sender
}

// loadConfig loads transmission configuration from file and environment variables
func (ds *DataSender) loadConfig() {
	// Load .env file if it exists
	envPath := filepath.Join(filepath.Dir(ds.configPath), "data-sender", ".env")
	if err := godotenv.Load(envPath); err != nil {
		// .env file is optional, so don't log error unless debug mode
		if os.Getenv("ROI_AGENT_DEBUG") == "true" {
			log.Printf("No .env file found at %s: %v", envPath, err)
		}
	}

	// Default configuration - Sample URLs/Keys
	// For Mac App, enable by default if environment variables are set
	enableByDefault := false
	if os.Getenv("ROI_AGENT_BASE_URL") != "" && os.Getenv("ROI_AGENT_API_KEY") != "" {
		enableByDefault = true
	}

	ds.config = Config{
		BaseURL:  "https://api.sample-server.com/v1/roi-agent-sample",
		APIKey:   "sample-api-key-replace-with-actual",
		DeviceID: ds.generateDeviceID(),
		Enabled:  enableByDefault,
	}

	// Override with environment variables if available
	if baseURL := os.Getenv("ROI_AGENT_BASE_URL"); baseURL != "" {
		ds.config.BaseURL = baseURL
	}
	if apiKey := os.Getenv("ROI_AGENT_API_KEY"); apiKey != "" {
		ds.config.APIKey = apiKey
	}
	if deviceID := os.Getenv("ROI_AGENT_DEVICE_ID"); deviceID != "" && deviceID != "auto-generated" {
		ds.config.DeviceID = deviceID
	}
	if enabled := os.Getenv("ROI_AGENT_ENABLED"); enabled != "" {
		if enabledBool, err := strconv.ParseBool(enabled); err == nil {
			ds.config.Enabled = enabledBool
		}
	}

	// Try to load existing config file (config file overrides environment)
	if data, err := ioutil.ReadFile(ds.configPath); err == nil {
		var fileConfig Config
		if err := json.Unmarshal(data, &fileConfig); err == nil {
			// Only override non-empty values from config file
			if fileConfig.BaseURL != "" {
				ds.config.BaseURL = fileConfig.BaseURL
			}
			if fileConfig.APIKey != "" {
				ds.config.APIKey = fileConfig.APIKey
			}
			if fileConfig.DeviceID != "" {
				ds.config.DeviceID = fileConfig.DeviceID
			}
			ds.config.Enabled = fileConfig.Enabled
		} else {
			log.Printf("Error loading config file: %v", err)
		}
	}

	// Save config (creates file if it doesn't exist)
	ds.saveConfig()
}

// saveConfig saves the current configuration
func (ds *DataSender) saveConfig() {
	data, err := json.MarshalIndent(ds.config, "", "  ")
	if err != nil {
		log.Printf("Error marshaling config: %v", err)
		return
	}

	if err := ioutil.WriteFile(ds.configPath, data, 0644); err != nil {
		log.Printf("Error saving config: %v", err)
	}
}

// generateDeviceID generates a unique device identifier
func (ds *DataSender) generateDeviceID() string {
	// Use hostname + current time for simple unique ID
	hostname, _ := os.Hostname()
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d", hostname, timestamp)
}

// ProcessData processes the latest monitoring data and prepares for transmission
func (ds *DataSender) ProcessData() error {
	if !ds.config.Enabled {
		log.Println("Data transmission is disabled")
		return nil
	}

	// Load today's data
	today := time.Now().Format("2006-01-02")
	dataFile := filepath.Join(ds.dataDir, fmt.Sprintf("combined_%s.json", today))

	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		log.Printf("Error reading data file: %v", err)
		return err
	}

	var combinedData CombinedData
	if err := json.Unmarshal(data, &combinedData); err != nil {
		log.Printf("Error unmarshaling data: %v", err)
		return err
	}

	// Process and create transmission payload
	payload := ds.createTransmissionPayload(combinedData)

	// Save to transmission folder
	if err := ds.saveTransmissionData(payload); err != nil {
		log.Printf("Error saving transmission data: %v", err)
		return err
	}

	// Send data to server
	if err := ds.sendData(payload); err != nil {
		log.Printf("Error sending data: %v", err)
		return err
	}

	log.Println("Data transmission completed successfully")
	return nil
}

// createTransmissionPayload creates a payload from the monitoring data
func (ds *DataSender) createTransmissionPayload(data CombinedData) TransmissionPayload {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	payload := TransmissionPayload{
		DeviceID:     ds.config.DeviceID,
		Timestamp:    timestamp,
		IntervalMins: 10,
		Apps:         make([]AppData, 0),
		Networks:     make([]NetworkData, 0),
	}

	// Process application data
	var focusedApp, activeApp string
	var maxFocusTime int64

	for appName, appInfo := range data.Apps {
		if appInfo.IsActive {
			activeApp = appName
		}
		if appInfo.IsFocused && appInfo.FocusTime > maxFocusTime {
			focusedApp = appName
			maxFocusTime = appInfo.FocusTime
		}
	}

	if activeApp != "" || focusedApp != "" {
		appData := AppData{
			ActiveApp:  activeApp,
			FocusedApp: focusedApp,
			FocusTime:  maxFocusTime,
			Timestamp:  timestamp,
		}
		payload.Apps = append(payload.Apps, appData)
	}

	// Process network data - count access frequency
	domainAccess := make(map[string]*NetworkData)

	for _, connInfo := range data.Network {
		if !connInfo.IsActive {
			continue
		}

		key := fmt.Sprintf("%s:%d", connInfo.Domain, connInfo.Port)
		if existing, exists := domainAccess[key]; exists {
			existing.AccessCount++
		} else {
			domainAccess[key] = &NetworkData{
				FQDN:        connInfo.Domain,
				Port:        connInfo.Port,
				AccessCount: 1,
				Protocol:    connInfo.Protocol,
				Timestamp:   timestamp,
			}
		}
	}

	// Convert map to slice
	for _, networkData := range domainAccess {
		payload.Networks = append(payload.Networks, *networkData)
	}

	// Add metadata
	payload.Metadata.OSVersion = "macOS" // Could be enhanced to get actual version
	payload.Metadata.AgentVersion = "1.0.0"
	payload.Metadata.TotalApps = len(data.Apps)
	payload.Metadata.TotalDomains = len(domainAccess)

	return payload
}

// saveTransmissionData saves the transmission data to local folder
func (ds *DataSender) saveTransmissionData(payload TransmissionPayload) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("transmission_%s.json", timestamp)
	filePath := filepath.Join(ds.transmissionDir, filename)

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, data, 0644)
}

// sendData sends the payload to the remote server
func (ds *DataSender) sendData(payload TransmissionPayload) error {
	if ds.config.BaseURL == "" || ds.config.APIKey == "" {
		return fmt.Errorf("BaseURL or APIKey not configured")
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %v", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/data", ds.config.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ds.config.APIKey))
	req.Header.Set("User-Agent", "ROI-Agent/1.0.0")

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Data sent successfully to %s", url)
	return nil
}

// CleanupOldTransmissionFiles removes old transmission files
func (ds *DataSender) CleanupOldTransmissionFiles() error {
	files, err := ioutil.ReadDir(ds.transmissionDir)
	if err != nil {
		return err
	}

	// Keep only files from the last 7 days
	cutoff := time.Now().AddDate(0, 0, -7)

	for _, file := range files {
		if file.ModTime().Before(cutoff) {
			filePath := filepath.Join(ds.transmissionDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("Error removing old file %s: %v", filePath, err)
			}
		}
	}

	return nil
}

// ShowConfig displays the current configuration
func (ds *DataSender) ShowConfig() {
	fmt.Printf("Data Transmission Configuration:\n")
	fmt.Printf("  Enabled: %v\n", ds.config.Enabled)
	fmt.Printf("  Base URL: %s\n", ds.config.BaseURL)
	fmt.Printf("  API Key: %s...\n", ds.config.APIKey[:min(len(ds.config.APIKey), 8)])
	fmt.Printf("  Device ID: %s\n", ds.config.DeviceID)
	fmt.Printf("  Config File: %s\n", ds.configPath)
	fmt.Printf("  Transmission Dir: %s\n", ds.transmissionDir)
}

// EnableTransmission enables data transmission
func (ds *DataSender) EnableTransmission(baseURL, apiKey string) {
	ds.config.Enabled = true
	if baseURL != "" {
		ds.config.BaseURL = baseURL
	}
	if apiKey != "" {
		ds.config.APIKey = apiKey
	}
	ds.saveConfig()
	fmt.Println("Data transmission enabled")
}

// DisableTransmission disables data transmission
func (ds *DataSender) DisableTransmission() {
	ds.config.Enabled = false
	ds.saveConfig()
	fmt.Println("Data transmission disabled")
}

// CreateEnvExample creates a .env.example file
func (ds *DataSender) CreateEnvExample() error {
	envExampleContent := `# ROI Agent Data Transmission Environment Variables
# Replace with your actual server URL and API key

ROI_AGENT_BASE_URL=https://api.yourserver.com/v1/roi-agent
ROI_AGENT_API_KEY=your-actual-api-key-here
`

	envExamplePath := filepath.Join(filepath.Dir(ds.configPath), "data-sender", ".env.example")
	return ioutil.WriteFile(envExamplePath, []byte(envExampleContent), 0644)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ROI Agent Data Sender")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  data-sender process          # Process and send current data")
		fmt.Println("  data-sender config           # Show current configuration")
		fmt.Println("  data-sender enable <url> <key>  # Enable transmission")
		fmt.Println("  data-sender disable          # Disable transmission")
		fmt.Println("  data-sender cleanup          # Cleanup old transmission files")
		fmt.Println("  data-sender env-example      # Create .env.example file")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  data-sender enable https://api.yourserver.com/v1/roi-agent your-api-key")
		fmt.Println("")
		fmt.Println("Environment Variables:")
		fmt.Println("  ROI_AGENT_BASE_URL      # Server base URL")
		fmt.Println("  ROI_AGENT_API_KEY       # API authentication key")
		fmt.Println("  ROI_AGENT_ENABLED       # Enable/disable transmission (true/false)")
		fmt.Println("  ROI_AGENT_DEVICE_ID     # Custom device identifier")
		return
	}

	sender := NewDataSender()
	command := os.Args[1]

	switch command {
	case "process":
		if err := sender.ProcessData(); err != nil {
			log.Fatalf("Error processing data: %v", err)
		}
	case "config":
		sender.ShowConfig()
	case "enable":
		if len(os.Args) >= 4 {
			sender.EnableTransmission(os.Args[2], os.Args[3])
		} else {
			fmt.Println("Usage: data-sender enable <base-url> <api-key>")
		}
	case "disable":
		sender.DisableTransmission()
	case "cleanup":
		if err := sender.CleanupOldTransmissionFiles(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		} else {
			fmt.Println("Cleanup completed")
		}
	case "env-example":
		if err := sender.CreateEnvExample(); err != nil {
			log.Printf("Error creating .env.example: %v", err)
		} else {
			fmt.Println(".env.example file created")
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
