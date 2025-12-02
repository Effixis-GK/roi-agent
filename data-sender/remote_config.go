package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RemoteConfig represents the configuration received from the server
type RemoteConfig struct {
	ConfigVersion         int      `json:"config_version"`
	IntervalMinutes       int      `json:"interval_minutes"`
	Enabled               bool     `json:"enabled"`
	CollectApps           bool     `json:"collect_apps"`
	CollectNetwork        bool     `json:"collect_network"`
	CollectSystemMetrics  bool     `json:"collect_system_metrics"`
	CollectProcessMetrics bool     `json:"collect_process_metrics"`
	ExcludedApps          []string `json:"excluded_apps"`
	ExcludedDomains       []string `json:"excluded_domains"`
	DetailLevel           string   `json:"detail_level"`
	SampleRateSeconds     int      `json:"sample_rate_seconds"`
	Commands              []RemoteCommand `json:"commands,omitempty"`
	
	// Auto-update information
	LatestAgentVersion    string   `json:"latest_agent_version,omitempty"`
	UpdateURL             string   `json:"update_url,omitempty"`
	UpdateChecksum        string   `json:"update_checksum,omitempty"`
	UpdateRequired        bool     `json:"update_required,omitempty"`
	UpdateMode            string   `json:"update_mode,omitempty"` // "auto", "notify", "disabled"
}

// RemoteCommand represents a command from the server
type RemoteCommand struct {
	ID         string          `json:"id"`
	Command    string          `json:"command"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
}

// LocalRemoteConfig stores the last known remote config locally
type LocalRemoteConfig struct {
	RemoteConfig
	LastFetched time.Time `json:"last_fetched"`
}

// ConfigPoller handles polling for remote configuration
type ConfigPoller struct {
	config          *Config
	localConfigPath string
	remoteConfig    *RemoteConfig
	lastFetch       time.Time
	pollInterval    time.Duration
	onConfigChange  func(*RemoteConfig)
}

// NewConfigPoller creates a new configuration poller
func NewConfigPoller(config *Config) *ConfigPoller {
	homeDir, _ := os.UserHomeDir()
	localConfigPath := filepath.Join(homeDir, ".roiagent", "remote_config.json")

	poller := &ConfigPoller{
		config:          config,
		localConfigPath: localConfigPath,
		pollInterval:    10 * time.Minute, // Default poll interval
	}

	// Try to load cached remote config
	poller.loadLocalConfig()

	return poller
}

// SetOnConfigChange sets the callback for when config changes
func (cp *ConfigPoller) SetOnConfigChange(callback func(*RemoteConfig)) {
	cp.onConfigChange = callback
}

// loadLocalConfig loads the cached remote configuration from disk
func (cp *ConfigPoller) loadLocalConfig() {
	data, err := ioutil.ReadFile(cp.localConfigPath)
	if err != nil {
		log.Printf("No cached remote config found: %v", err)
		return
	}

	var localConfig LocalRemoteConfig
	if err := json.Unmarshal(data, &localConfig); err != nil {
		log.Printf("Error parsing cached remote config: %v", err)
		return
	}

	cp.remoteConfig = &localConfig.RemoteConfig
	cp.lastFetch = localConfig.LastFetched
	log.Printf("Loaded cached remote config (version %d, fetched %s ago)",
		localConfig.ConfigVersion, time.Since(cp.lastFetch).Round(time.Minute))
}

// saveLocalConfig saves the remote configuration to disk
func (cp *ConfigPoller) saveLocalConfig() {
	if cp.remoteConfig == nil {
		return
	}

	localConfig := LocalRemoteConfig{
		RemoteConfig: *cp.remoteConfig,
		LastFetched:  cp.lastFetch,
	}

	data, err := json.MarshalIndent(localConfig, "", "  ")
	if err != nil {
		log.Printf("Error marshaling remote config: %v", err)
		return
	}

	// Ensure directory exists
	dir := filepath.Dir(cp.localConfigPath)
	os.MkdirAll(dir, 0755)

	if err := ioutil.WriteFile(cp.localConfigPath, data, 0644); err != nil {
		log.Printf("Error saving remote config: %v", err)
	}
}

// FetchConfig fetches the latest configuration from the server
func (cp *ConfigPoller) FetchConfig() (*RemoteConfig, error) {
	if cp.config.BaseURL == "" || cp.config.APIKey == "" {
		return nil, fmt.Errorf("BaseURL or APIKey not configured")
	}

	// Build config URL (replace /device with /config if needed)
	configURL := cp.config.BaseURL
	if strings.HasSuffix(configURL, "/device") {
		configURL = strings.TrimSuffix(configURL, "/device") + "/config"
	} else if !strings.HasSuffix(configURL, "/config") {
		configURL = strings.TrimSuffix(configURL, "/") + "/config"
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", configURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating config request: %v", err)
	}

	// Set headers
	req.Header.Set("X-API-Key", cp.config.APIKey)
	req.Header.Set("X-Device-ID", cp.config.DeviceID)
	req.Header.Set("User-Agent", fmt.Sprintf("ROI-Agent/%s", GetAgentVersion()))

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching config: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading config response: %v", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("config request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var remoteConfig RemoteConfig
	if err := json.Unmarshal(body, &remoteConfig); err != nil {
		return nil, fmt.Errorf("error parsing config response: %v", err)
	}

	// Check if config changed
	configChanged := cp.remoteConfig == nil || cp.remoteConfig.ConfigVersion != remoteConfig.ConfigVersion

	// Update local state
	cp.remoteConfig = &remoteConfig
	cp.lastFetch = time.Now()
	cp.saveLocalConfig()

	if configChanged {
		log.Printf("Remote config updated to version %d", remoteConfig.ConfigVersion)
		if cp.onConfigChange != nil {
			cp.onConfigChange(&remoteConfig)
		}
	}

	return &remoteConfig, nil
}

// AckCommand acknowledges a command execution to the server
func (cp *ConfigPoller) AckCommand(commandID string, success bool, result interface{}) error {
	if cp.config.BaseURL == "" || cp.config.APIKey == "" {
		return fmt.Errorf("BaseURL or APIKey not configured")
	}

	// Build ack URL
	ackURL := cp.config.BaseURL
	if strings.HasSuffix(ackURL, "/device") {
		ackURL = strings.TrimSuffix(ackURL, "/device") + "/config/ack"
	} else {
		ackURL = strings.TrimSuffix(ackURL, "/") + "/config/ack"
	}

	// Build request body
	resultJSON, _ := json.Marshal(result)
	ackReq := map[string]interface{}{
		"command_id": commandID,
		"success":    success,
		"result":     json.RawMessage(resultJSON),
	}

	reqBody, err := json.Marshal(ackReq)
	if err != nil {
		return fmt.Errorf("error marshaling ack request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", ackURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Errorf("error creating ack request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", cp.config.APIKey)
	req.Header.Set("X-Device-ID", cp.config.DeviceID)
	req.Header.Set("User-Agent", fmt.Sprintf("ROI-Agent/%s", GetAgentVersion()))

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending ack: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("ack request failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Command %s acknowledged successfully", commandID)
	return nil
}

// ExecuteCommands executes any pending commands from the remote config
func (cp *ConfigPoller) ExecuteCommands() {
	if cp.remoteConfig == nil || len(cp.remoteConfig.Commands) == 0 {
		return
	}

	for _, cmd := range cp.remoteConfig.Commands {
		log.Printf("Executing command: %s (ID: %s)", cmd.Command, cmd.ID)
		
		var success bool
		var result interface{}

		switch cmd.Command {
		case "pause":
			// Pause data collection
			log.Println("Pausing data collection...")
			success = true
			result = map[string]string{"status": "paused"}

		case "resume":
			// Resume data collection
			log.Println("Resuming data collection...")
			success = true
			result = map[string]string{"status": "resumed"}

		case "sync_now":
			// Trigger immediate data sync
			log.Println("Triggering immediate data sync...")
			success = true
			result = map[string]string{"status": "sync_triggered"}

		case "clear_cache":
			// Clear local cache
			log.Println("Clearing local cache...")
			success = true
			result = map[string]string{"status": "cache_cleared"}

		case "update_agent":
			// Notify about available update
			log.Println("Agent update notification received")
			success = true
			result = map[string]string{"status": "update_acknowledged", "current_version": GetAgentVersion()}

		default:
			log.Printf("Unknown command: %s", cmd.Command)
			success = false
			result = map[string]string{"error": "unknown_command"}
		}

		// Acknowledge command execution
		if err := cp.AckCommand(cmd.ID, success, result); err != nil {
			log.Printf("Error acknowledging command %s: %v", cmd.ID, err)
		}
	}

	// Clear commands after execution
	cp.remoteConfig.Commands = nil
}

// GetCurrentConfig returns the current remote configuration
func (cp *ConfigPoller) GetCurrentConfig() *RemoteConfig {
	return cp.remoteConfig
}

// ShouldPoll returns true if it's time to poll for new config
func (cp *ConfigPoller) ShouldPoll() bool {
	return time.Since(cp.lastFetch) >= cp.pollInterval
}

// GetEffectiveIntervalMinutes returns the effective data transmission interval
func (cp *ConfigPoller) GetEffectiveIntervalMinutes() int {
	if cp.remoteConfig != nil && cp.remoteConfig.IntervalMinutes > 0 {
		return cp.remoteConfig.IntervalMinutes
	}
	return 10 // Default
}

// IsEnabled returns whether data collection is enabled
func (cp *ConfigPoller) IsEnabled() bool {
	if cp.remoteConfig != nil {
		return cp.remoteConfig.Enabled
	}
	return true // Default enabled
}

// IsAppExcluded checks if an app should be excluded from monitoring
func (cp *ConfigPoller) IsAppExcluded(appName string) bool {
	if cp.remoteConfig == nil {
		return false
	}
	for _, excluded := range cp.remoteConfig.ExcludedApps {
		if strings.EqualFold(excluded, appName) {
			return true
		}
		// Support wildcard matching
		if strings.HasPrefix(excluded, "*") && strings.HasSuffix(strings.ToLower(appName), strings.ToLower(excluded[1:])) {
			return true
		}
		if strings.HasSuffix(excluded, "*") && strings.HasPrefix(strings.ToLower(appName), strings.ToLower(excluded[:len(excluded)-1])) {
			return true
		}
	}
	return false
}

// IsDomainExcluded checks if a domain should be excluded from monitoring
func (cp *ConfigPoller) IsDomainExcluded(domain string) bool {
	if cp.remoteConfig == nil {
		return false
	}
	for _, excluded := range cp.remoteConfig.ExcludedDomains {
		if strings.EqualFold(excluded, domain) {
			return true
		}
		// Support wildcard matching (e.g., *.apple.com)
		if strings.HasPrefix(excluded, "*.") {
			suffix := excluded[1:] // Keep the dot
			if strings.HasSuffix(strings.ToLower(domain), strings.ToLower(suffix)) {
				return true
			}
		}
	}
	return false
}

