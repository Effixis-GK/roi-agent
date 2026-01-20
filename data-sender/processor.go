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
	config          Config
	dataDir         string
	transmissionDir string
	configPath      string
	logPath         string
	intervalMinutes int
	defaultInterval int
	configPoller    *ConfigPoller
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
	
	// Initialize remote config poller
	sender.configPoller = NewConfigPoller(&sender.config)
	
	// Apply cached remote config interval if available (from dashboard settings)
	if sender.configPoller != nil {
		effectiveInterval := sender.configPoller.GetEffectiveIntervalMinutes()
		if effectiveInterval != sender.intervalMinutes {
			log.Printf("Using cached remote config interval: %d minutes", effectiveInterval)
			sender.intervalMinutes = effectiveInterval
		}
	}
	
	return sender
}

// CheckAndApplyRemoteConfig checks for remote config updates and applies them
func (ds *DataSender) CheckAndApplyRemoteConfig() {
	if ds.configPoller == nil {
		return
	}
	
	// Fetch remote config if it's time to poll
	if ds.configPoller.ShouldPoll() {
		log.Println("Checking for remote configuration updates...")
		remoteConfig, err := ds.configPoller.FetchConfig()
		if err != nil {
			log.Printf("Error fetching remote config: %v", err)
			return
		}
		
		// Apply remote config settings
		if remoteConfig != nil {
			// Update interval if changed
			if remoteConfig.IntervalMinutes > 0 && remoteConfig.IntervalMinutes != ds.intervalMinutes {
				log.Printf("Remote config: Updating interval from %d to %d minutes", 
					ds.intervalMinutes, remoteConfig.IntervalMinutes)
				ds.intervalMinutes = remoteConfig.IntervalMinutes
			}
			
			// Update enabled status
			ds.config.Enabled = remoteConfig.Enabled
			
			// Execute any pending commands
			ds.configPoller.ExecuteCommands()
			
			// Check for updates (auto-update when new version detected)
			if remoteConfig.LatestAgentVersion != "" && remoteConfig.UpdateMode != "disabled" {
				updater := NewAutoUpdater(&ds.config)
				if err := updater.CheckAndUpdate(remoteConfig); err != nil {
					log.Printf("Auto-update error: %v", err)
				}
			}
		}
	}
}

// FetchAndShowRemoteConfig fetches and displays the remote configuration
func (ds *DataSender) FetchAndShowRemoteConfig() {
	if ds.configPoller == nil {
		fmt.Println("Remote config poller not initialized")
		return
	}
	
	fmt.Println("Fetching remote configuration from server...")
	remoteConfig, err := ds.configPoller.FetchConfig()
	if err != nil {
		fmt.Printf("❌ Error fetching remote config: %v\n", err)
		return
	}
	
	fmt.Println("✅ Remote configuration fetched successfully!")
	fmt.Println("")
	ds.printRemoteConfig(remoteConfig)
	
	// Execute any pending commands
	if len(remoteConfig.Commands) > 0 {
		fmt.Printf("\n📋 Pending commands: %d\n", len(remoteConfig.Commands))
		for _, cmd := range remoteConfig.Commands {
			fmt.Printf("  - %s (ID: %s)\n", cmd.Command, cmd.ID)
		}
		fmt.Println("\nExecuting commands...")
		ds.configPoller.ExecuteCommands()
	}
}

// ShowRemoteConfig displays the current remote configuration
func (ds *DataSender) ShowRemoteConfig() {
	if ds.configPoller == nil {
		fmt.Println("Remote config poller not initialized")
		return
	}
	
	remoteConfig := ds.configPoller.GetCurrentConfig()
	if remoteConfig == nil {
		fmt.Println("No remote configuration cached locally.")
		fmt.Println("Run 'data-sender fetch-config' to fetch from server.")
		return
	}
	
	fmt.Println("📋 Current Remote Configuration (cached)")
	fmt.Println("")
	ds.printRemoteConfig(remoteConfig)
}

// printRemoteConfig prints the remote configuration in a readable format
func (ds *DataSender) printRemoteConfig(config *RemoteConfig) {
	fmt.Printf("  Config Version:    %d\n", config.ConfigVersion)
	fmt.Printf("  Interval Minutes:  %d\n", config.IntervalMinutes)
	fmt.Printf("  Enabled:           %v\n", config.Enabled)
	fmt.Println("")
	fmt.Println("  Collection Settings:")
	fmt.Printf("    - Apps:           %v\n", config.CollectApps)
	fmt.Printf("    - Network:        %v\n", config.CollectNetwork)
	fmt.Printf("    - System Metrics: %v\n", config.CollectSystemMetrics)
	fmt.Printf("    - Process Metrics:%v\n", config.CollectProcessMetrics)
	fmt.Println("")
	fmt.Printf("  Detail Level:      %s\n", config.DetailLevel)
	fmt.Printf("  Sample Rate:       %d seconds\n", config.SampleRateSeconds)
	
	if len(config.ExcludedApps) > 0 {
		fmt.Println("")
		fmt.Println("  Excluded Apps:")
		for _, app := range config.ExcludedApps {
			fmt.Printf("    - %s\n", app)
		}
	}
	
	if len(config.ExcludedDomains) > 0 {
		fmt.Println("")
		fmt.Println("  Excluded Domains:")
		for _, domain := range config.ExcludedDomains {
			fmt.Printf("    - %s\n", domain)
		}
	}
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
		data, dataFilePath, err := ds.loadDataForInterval(startTime, endTime)
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
			
			// 🔧 送信成功後、combined_*.jsonファイルの累積データをリセット
			if err := ds.resetCombinedDataAfterTransmission(dataFilePath); err != nil {
				log.Printf("Warning: Failed to reset combined data: %v", err)
			} else {
				log.Printf("Successfully reset combined data after transmission")
			}
			
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
func (ds *DataSender) loadDataForInterval(startTime, endTime time.Time) (*CombinedData, string, error) {
	// Load today's data file
	today := startTime.Format("2006-01-02")
	dataFile := filepath.Join(ds.dataDir, fmt.Sprintf("combined_%s.json", today))

	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		log.Printf("Error reading data file %s: %v", dataFile, err)
		return nil, "", err
	}

	var combinedData CombinedData
	if err := json.Unmarshal(data, &combinedData); err != nil {
		log.Printf("Error parsing data file: %v", err)
		return nil, "", err
	}

	// Filter the data for the specified interval
	filtered := ds.filterDataForInterval(&combinedData, startTime, endTime)
	return filtered, dataFile, nil
}

// filterDataForInterval filters data to only include activity within the specified interval
func (ds *DataSender) filterDataForInterval(data *CombinedData, startTime, endTime time.Time) *CombinedData {
	filtered := &CombinedData{
		Date:           data.Date,
		Apps:           make(map[string]*AppUsage),
		Network:        make(map[string]*NetworkConn),
		SystemMetrics:  make([]*SystemMetricsLocal, 0),
		ProcessMetrics: make([]*ProcessMetricsLocal, 0),
	}

	// 累積データをそのまま使用（10分間の累積データとして扱う）
	for appName, appInfo := range data.Apps {
		if appInfo.FocusTime > 0 || appInfo.ForegroundTime > 0 {
			appCopy := &AppUsage{
				Name:           appInfo.Name,
				ForegroundTime: appInfo.ForegroundTime,
				FocusTime:      appInfo.FocusTime,
				LastSeen:       appInfo.LastSeen,
				IsActive:       appInfo.IsActive,
				IsFocused:      appInfo.IsFocused,
			}
			filtered.Apps[appName] = appCopy
		}
	}

	// ネットワークデータも同様に処理
	for connKey, connInfo := range data.Network {
		if connInfo.IsActive {
			connCopy := &NetworkConn{
				Domain:   connInfo.Domain,
				Port:     connInfo.Port,
				Protocol: connInfo.Protocol,
				Duration: connInfo.Duration,
				LastSeen: connInfo.LastSeen,
				IsActive: connInfo.IsActive,
			}
			filtered.Network[connKey] = connCopy
		}
	}

	// Filter system metrics for the interval
	for _, metric := range data.SystemMetrics {
		if !metric.Timestamp.Before(startTime) && !metric.Timestamp.After(endTime) {
			filtered.SystemMetrics = append(filtered.SystemMetrics, metric)
		}
	}

	// Filter process metrics for the interval
	for _, metric := range data.ProcessMetrics {
		if !metric.Timestamp.Before(startTime) && !metric.Timestamp.After(endTime) {
			filtered.ProcessMetrics = append(filtered.ProcessMetrics, metric)
		}
	}

	log.Printf("Filtered data for interval %s-%s: %d apps, %d network connections, %d system metrics, %d process metrics",
		startTime.Format("15:04"), endTime.Format("15:04"),
		len(filtered.Apps), len(filtered.Network), len(filtered.SystemMetrics), len(filtered.ProcessMetrics))

	return filtered
}

// resetCombinedDataAfterTransmission resets the combined data file after successful transmission
func (ds *DataSender) resetCombinedDataAfterTransmission(dataFilePath string) error {
	// ファイルが存在しない場合はスキップ
	if _, err := os.Stat(dataFilePath); os.IsNotExist(err) {
		return nil
	}

	// ファイルを読み込む
	data, err := ioutil.ReadFile(dataFilePath)
	if err != nil {
		return fmt.Errorf("failed to read data file: %v", err)
	}

	var combinedData CombinedData
	if err := json.Unmarshal(data, &combinedData); err != nil {
		return fmt.Errorf("failed to parse data file: %v", err)
	}

		// 累積データをリセット（空のマップで初期化）
		combinedData.Apps = make(map[string]*AppUsage)
		combinedData.Network = make(map[string]*NetworkConn)
		combinedData.SystemMetrics = make([]*SystemMetricsLocal, 0)
		combinedData.ProcessMetrics = make([]*ProcessMetricsLocal, 0)
		
		// AppTotalとNetworkTotalもリセット
		combinedData.AppTotal.ForegroundTime = 0
		combinedData.AppTotal.BackgroundTime = 0
		combinedData.AppTotal.FocusTime = 0
		combinedData.NetworkTotal.TotalDuration = 0
		combinedData.NetworkTotal.UniqueConnections = 0
		combinedData.NetworkTotal.UniqueDomains = 0

	// リセット後のデータを保存
	resetData, err := json.MarshalIndent(combinedData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reset data: %v", err)
	}

	if err := ioutil.WriteFile(dataFilePath, resetData, 0644); err != nil {
		return fmt.Errorf("failed to write reset data: %v", err)
	}

	log.Printf("Reset combined data file: %s", dataFilePath)
	return nil
}

// createIntervalTransmissionPayload creates a payload for a specific interval
func (ds *DataSender) createIntervalTransmissionPayload(data *CombinedData, startTime, endTime time.Time) TransmissionPayload {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	payload := TransmissionPayload{
		DeviceID:       ds.config.DeviceID,
		Timestamp:      timestamp,
		IntervalMins:   ds.intervalMinutes,
		StartTime:      startTime.UTC().Format(time.RFC3339),
		EndTime:        endTime.UTC().Format(time.RFC3339),
		Apps:           make([]AppData, 0),
		Networks:       make([]NetworkData, 0),
		SystemMetrics:  make([]SystemMetricsData, 0),
		ProcessMetrics: make([]ProcessMetricsData, 0),
	}

	// Process application data - send ALL apps with their individual focus times
	for appName, appInfo := range data.Apps {
		// Skip apps with no activity
		if !appInfo.IsActive && !appInfo.IsFocused {
			continue
		}

		// Determine focused_app based on actual focus time
		focusedApp := ""
		if appInfo.FocusTime > 0 {
			focusedApp = appName
		}

		appData := AppData{
			ActiveApp:             appName,
			FocusedApp:            focusedApp,                  // Only set if actually focused
			FocusTimeSeconds:      int(appInfo.FocusTime),      // フォーカス時間（操作時間）
			ForegroundTimeSeconds: int(appInfo.ForegroundTime), // 起動時間（バックグラウンド含む）
			IdleTimeSeconds:       int(appInfo.IdleTime),       // アイドル時間
			Timestamp:             timestamp,
		}

		payload.Apps = append(payload.Apps, appData)

		if focusedApp != "" {
			log.Printf("  Including app: %s (focus: %ds, foreground: %ds)", appName, appInfo.FocusTime, appInfo.ForegroundTime)
		} else {
			log.Printf("  Including app: %s (foreground only: %ds, no focus)", appName, appInfo.ForegroundTime)
		}
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

	// Convert system metrics to transmission format
	for _, metric := range data.SystemMetrics {
		systemMetricData := SystemMetricsData{
			Timestamp:           metric.Timestamp.UTC().Format(time.RFC3339),
			CPUPercent:          metric.CPUPercent,
			MemoryUsedMB:        metric.MemoryUsedMB,
			MemoryTotalMB:       metric.MemoryTotalMB,
			MemoryPercent:       metric.MemoryPercent,
			DiskReadMBps:        metric.DiskReadMBps,
			DiskWriteMBps:       metric.DiskWriteMBps,
			DiskReadOps:         metric.DiskReadOps,
			DiskWriteOps:        metric.DiskWriteOps,
			IdleTimeSec:         metric.IdleTimeSec,
			ScreenLocked:        metric.ScreenLocked,
			ProcessCount:        metric.ProcessCount,
			SystemUptimeSec:     metric.SystemUptimeSec,
			// New metrics from datadog-agent
			Load1:              metric.Load1,
			Load5:              metric.Load5,
			Load15:             metric.Load15,
			LoadNorm1:          metric.LoadNorm1,
			LoadNorm5:          metric.LoadNorm5,
			LoadNorm15:         metric.LoadNorm15,
			SwapUsedMB:         metric.SwapUsedMB,
			SwapTotalMB:        metric.SwapTotalMB,
			SwapPercent:        metric.SwapPercent,
			MemoryPressure:     metric.MemoryPressure,
			NetBytesRecv:       metric.NetBytesRecv,
			NetBytesSent:       metric.NetBytesSent,
			NetPacketsRecv:     metric.NetPacketsRecv,
			NetPacketsSent:     metric.NetPacketsSent,
			NetErrorsIn:        metric.NetErrorsIn,
			NetErrorsOut:       metric.NetErrorsOut,
			WiFiSSID:           metric.WiFiSSID,
			WiFiRSSI:           metric.WiFiRSSI,
			WiFiNoise:          metric.WiFiNoise,
			WiFiChannel:        metric.WiFiChannel,
			WiFiTransmitRate:   metric.WiFiTransmitRate,
			WiFiPHYMode:        metric.WiFiPHYMode,
			WiFiSignalQuality:  metric.WiFiSignalQuality,
			TCPEstablished:     metric.TCPEstablished,
			TCPTimeWait:        metric.TCPTimeWait,
			TCPCloseWait:       metric.TCPCloseWait,
			// Additional metrics - Phase 2
			DiskUsedPercent:    metric.DiskUsedPercent,
			DiskFreeGB:         metric.DiskFreeGB,
			DiskTotalGB:        metric.DiskTotalGB,
			DiskHealth:         metric.DiskHealth,
			OpenFileHandles:    metric.OpenFileHandles,
			MaxFileHandles:     metric.MaxFileHandles,
			ExternalDisplays:   metric.ExternalDisplays,
			TotalDisplays:      metric.TotalDisplays,
			BluetoothDevices:   metric.BluetoothDevices,
			MeetingActive:      metric.MeetingActive,
			CameraInUse:        metric.CameraInUse,
			MicrophoneInUse:    metric.MicrophoneInUse,
			BrowserTabs:        metric.BrowserTabs,
			FocusScore:         metric.FocusScore,
			// New fields for DB compatibility
			MeetingStatus:   metric.MeetingStatus,
			UserActive:      metric.UserActive,
			BrowserTabCount: metric.BrowserTabCount,
		}
		// Only include battery fields if they are set (non-zero or true)
		if metric.BatteryLevel > 0 {
			systemMetricData.BatteryLevel = metric.BatteryLevel
			systemMetricData.BatteryCharging = metric.BatteryCharging
			if metric.BatteryTimeRemaining > 0 {
				systemMetricData.BatteryTimeRemaining = metric.BatteryTimeRemaining
			}
		}
		payload.SystemMetrics = append(payload.SystemMetrics, systemMetricData)
	}

	// Convert process metrics to transmission format
	for _, metric := range data.ProcessMetrics {
		processMetricData := ProcessMetricsData{
			PID:        metric.PID,
			Name:       metric.Name,
			CPUPercent: metric.CPUPercent,
			MemoryMB:   metric.MemoryMB,
			Timestamp:  metric.Timestamp.UTC().Format(time.RFC3339),
		}
		payload.ProcessMetrics = append(payload.ProcessMetrics, processMetricData)
	}

	// Add metadata
	payload.Metadata.OSVersion = "macOS"
	payload.Metadata.AgentVersion = GetAgentVersion()
	payload.Metadata.TotalApps = len(data.Apps)
	payload.Metadata.TotalDomains = len(domainAccess)

	// Add user information
	hostname, employeeName, employeeEmail := getUserInfo()
	payload.Metadata.Hostname = hostname
	payload.Metadata.EmployeeName = employeeName
	payload.Metadata.EmployeeEmail = employeeEmail

	log.Printf("Created payload with %d apps, %d network connections, %d system metrics, %d process metrics",
		len(payload.Apps), len(payload.Networks), len(payload.SystemMetrics), len(payload.ProcessMetrics))
	log.Printf("User Info: hostname=%s, name=%s, email=%s", hostname, employeeName, employeeEmail)

	return payload
}

// saveTransmissionData saves the transmission data to local folder
// Also saves to shared directory for easier access when running as LaunchDaemon
func (ds *DataSender) saveTransmissionData(payload TransmissionPayload, startTime time.Time) error {
	timestamp := startTime.Format("20060102_150405")
	filename := fmt.Sprintf("transmission_%s.json", timestamp)
	filePath := filepath.Join(ds.transmissionDir, filename)

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	// Save to primary location (user or root home directory)
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		log.Printf("Warning: Failed to save to primary location %s: %v", filePath, err)
	}

	// Also save to shared directory for easier access
	// This allows verification scripts to find the data regardless of which user ran the agent
	sharedTransmissionDir := "/var/lib/roiagent/transmission"
	if err := os.MkdirAll(sharedTransmissionDir, 0755); err == nil {
		sharedFilePath := filepath.Join(sharedTransmissionDir, filename)
		if err := ioutil.WriteFile(sharedFilePath, data, 0644); err != nil {
			log.Printf("Warning: Failed to save to shared location %s: %v", sharedFilePath, err)
		} else {
			log.Printf("Transmission data saved to shared location: %s", sharedFilePath)
		}
	}

	return nil
}

// SendInitialRegistration sends an initial registration payload with device info and version
// This is called immediately after installation to register the device without waiting for the first interval
func (ds *DataSender) SendInitialRegistration() error {
	if !ds.config.Enabled {
		log.Println("Data transmission is disabled")
		return fmt.Errorf("data transmission is disabled")
	}

	log.Println("Sending initial registration...")

	now := time.Now()
	timestamp := now.UTC().Format(time.RFC3339)

	// Create a minimal payload with device info only (no app/network data)
	// Note: IntervalMins must be > 0 for API validation
	payload := TransmissionPayload{
		DeviceID:       ds.config.DeviceID,
		Timestamp:      timestamp,
		IntervalMins:   1, // Minimum valid interval (API requires > 0)
		StartTime:      timestamp,
		EndTime:        timestamp,
		Apps:           make([]AppData, 0),
		Networks:       make([]NetworkData, 0),
		SystemMetrics:  make([]SystemMetricsData, 0),
		ProcessMetrics: make([]ProcessMetricsData, 0),
	}

	// Add metadata with device and version info
	payload.Metadata.OSVersion = getOSVersion()
	payload.Metadata.AgentVersion = GetAgentVersion()
	payload.Metadata.TotalApps = 0
	payload.Metadata.TotalDomains = 0

	// Add user information
	hostname, employeeName, employeeEmail := getUserInfo()
	payload.Metadata.Hostname = hostname
	payload.Metadata.EmployeeName = employeeName
	payload.Metadata.EmployeeEmail = employeeEmail

	log.Printf("Initial registration payload: DeviceID=%s, Version=%s, Hostname=%s",
		ds.config.DeviceID, payload.Metadata.AgentVersion, hostname)

	// Send data to server with retry
	retryCount := 0
	maxRetries := 3
	var lastErr error

	for retryCount <= maxRetries {
		err := ds.sendData(payload)
		if err == nil {
			log.Printf("Initial registration sent successfully")
			return nil
		}

		lastErr = err
		log.Printf("Initial registration failed (attempt %d/%d): %v", retryCount+1, maxRetries+1, err)
		retryCount++

		if retryCount <= maxRetries {
			waitTime := time.Duration(retryCount*2) * time.Second
			log.Printf("Retrying in %v...", waitTime)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("initial registration failed after %d attempts: %v", maxRetries+1, lastErr)
}

// CheckForUpdate checks if an update is available
func (ds *DataSender) CheckForUpdate() {
	fmt.Println("🔍 Checking for updates...")
	fmt.Println("")

	// Get current version
	currentVersion := GetAgentVersion()
	fmt.Printf("Current version: %s\n", currentVersion)

	// Fetch remote config to get latest version info
	if ds.configPoller == nil {
		fmt.Println("❌ Remote config poller not initialized")
		return
	}

	remoteConfig, err := ds.configPoller.FetchConfig()
	if err != nil {
		fmt.Printf("❌ Error fetching remote config: %v\n", err)
		return
	}

	if remoteConfig == nil || remoteConfig.LatestAgentVersion == "" {
		fmt.Println("ℹ️  No update information available from server")
		return
	}

	fmt.Printf("Latest version: %s\n", remoteConfig.LatestAgentVersion)
	fmt.Printf("Update mode: %s\n", remoteConfig.UpdateMode)
	fmt.Println("")

	// Check if update is available
	if currentVersion == remoteConfig.LatestAgentVersion {
		fmt.Println("✅ You are running the latest version!")
		return
	}

	if currentVersion == "dev" {
		fmt.Println("ℹ️  Running development build - updates skipped")
		return
	}

	fmt.Println("📦 Update available!")
	if remoteConfig.UpdateRequired {
		fmt.Println("⚠️  This is a REQUIRED update")
	}
	if remoteConfig.UpdateURL != "" {
		fmt.Printf("Download URL: %s\n", remoteConfig.UpdateURL)
	}
	fmt.Println("")
	fmt.Println("Run 'data-sender update' to install the update")
}

// CheckAndInstallUpdate checks for and installs updates
func (ds *DataSender) CheckAndInstallUpdate() {
	fmt.Println("🔄 Checking and installing updates...")
	fmt.Println("")

	// Fetch remote config
	if ds.configPoller == nil {
		fmt.Println("❌ Remote config poller not initialized")
		return
	}

	remoteConfig, err := ds.configPoller.FetchConfig()
	if err != nil {
		fmt.Printf("❌ Error fetching remote config: %v\n", err)
		return
	}

	if remoteConfig == nil {
		fmt.Println("❌ No remote configuration available")
		return
	}

	// Create auto updater
	updater := NewAutoUpdater(&ds.config)
	currentVersion := updater.GetCurrentVersion()

	fmt.Printf("Current version: %s\n", currentVersion)

	if remoteConfig.LatestAgentVersion != "" {
		fmt.Printf("Latest version: %s\n", remoteConfig.LatestAgentVersion)
	}

	// Check and install update
	err = updater.CheckAndUpdate(remoteConfig)
	if err != nil {
		fmt.Printf("❌ Update failed: %v\n", err)
		return
	}

	fmt.Println("✅ Update check completed")
}
