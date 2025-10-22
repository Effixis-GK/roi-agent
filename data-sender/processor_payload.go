package main

import (
	"fmt"
	"log"
	"time"
)

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

	// Process application data - send ALL apps with BOTH foreground and focus times
	for appName, appInfo := range data.Apps {
		// Skip apps with no activity
		if !appInfo.IsActive && !appInfo.IsFocused {
			continue
		}

		appData := AppData{
			ActiveApp:             appName,
			FocusedApp:            appName,
			FocusTimeSeconds:      int(appInfo.FocusTime),      // フロントで見ていた時間
			ForegroundTimeSeconds: int(appInfo.ForegroundTime), // アプリの起動時間
			Timestamp:             timestamp,
		}
		
		payload.Apps = append(payload.Apps, appData)
		
		log.Printf("  Including app: %s (foreground: %d sec, focus: %d sec)", 
			appName, appInfo.ForegroundTime, appInfo.FocusTime)
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

	log.Printf("Created payload with %d apps and %d network connections", len(payload.Apps), len(payload.Networks))

	return payload
}
