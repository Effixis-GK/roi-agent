package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// Configuration for data transmission
type Config struct {
	BaseURL  string `json:"base_url" yaml:"base_url"`
	APIKey   string `json:"api_key" yaml:"api_key"`
	DeviceID string `json:"device_id" yaml:"device_id"`
	Enabled  bool   `json:"enabled" yaml:"enabled"`
}

// YAMLConfig represents the structure of config.yaml
type YAMLConfig struct {
	APIURL   string `yaml:"api_url"`
	APIKey   string `yaml:"api_key"`
	DeviceID string `yaml:"device_id"`
	OrgSlug  string `yaml:"org_slug"`
}

// loadConfig loads transmission configuration from file and environment variables
func (ds *DataSender) loadConfig() {
	// Try to load config.yaml first (for .app bundles)
	if loaded := ds.tryLoadYAMLConfig(); loaded {
		log.Printf("Loaded configuration from config.yaml")
		return
	}

	// Fall back to .env file
	ds.loadEnvConfig()
}

// tryLoadYAMLConfig attempts to load config from config.yaml
func (ds *DataSender) tryLoadYAMLConfig() bool {
	// Try multiple locations for config.yaml
	configPaths := []string{
		"config.yaml",                                    // Current directory
		"../Resources/config.yaml",                      // .app bundle Resources
		"../../Resources/config.yaml",                   // From MacOS folder
		filepath.Join(os.Getenv("HOME"), ".roiagent", "config.yaml"), // Home directory
	}

	for _, configPath := range configPaths {
		if data, err := ioutil.ReadFile(configPath); err == nil {
			var yamlConfig YAMLConfig
			if err := yaml.Unmarshal(data, &yamlConfig); err == nil {
				// Successfully parsed YAML config
				ds.config = Config{
					BaseURL:  yamlConfig.APIURL,
					APIKey:   yamlConfig.APIKey,
					DeviceID: yamlConfig.DeviceID,
					Enabled:  true,
				}
				log.Printf("Loaded config.yaml from: %s", configPath)
				return true
			}
		}
	}

	return false
}

// loadEnvConfig loads configuration from .env file
func (ds *DataSender) loadEnvConfig() {
	// Load .env file if it exists - try multiple locations
	envPaths := []string{
		"/Applications/ROI Agent/Resources/.env", // LaunchDaemon installation (PKG)
		".env",               // Current directory
		"./data-sender/.env", // From project root
		"../.env",            // Parent directory
	}

	for _, envPath := range envPaths {
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("Loaded .env file from: %s", envPath)
			break
		}
	}

	// Default configuration - Sample URLs/Keys
	// For Mac App, enable by default if environment variables are set
	enableByDefault := false
	if os.Getenv("ROI_AGENT_BASE_URL") != "" && os.Getenv("ROI_AGENT_API_KEY") != "" {
		enableByDefault = true
		log.Printf("Environment variables detected - enabling data transmission")
	}

	ds.config = Config{
		BaseURL:  "https://api.sample-server.com/v1/roi-agent-sample",
		APIKey:   "sample-api-key-replace-with-actual",
		DeviceID: ds.getOrGenerateDeviceID(), // ← 修正：永続的なdevice_IDを使用
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
		log.Printf("Using device ID from environment: %s", deviceID)
	}
	if enabled := os.Getenv("ROI_AGENT_ENABLED"); enabled != "" {
		if enabledBool, err := strconv.ParseBool(enabled); err == nil {
			ds.config.Enabled = enabledBool
		}
	}

	log.Printf("Data transmission enabled: %v", ds.config.Enabled)
	log.Printf("Base URL: %s", ds.config.BaseURL)
	log.Printf("Device ID: %s", ds.config.DeviceID)
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

// getOrGenerateDeviceID gets existing device ID or generates a new persistent one
func (ds *DataSender) getOrGenerateDeviceID() string {
	// Try to load existing device ID from saved config
	if existingID := ds.loadSavedDeviceID(); existingID != "" {
		log.Printf("Using saved device ID: %s", existingID)
		return existingID
	}

	// Generate a new persistent device ID
	newID := ds.generatePersistentDeviceID()
	log.Printf("Generated new persistent device ID: %s", newID)
	
	// Save it for future use
	ds.saveDeviceID(newID)
	
	return newID
}

// loadSavedDeviceID loads device ID from saved configuration
func (ds *DataSender) loadSavedDeviceID() string {
	data, err := ioutil.ReadFile(ds.configPath)
	if err != nil {
		return ""
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return ""
	}

	return config.DeviceID
}

// saveDeviceID saves device ID to configuration file
func (ds *DataSender) saveDeviceID(deviceID string) {
	tempConfig := Config{
		DeviceID: deviceID,
		BaseURL:  ds.config.BaseURL,
		APIKey:   ds.config.APIKey,
		Enabled:  ds.config.Enabled,
	}

	data, err := json.MarshalIndent(tempConfig, "", "  ")
	if err != nil {
		log.Printf("Error marshaling config for device ID save: %v", err)
		return
	}

	if err := ioutil.WriteFile(ds.configPath, data, 0644); err != nil {
		log.Printf("Error saving device ID: %v", err)
	} else {
		log.Printf("Device ID saved to: %s", ds.configPath)
	}
}

// generatePersistentDeviceID generates a unique, persistent device identifier
// This ID should remain constant across reboots and re-installations
func (ds *DataSender) generatePersistentDeviceID() string {
	hostname, _ := os.Hostname()
	
	// Use hardware-based unique identifier (macOS serial number)
	// If not available, fall back to hostname + hash
	serialNumber := ds.getHardwareSerialNumber()
	
	if serialNumber != "" {
		// Use serial number for guaranteed uniqueness
		return fmt.Sprintf("%s-%s", hostname, serialNumber)
	}
	
	// Fallback: Use hostname hash (deterministic)
	hash := sha256.Sum256([]byte(hostname))
	hashStr := hex.EncodeToString(hash[:])[:12] // First 12 characters
	
	return fmt.Sprintf("%s-%s", hostname, hashStr)
}

// getHardwareSerialNumber attempts to get hardware serial number (macOS)
func (ds *DataSender) getHardwareSerialNumber() string {
	// On macOS, try to get serial number using system_profiler
	// This is optional - if it fails, we use the hash fallback
	
	// Note: This requires execution permissions
	// For now, we return empty and use the hash fallback
	// Future enhancement: Execute `ioreg -l | grep IOPlatformSerialNumber`
	
	return ""
}

// generateDeviceID - DEPRECATED: kept for backward compatibility
// Use getOrGenerateDeviceID instead
func (ds *DataSender) generateDeviceID() string {
	log.Println("WARNING: generateDeviceID is deprecated, use getOrGenerateDeviceID instead")
	return ds.getOrGenerateDeviceID()
}
