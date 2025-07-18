package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// DataSender handles the transmission of monitoring data
type DataSender struct {
	config            Config
	dataDir           string
	transmissionDir   string
	configPath        string
	logPath           string
	intervalMinutes   int
	defaultInterval   int
}

// NewDataSender creates a new data sender instance
func NewDataSender() *DataSender {
	homeDir, _ := os.UserHomeDir()
	userDataDir := filepath.Join(homeDir, ".roiagent")
	dataDir := filepath.Join(userDataDir, "data")
	transmissionDir := filepath.Join(userDataDir, "transmission")
	configPath := filepath.Join(userDataDir, "transmission_config.json")
	logPath := filepath.Join(userDataDir, "transmission_logs.json")

	// Create directories
	os.MkdirAll(transmissionDir, 0755)

	sender := &DataSender{
		dataDir:         dataDir,
		transmissionDir: transmissionDir,
		configPath:      configPath,
		logPath:         logPath,
		intervalMinutes: 10,
		defaultInterval: 10,
	}

	// Load interval from environment variable if set
	if intervalStr := os.Getenv("ROI_AGENT_INTERVAL_MINUTES"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			sender.intervalMinutes = interval
			log.Printf("Using custom transmission interval: %d minutes", interval)
		}
	}

	sender.loadConfig()
	return sender
}

// processCurrentInterval processes and sends data for the current interval
func (ds *DataSender) processCurrentInterval() error {
	if !ds.config.Enabled {
		log.Println("Data transmission is disabled")
		return nil
	}

	now := time.Now()
	// Calculate the previous interval from now
	endTime := now
	startTime := endTime.Add(-time.Duration(ds.intervalMinutes) * time.Minute)

	log.Printf("Processing interval: %s to %s", startTime.Format("15:04:05"), endTime.Format("15:04:05"))

	return ds.ProcessDataInterval(startTime, endTime)
}

// ProcessDataInterval processes and sends data for a specific time interval
func (ds *DataSender) ProcessDataInterval(startTime, endTime time.Time) error {
	if !ds.config.Enabled {
		log.Println("Data transmission is disabled")
		return nil
	}

	// Check if this interval was already transmitted
	if ds.wasIntervalTransmitted(startTime, endTime) {
		log.Printf("Interval %s-%s already transmitted, skipping", 
			startTime.Format("15:04"), endTime.Format("15:04"))
		return nil
	}

	retryCount := 0
	maxRetries := 3

	for retryCount <= maxRetries {
		// Load data for the specific interval
		data, err := ds.loadDataForInterval(startTime, endTime)
		if err != nil {
			log.Printf("Error loading data for interval: %v", err)
			ds.logTransmissionResult(startTime, endTime, false, err, retryCount, 0)
			return err
		}

		// Create transmission payload
		payload := ds.createIntervalTransmissionPayload(data, startTime, endTime)

		// Save transmission data locally
		if err := ds.saveTransmissionData(payload, startTime); err != nil {
			log.Printf("Error saving transmission data: %v", err)
		}

		// Send data to server
		err = ds.sendData(payload)
		payloadSize := len(payload.Apps) + len(payload.Networks)

		if err == nil {
			log.Printf("Successfully transmitted interval %s-%s (attempt %d)", 
				startTime.Format("15:04"), endTime.Format("15:04"), retryCount+1)
			ds.logTransmissionResult(startTime, endTime, true, nil, retryCount, payloadSize)
			return nil
		}

		log.Printf("Transmission failed (attempt %d/%d): %v", retryCount+1, maxRetries+1, err)
		retryCount++

		if retryCount <= maxRetries {
			waitTime := time.Duration(retryCount*2) * time.Second
			log.Printf("Retrying in %v...", waitTime)
			time.Sleep(waitTime)
		}
	}

	// Log final failure
	ds.logTransmissionResult(startTime, endTime, false, fmt.Errorf("max retries exceeded"), retryCount-1, 0)
	return fmt.Errorf("failed to transmit after %d attempts", maxRetries+1)
}

// loadDataForInterval loads and filters data for a specific time interval
func (ds *DataSender) loadDataForInterval(startTime, endTime time.Time) (*CombinedData, error) {
	// Load today's data file
	today := startTime.Format("2006-01-02")
	dataFile := filepath.Join(ds.dataDir, fmt.Sprintf("combined_%s.json", today))

	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		log.Printf("Error reading data file %s: %v", dataFile, err)
		return nil, err
	}

	var combinedData CombinedData
	if err := json.Unmarshal(data, &combinedData); err != nil {
		log.Printf("Error unmarshaling data: %v", err)
		return nil, err
	}

	// Filter data for the specific interval
	filteredData := ds.filterDataForInterval(&combinedData, startTime, endTime)
	return filteredData, nil
}

// filterDataForInterval filters data to only include activity within the specified interval
func (ds *DataSender) filterDataForInterval(data *CombinedData, startTime, endTime time.Time) *CombinedData {
	filtered := &CombinedData{
		Date:    data.Date,
		Apps:    make(map[string]*AppUsage),
		Network: make(map[string]*NetworkConn),
	}

	// Filter apps based on LastSeen timestamp
	for appName, appInfo := range data.Apps {
		if appInfo.LastSeen.After(startTime) && appInfo.LastSeen.Before(endTime) {
			filtered.Apps[appName] = appInfo
		}
	}

	// Filter network connections based on LastSeen timestamp
	for connKey, connInfo := range data.Network {
		if connInfo.LastSeen.After(startTime) && connInfo.LastSeen.Before(endTime) {
			filtered.Network[connKey] = connInfo
		}
	}

	log.Printf("Filtered data for interval %s-%s: %d apps, %d network connections",
		startTime.Format("15:04"), endTime.Format("15:04"), 
		len(filtered.Apps), len(filtered.Network))

	return filtered
}

// createIntervalTransmissionPayload creates a payload for a specific interval
func (ds *DataSender) createIntervalTransmissionPayload(data *CombinedData, startTime, endTime time.Time) TransmissionPayload {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	payload := TransmissionPayload{
		DeviceID:     ds.config.DeviceID,
		Timestamp:    timestamp,
		IntervalMins: ds.intervalMinutes,
		StartTime:    startTime.UTC().Format(time.RFC3339),
		EndTime:      endTime.UTC().Format(time.RFC3339),
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
			FocusTime:  int(maxFocusTime), // Convert int64 to int
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
	payload.Metadata.OSVersion = "macOS"
	payload.Metadata.AgentVersion = "1.0.0"
	payload.Metadata.TotalApps = len(data.Apps)
	payload.Metadata.TotalDomains = len(domainAccess)

	return payload
}

// saveTransmissionData saves the transmission data to local folder
func (ds *DataSender) saveTransmissionData(payload TransmissionPayload, startTime time.Time) error {
	timestamp := startTime.Format("20060102_150405")
	filename := fmt.Sprintf("transmission_%s.json", timestamp)
	filePath := filepath.Join(ds.transmissionDir, filename)

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, data, 0644)
}
