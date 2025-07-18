package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Configuration for data transmission
type Config struct {
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	DeviceID string `json:"device_id"`
	Enabled  bool   `json:"enabled"`
}

// loadConfig loads transmission configuration from file and environment variables
func (ds *DataSender) loadConfig() {
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
