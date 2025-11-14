package main

import (
	"time"
)

// AppData represents application usage data for transmission
type AppData struct {
	ActiveApp        string `json:"active_app"`
	FocusedApp       string `json:"focused_app"`
	FocusTimeSeconds int    `json:"focus_time_seconds"`      // フロントで見ていた時間
	ForegroundTimeSeconds int `json:"foreground_time_seconds"` // アプリの起動時間（追加）
	Timestamp        string `json:"timestamp"`
}

// NetworkData represents network access data for transmission
type NetworkData struct {
	FQDN        string `json:"fqdn"`
	Port        int    `json:"port"`
	AccessCount int    `json:"access_count"`
	Protocol    string `json:"protocol"`
	Timestamp   string `json:"timestamp"`
}

// SystemMetricsData represents system-wide performance metrics for transmission
type SystemMetricsData struct {
	Timestamp           string  `json:"timestamp"`
	CPUPercent          float64 `json:"cpu_percent"`
	MemoryUsedMB        int64   `json:"memory_used_mb"`
	MemoryTotalMB       int64   `json:"memory_total_mb"`
	MemoryPercent       float64 `json:"memory_percent"`
	DiskReadMBps        float64 `json:"disk_read_mbps"`
	DiskWriteMBps       float64 `json:"disk_write_mbps"`
	DiskReadOps         int64   `json:"disk_read_ops"`
	DiskWriteOps        int64   `json:"disk_write_ops"`
	BatteryLevel        int     `json:"battery_level,omitempty"`
	BatteryCharging     bool    `json:"battery_charging,omitempty"`
	BatteryTimeRemaining int    `json:"battery_time_remaining,omitempty"`
	IdleTimeSec         int64   `json:"idle_time_sec"`
	ScreenLocked        bool    `json:"screen_locked"`
	ProcessCount        int     `json:"process_count"`
	SystemUptimeSec     int64   `json:"system_uptime_sec"`
}

// ProcessMetricsData represents per-process CPU and memory metrics for transmission
type ProcessMetricsData struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	MemoryMB   int64   `json:"memory_mb"`
	Timestamp  string  `json:"timestamp"`
}

// TransmissionPayload represents the complete data package to send
type TransmissionPayload struct {
	DeviceID       string              `json:"device_id"`
	Timestamp      string              `json:"timestamp"`
	IntervalMins   int                 `json:"interval_minutes"`
	StartTime      string              `json:"start_time"`
	EndTime        string              `json:"end_time"`
	Apps           []AppData           `json:"apps"`
	Networks       []NetworkData       `json:"networks"`
	SystemMetrics  []SystemMetricsData  `json:"system_metrics"`
	ProcessMetrics []ProcessMetricsData `json:"process_metrics"`
	Metadata       struct {
		OSVersion     string `json:"os_version"`
		AgentVersion  string `json:"agent_version"`
		TotalApps     int    `json:"total_apps"`
		TotalDomains  int    `json:"total_domains"`
		Hostname      string `json:"hostname,omitempty"`
		EmployeeName  string `json:"employee_name,omitempty"`
		Department    string `json:"department,omitempty"`
		EmployeeEmail string `json:"employee_email,omitempty"`
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
	Date           string                     `json:"date"`
	Apps           map[string]*AppUsage       `json:"apps"`
	Network        map[string]*NetworkConn    `json:"network"`
	SystemMetrics  []*SystemMetricsLocal       `json:"system_metrics"`
	ProcessMetrics []*ProcessMetricsLocal     `json:"process_metrics"`
	AppTotal       struct {
		ForegroundTime int64 `json:"foreground_time"`
		BackgroundTime int64 `json:"background_time"`
		FocusTime      int64 `json:"focus_time"`
	} `json:"app_total"`
	NetworkTotal struct {
		TotalDuration     int64 `json:"total_duration"`
		UniqueConnections int   `json:"unique_connections"`
		UniqueDomains     int   `json:"unique_domains"`
	} `json:"network_total"`
}

// SystemMetricsLocal represents system metrics in local storage (with time.Time)
type SystemMetricsLocal struct {
	Timestamp          time.Time `json:"timestamp"`
	CPUPercent         float64   `json:"cpu_percent"`
	MemoryUsedMB       int64     `json:"memory_used_mb"`
	MemoryTotalMB      int64     `json:"memory_total_mb"`
	MemoryPercent      float64   `json:"memory_percent"`
	DiskReadMBps       float64   `json:"disk_read_mbps"`
	DiskWriteMBps      float64   `json:"disk_write_mbps"`
	DiskReadOps        int64     `json:"disk_read_ops"`
	DiskWriteOps       int64     `json:"disk_write_ops"`
	BatteryLevel       int       `json:"battery_level,omitempty"`
	BatteryCharging    bool      `json:"battery_charging,omitempty"`
	BatteryTimeRemaining int    `json:"battery_time_remaining,omitempty"`
	IdleTimeSec        int64     `json:"idle_time_sec"`
	ScreenLocked       bool      `json:"screen_locked"`
	ProcessCount       int       `json:"process_count"`
	SystemUptimeSec     int64     `json:"system_uptime_sec"`
}

// ProcessMetricsLocal represents process metrics in local storage (with time.Time)
type ProcessMetricsLocal struct {
	PID        int       `json:"pid"`
	Name       string    `json:"name"`
	CPUPercent float64   `json:"cpu_percent"`
	MemoryMB   int64     `json:"memory_mb"`
	Timestamp  time.Time `json:"timestamp"`
}

type AppUsage struct {
	Name           string    `json:"name"`
	ForegroundTime int64     `json:"foreground_time"`  // アプリの起動時間
	FocusTime      int64     `json:"focus_time"`       // フロントで見ていた時間
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
