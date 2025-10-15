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

	log.Printf("Data transmission enabled: %v", ds.config.Enabled)
	log.Printf("Base URL: %s", ds.config.BaseURL)
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
