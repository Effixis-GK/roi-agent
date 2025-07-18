package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

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

	// Create HTTP request - use BaseURL directly (not BaseURL/data)
	url := ds.config.BaseURL
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", ds.config.APIKey)
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

	// Read response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		responseBody = []byte("(unable to read response)")
	}

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	log.Printf("Data sent successfully to %s", url)
	log.Printf("Server response (Status %d): %s", resp.StatusCode, string(responseBody))
	return nil
}
