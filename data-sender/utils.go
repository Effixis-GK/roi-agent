package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// TestConnection tests the data transmission connection
func (ds *DataSender) TestConnection() {
	fmt.Println("Testing Data Transmission Configuration...")
	fmt.Println("=========================================")
	
	// Check configuration
	if !ds.config.Enabled {
		fmt.Println("❌ Data transmission is DISABLED")
		fmt.Println("   Please check .env file configuration")
		return
	}
	
	fmt.Println("✅ Data transmission is ENABLED")
	fmt.Printf("   Server URL: %s\n", ds.config.BaseURL)
	fmt.Printf("   Interval: %d minutes\n", ds.intervalMinutes)
	
	// Test connection with a minimal valid payload containing sample data
	testTimestamp := time.Now().UTC().Format(time.RFC3339)
	
	// Create minimal valid app data
	testAppData := AppData{
		ActiveApp:  "TestApp",
		FocusedApp: "TestApp", 
		FocusTime:  60, // 60 seconds for test
		Timestamp:  testTimestamp,
	}
	
	// Create minimal valid network data
	testNetworkData := NetworkData{
		FQDN:        "test.example.com",
		Port:        443,
		AccessCount: 1,
		Protocol:    "HTTPS",
		Timestamp:   testTimestamp,
	}
	
	testPayload := TransmissionPayload{
		DeviceID:     ds.config.DeviceID,
		Timestamp:    testTimestamp,
		IntervalMins: ds.intervalMinutes,
		StartTime:    time.Now().Add(-time.Duration(ds.intervalMinutes)*time.Minute).UTC().Format(time.RFC3339),
		EndTime:      time.Now().UTC().Format(time.RFC3339),
		Apps:         []AppData{testAppData},
		Networks:     []NetworkData{testNetworkData},
	}
	
	// Set metadata with correct counts
	testPayload.Metadata.OSVersion = "macOS"
	testPayload.Metadata.AgentVersion = "1.0.0-test"
	testPayload.Metadata.TotalApps = 1
	testPayload.Metadata.TotalDomains = 1
	
	fmt.Println("\n🔄 Testing connection to server...")
	fmt.Println("   Sending minimal test payload with sample app and network data")
	if err := ds.sendData(testPayload); err != nil {
		fmt.Printf("❌ Connection test FAILED: %v\n", err)
	} else {
		fmt.Println("✅ Connection test SUCCESSFUL")
		fmt.Println("   Data transmission is working properly")
		fmt.Println("   Sample data was accepted by the server")
	}
}

// ShowStatus shows current status and configuration (alias for ShowConfig)
func (ds *DataSender) ShowStatus() {
	ds.ShowConfig()
}

// CleanupOldFiles removes all old files (data, transmission, logs) older than 7 days
func (ds *DataSender) CleanupOldFiles() error {
	homeDir, _ := os.UserHomeDir()
	userDataDir := filepath.Join(homeDir, ".roiagent")

	// Directories to clean up
	dirsToClean := []string{
		filepath.Join(userDataDir, "data"),         // Agent data files
		filepath.Join(userDataDir, "transmission"), // Transmission files
		filepath.Join(userDataDir, "logs"),         // Log files
	}

	// Keep only files from the last 7 days
	cutoff := time.Now().AddDate(0, 0, -7)

	for _, dirPath := range dirsToClean {
		files, err := ioutil.ReadDir(dirPath)
		if err != nil {
			// Directory might not exist, skip
			log.Printf("Directory %s not found, skipping: %v", dirPath, err)
			continue
		}

		for _, file := range files {
			if file.ModTime().Before(cutoff) {
				filePath := filepath.Join(dirPath, file.Name())
				if err := os.Remove(filePath); err != nil {
					log.Printf("Error removing old file %s: %v", filePath, err)
				} else {
					log.Printf("Removed old file: %s", filePath)
				}
			}
		}
	}

	return nil
}

// CleanupOldTransmissionFiles removes old transmission files (deprecated - use CleanupOldFiles)
func (ds *DataSender) CleanupOldTransmissionFiles() error {
	return ds.CleanupOldFiles()
}

// ShowConfig displays the current configuration
func (ds *DataSender) ShowConfig() {
	fmt.Printf("Data Transmission Configuration:\n")
	fmt.Printf("  Enabled: %v\n", ds.config.Enabled)
	fmt.Printf("  Base URL: %s\n", ds.config.BaseURL)
	apiKeyDisplay := "***"
	if len(ds.config.APIKey) > 8 {
		apiKeyDisplay = ds.config.APIKey[:8] + "..."
	}
	fmt.Printf("  API Key: %s\n", apiKeyDisplay)
	fmt.Printf("  Device ID: %s\n", ds.config.DeviceID)
	fmt.Printf("  Interval: %d minutes\n", ds.intervalMinutes)
	fmt.Printf("  Config File: %s\n", ds.configPath)
	fmt.Printf("  Transmission Dir: %s\n", ds.transmissionDir)
	fmt.Printf("  Log File: %s\n", ds.logPath)
}

// ShowTransmissionLogs displays recent transmission logs
func (ds *DataSender) ShowTransmissionLogs(limit int) {
	logs, err := ds.loadTransmissionLogs()
	if err != nil {
		log.Printf("Error loading transmission logs: %v", err)
		return
	}

	if len(logs) == 0 {
		fmt.Println("No transmission logs found")
		return
	}

	// Show most recent logs first
	start := len(logs) - limit
	if start < 0 {
		start = 0
	}

	fmt.Printf("Recent Transmission Logs (showing %d of %d):\n", len(logs)-start, len(logs))
	fmt.Println("=======================================================")

	for i := len(logs) - 1; i >= start; i-- {
		logEntry := logs[i]
		status := "SUCCESS"
		if !logEntry.Success {
			status = "FAILED"
		}

		fmt.Printf("[%s] %s\n", logEntry.Timestamp.Format("2006-01-02 15:04:05"), status)
		fmt.Printf("  Interval: %s - %s\n", 
			logEntry.StartTime.Format("15:04:05"), logEntry.EndTime.Format("15:04:05"))
		fmt.Printf("  Payload Size: %d items, Retries: %d\n", logEntry.PayloadSize, logEntry.RetryCount)
		
		if logEntry.Error != "" {
			fmt.Printf("  Error: %s\n", logEntry.Error)
		}
		fmt.Println()
	}
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
ROI_AGENT_INTERVAL_MINUTES=10
`

	envExamplePath := filepath.Join(filepath.Dir(ds.configPath), "data-sender", ".env.example")
	return ioutil.WriteFile(envExamplePath, []byte(envExampleContent), 0644)
}

// SetTransmissionInterval sets the transmission interval and saves to environment
func (ds *DataSender) SetTransmissionInterval(minutes int) {
	if minutes < 1 || minutes > 1440 {
		fmt.Printf("Error: Invalid interval %d minutes. Must be between 1 and 1440 minutes.\n", minutes)
		return
	}

	ds.intervalMinutes = minutes

	// Update .env file
	envPath := filepath.Join(filepath.Dir(ds.configPath), "data-sender", ".env")
	envContent := fmt.Sprintf(`# ROI Agent Data Transmission Environment Variables
# Replace with your actual server URL and API key

ROI_AGENT_BASE_URL=%s
ROI_AGENT_API_KEY=%s
ROI_AGENT_INTERVAL_MINUTES=%d
`, ds.config.BaseURL, ds.config.APIKey, minutes)

	if err := ioutil.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		log.Printf("Error updating .env file: %v", err)
	} else {
		fmt.Printf("Transmission interval set to %d minutes and saved to .env file\n", minutes)
	}

	// Set environment variable for current session
	os.Setenv("ROI_AGENT_INTERVAL_MINUTES", fmt.Sprintf("%d", minutes))
	fmt.Println("Note: Restart the agent for the new interval to take effect.")
}
