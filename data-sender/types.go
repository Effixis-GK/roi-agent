package main

import (
	"time"
)

// AppData represents application usage data for transmission
type AppData struct {
	ActiveApp  string `json:"active_app"`
	FocusedApp string `json:"focused_app"`
	FocusTime  int    `json:"focus_time_seconds"`  // Changed from int64 to int to match API
	Timestamp  string `json:"timestamp"`
}

// NetworkData represents network access data for transmission
type NetworkData struct {
	FQDN        string `json:"fqdn"`
	Port        int    `json:"port"`
	AccessCount int    `json:"access_count"`
	Protocol    string `json:"protocol"`
	Timestamp   string `json:"timestamp"`
}

// TransmissionPayload represents the complete data package to send
type TransmissionPayload struct {
	DeviceID     string        `json:"device_id"`
	Timestamp    string        `json:"timestamp"`
	IntervalMins int           `json:"interval_minutes"`
	StartTime    string        `json:"start_time"`
	EndTime      string        `json:"end_time"`
	Apps         []AppData     `json:"apps"`
	Networks     []NetworkData `json:"networks"`
	Metadata     struct {
		OSVersion    string `json:"os_version"`
		AgentVersion string `json:"agent_version"`
		TotalApps    int    `json:"total_apps"`
		TotalDomains int    `json:"total_domains"`
	} `json:"metadata"`
}

// TransmissionLog represents a transmission attempt log
type TransmissionLog struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Timestamp   time.Time `json:"timestamp"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	RetryCount  int       `json:"retry_count"`
	PayloadSize int       `json:"payload_size"`
}

// CombinedData represents the local data structure (matching main.go)
type CombinedData struct {
	Date    string                     `json:"date"`
	Apps    map[string]*AppUsage       `json:"apps"`
	Network map[string]*NetworkConn    `json:"network"`
}

type AppUsage struct {
	Name           string    `json:"name"`
	ForegroundTime int64     `json:"foreground_time"`
	FocusTime      int64     `json:"focus_time"`
	LastSeen       time.Time `json:"last_seen"`
	IsActive       bool      `json:"is_active"`
	IsFocused      bool      `json:"is_focused"`
}

type NetworkConn struct {
	Domain   string    `json:"domain"`
	Port     int       `json:"port"`
	Protocol string    `json:"protocol"`
	Duration int64     `json:"duration"`
	LastSeen time.Time `json:"last_seen"`
	IsActive bool      `json:"is_active"`
}
