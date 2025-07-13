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
	"time"

	"github.com/joho/godotenv"
)

// Test structures matching the API
type TestAppData struct {
	ActiveApp  string `json:"active_app"`
	FocusedApp string `json:"focused_app"`
	FocusTime  int64  `json:"focus_time_seconds"`
	Timestamp  string `json:"timestamp"`
}

type TestNetworkData struct {
	FQDN        string `json:"fqdn"`
	Port        int    `json:"port"`
	AccessCount int    `json:"access_count"`
	Protocol    string `json:"protocol"`
	Timestamp   string `json:"timestamp"`
}

type TestPayload struct {
	DeviceID     string            `json:"device_id"`
	Timestamp    string            `json:"timestamp"`
	IntervalMins int               `json:"interval_minutes"`
	Apps         []TestAppData     `json:"apps"`
	Networks     []TestNetworkData `json:"networks"`
	Metadata     struct {
		OSVersion    string `json:"os_version"`
		AgentVersion string `json:"agent_version"`
		TotalApps    int    `json:"total_apps"`
		TotalDomains int    `json:"total_domains"`
	} `json:"metadata"`
}

func loadEnvVars() (string, string, error) {
	// Load .env file from data-sender directory
	homeDir, _ := os.UserHomeDir()
	projectRoot := filepath.Join(homeDir, "Local", "GitHub", "Test_ROI", "roi-agent")
	envPath := filepath.Join(projectRoot, "data-sender", ".env")
	
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("Warning: Could not load .env file from %s: %v\n", envPath, err)
		fmt.Println("Using environment variables instead...")
	}

	baseURL := os.Getenv("ROI_AGENT_BASE_URL")
	apiKey := os.Getenv("ROI_AGENT_API_KEY")

	if baseURL == "" || apiKey == "" {
		return "", "", fmt.Errorf("missing required environment variables: ROI_AGENT_BASE_URL or ROI_AGENT_API_KEY")
	}

	return baseURL, apiKey, nil
}

func createTestPayload() TestPayload {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	
	payload := TestPayload{
		DeviceID:     "MacBook-Pro-1752306890",
		Timestamp:    timestamp,
		IntervalMins: 10,
		Apps: []TestAppData{
			{
				ActiveApp:  "Google Chrome",
				FocusedApp: "Google Chrome",
				FocusTime:  180,
				Timestamp:  timestamp,
			},
		},
		Networks: []TestNetworkData{
			{
				FQDN:        "www.yahoo.co.jp",
				Port:        443,
				AccessCount: 3,
				Protocol:    "HTTPS",
				Timestamp:   timestamp,
			},
			{
				FQDN:        "chatgpt.com",
				Port:        443,
				AccessCount: 1,
				Protocol:    "HTTPS",
				Timestamp:   timestamp,
			},
		},
	}

	// Set metadata
	payload.Metadata.OSVersion = "macOS"
	payload.Metadata.AgentVersion = "1.0.0"
	payload.Metadata.TotalApps = 15
	payload.Metadata.TotalDomains = 8

	return payload
}

func sendTestData(baseURL, apiKey string, payload TestPayload, useCorrectHeaders bool) error {
	// Convert payload to JSON
	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling payload: %v", err)
	}

	fmt.Println("=== Request Payload ===")
	fmt.Println(string(jsonData))
	fmt.Println()

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	if useCorrectHeaders {
		// Use X-API-Key header (matching the curl example)
		req.Header.Set("X-API-Key", apiKey)
		fmt.Println("=== Using X-API-Key Header (Correct) ===")
	} else {
		// Use Authorization Bearer header (current implementation)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		fmt.Println("=== Using Authorization Bearer Header (Current) ===")
	}
	
	req.Header.Set("User-Agent", "ROI-Agent-Debug/1.0.0")

	fmt.Printf("POST %s\n", baseURL)
	fmt.Println("Headers:")
	for name, values := range req.Header {
		for _, value := range values {
			if name == "X-API-Key" || name == "Authorization" {
				// Hide sensitive data partially
				hiddenValue := value[:8] + "..."
				fmt.Printf("  %s: %s\n", name, hiddenValue)
			} else {
				fmt.Printf("  %s: %s\n", name, value)
			}
		}
	}
	fmt.Println()

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	fmt.Printf("=== Response Status: %d %s ===\n", resp.StatusCode, resp.Status)
	fmt.Println("Response Headers:")
	for name, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", name, value)
		}
	}
	fmt.Println()
	fmt.Println("Response Body:")
	
	// Try to pretty print JSON response
	var prettyJSON interface{}
	if err := json.Unmarshal(body, &prettyJSON); err == nil {
		prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")
		fmt.Println(string(prettyBytes))
	} else {
		fmt.Println(string(body))
	}
	fmt.Println()

	// Check if successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("✅ Request successful!")
		return nil
	} else {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}
}

func main() {
	fmt.Println("🔧 ROI Agent Data Transmission Debug Tool")
	fmt.Println("=========================================")
	fmt.Println()

	// Load environment variables
	baseURL, apiKey, err := loadEnvVars()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	fmt.Printf("📡 Server URL: %s\n", baseURL)
	fmt.Printf("🔑 API Key: %s...\n", apiKey[:8])
	fmt.Println()

	// Create test payload
	payload := createTestPayload()

	// Test with current headers (Authorization Bearer)
	fmt.Println("🧪 Test 1: Current Implementation (Authorization Bearer)")
	fmt.Println("======================================================")
	if err := sendTestData(baseURL, apiKey, payload, false); err != nil {
		fmt.Printf("❌ Current implementation failed: %v\n", err)
	}
	fmt.Println()

	// Test with correct headers (X-API-Key)
	fmt.Println("🧪 Test 2: Corrected Implementation (X-API-Key)")
	fmt.Println("==============================================")
	if err := sendTestData(baseURL, apiKey, payload, true); err != nil {
		fmt.Printf("❌ Corrected implementation failed: %v\n", err)
	}
	fmt.Println()

	fmt.Println("🏁 Debug test completed!")
	fmt.Println()
	fmt.Println("💡 Notes:")
	fmt.Println("- The curl example uses 'X-API-Key' header")
	fmt.Println("- Current data-sender uses 'Authorization: Bearer' header")
	fmt.Println("- If Test 2 succeeds and Test 1 fails, we need to update data-sender")
}
