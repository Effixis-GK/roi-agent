package main

import (
	"time"
)

// AppData represents application usage data for transmission
type AppData struct {
	ActiveApp             string `json:"active_app"`
	FocusedApp            string `json:"focused_app"`
	FocusTimeSeconds      int    `json:"focus_time_seconds"`      // フロントで見ていた時間
	ForegroundTimeSeconds int    `json:"foreground_time_seconds"` // アプリの起動時間（追加）
	IdleTimeSeconds       int    `json:"idle_time_seconds"`       // アイドル時間（追加）
	Timestamp             string `json:"timestamp"`
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
	Timestamp            string  `json:"timestamp"`
	CPUPercent           float64 `json:"cpu_percent"`
	MemoryUsedMB         int64   `json:"memory_used_mb"`
	MemoryTotalMB        int64   `json:"memory_total_mb"`
	MemoryPercent        float64 `json:"memory_percent"`
	DiskReadMBps         float64 `json:"disk_read_mbps"`
	DiskWriteMBps        float64 `json:"disk_write_mbps"`
	DiskReadOps          int64   `json:"disk_read_ops"`
	DiskWriteOps         int64   `json:"disk_write_ops"`
	BatteryLevel         int     `json:"battery_level,omitempty"`
	BatteryCharging      bool    `json:"battery_charging,omitempty"`
	BatteryTimeRemaining int     `json:"battery_time_remaining,omitempty"`
	IdleTimeSec          int64   `json:"idle_time_sec"`
	ScreenLocked         bool    `json:"screen_locked"`
	ProcessCount         int     `json:"process_count"`
	SystemUptimeSec      int64   `json:"system_uptime_sec"`
	// New metrics from datadog-agent
	Load1              float64 `json:"load_1,omitempty"`
	Load5              float64 `json:"load_5,omitempty"`
	Load15             float64 `json:"load_15,omitempty"`
	LoadNorm1          float64 `json:"load_norm_1,omitempty"`
	LoadNorm5          float64 `json:"load_norm_5,omitempty"`
	LoadNorm15         float64 `json:"load_norm_15,omitempty"`
	SwapUsedMB         int64   `json:"swap_used_mb,omitempty"`
	SwapTotalMB        int64   `json:"swap_total_mb,omitempty"`
	SwapPercent        float64 `json:"swap_percent,omitempty"`
	MemoryPressure     string  `json:"memory_pressure,omitempty"`
	NetBytesRecv       uint64  `json:"net_bytes_recv,omitempty"`
	NetBytesSent       uint64  `json:"net_bytes_sent,omitempty"`
	NetPacketsRecv     uint64  `json:"net_packets_recv,omitempty"`
	NetPacketsSent     uint64  `json:"net_packets_sent,omitempty"`
	NetErrorsIn        uint64  `json:"net_errors_in,omitempty"`
	NetErrorsOut       uint64  `json:"net_errors_out,omitempty"`
	WiFiSSID           string  `json:"wifi_ssid,omitempty"`
	WiFiRSSI           int     `json:"wifi_rssi,omitempty"`
	WiFiNoise          int     `json:"wifi_noise,omitempty"`
	WiFiChannel        int     `json:"wifi_channel,omitempty"`
	WiFiTransmitRate   float64 `json:"wifi_transmit_rate,omitempty"`
	WiFiPHYMode        string  `json:"wifi_phy_mode,omitempty"`
	WiFiSignalQuality  string  `json:"wifi_signal_quality,omitempty"`
	TCPEstablished     int     `json:"tcp_established,omitempty"`
	TCPTimeWait        int     `json:"tcp_time_wait,omitempty"`
	TCPCloseWait       int     `json:"tcp_close_wait,omitempty"`
	// Additional metrics - Phase 2
	DiskUsedPercent    float64 `json:"disk_used_percent,omitempty"`
	DiskFreeGB         float64 `json:"disk_free_gb,omitempty"`
	DiskTotalGB        float64 `json:"disk_total_gb,omitempty"`
	DiskHealth         string  `json:"disk_health,omitempty"`
	OpenFileHandles    int64   `json:"open_file_handles,omitempty"`
	MaxFileHandles     int64   `json:"max_file_handles,omitempty"`
	ExternalDisplays   int     `json:"external_displays,omitempty"`
	TotalDisplays      int     `json:"total_displays,omitempty"`
	BluetoothDevices   int     `json:"bluetooth_devices,omitempty"`
	MeetingActive      bool    `json:"meeting_active,omitempty"`
	CameraInUse        bool    `json:"camera_in_use,omitempty"`
	MicrophoneInUse    bool    `json:"microphone_in_use,omitempty"`
	BrowserTabs        int     `json:"browser_tabs,omitempty"`
	FocusScore         float64 `json:"focus_score,omitempty"`
	// New fields for DB compatibility
	MeetingStatus    string `json:"meeting_status,omitempty"`
	UserActive       bool   `json:"user_active,omitempty"`
	BrowserTabCount  int    `json:"browser_tab_count,omitempty"`
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
	Timestamp            time.Time `json:"timestamp"`
	CPUPercent           float64   `json:"cpu_percent"`
	MemoryUsedMB         int64     `json:"memory_used_mb"`
	MemoryTotalMB        int64     `json:"memory_total_mb"`
	MemoryPercent        float64   `json:"memory_percent"`
	DiskReadMBps         float64   `json:"disk_read_mbps"`
	DiskWriteMBps        float64   `json:"disk_write_mbps"`
	DiskReadOps          int64     `json:"disk_read_ops"`
	DiskWriteOps         int64     `json:"disk_write_ops"`
	BatteryLevel         int       `json:"battery_level,omitempty"`
	BatteryCharging      bool      `json:"battery_charging,omitempty"`
	BatteryTimeRemaining int       `json:"battery_time_remaining,omitempty"`
	IdleTimeSec          int64     `json:"idle_time_sec"`
	ScreenLocked         bool      `json:"screen_locked"`
	ProcessCount         int       `json:"process_count"`
	SystemUptimeSec      int64     `json:"system_uptime_sec"`
	// New metrics from datadog-agent
	Load1              float64 `json:"load_1,omitempty"`
	Load5              float64 `json:"load_5,omitempty"`
	Load15             float64 `json:"load_15,omitempty"`
	LoadNorm1          float64 `json:"load_norm_1,omitempty"`
	LoadNorm5          float64 `json:"load_norm_5,omitempty"`
	LoadNorm15         float64 `json:"load_norm_15,omitempty"`
	SwapUsedMB         int64   `json:"swap_used_mb,omitempty"`
	SwapTotalMB        int64   `json:"swap_total_mb,omitempty"`
	SwapPercent        float64 `json:"swap_percent,omitempty"`
	MemoryPressure     string  `json:"memory_pressure,omitempty"`
	NetBytesRecv       uint64  `json:"net_bytes_recv,omitempty"`
	NetBytesSent       uint64  `json:"net_bytes_sent,omitempty"`
	NetPacketsRecv     uint64  `json:"net_packets_recv,omitempty"`
	NetPacketsSent     uint64  `json:"net_packets_sent,omitempty"`
	NetErrorsIn        uint64  `json:"net_errors_in,omitempty"`
	NetErrorsOut       uint64  `json:"net_errors_out,omitempty"`
	WiFiSSID           string  `json:"wifi_ssid,omitempty"`
	WiFiRSSI           int     `json:"wifi_rssi,omitempty"`
	WiFiNoise          int     `json:"wifi_noise,omitempty"`
	WiFiChannel        int     `json:"wifi_channel,omitempty"`
	WiFiTransmitRate   float64 `json:"wifi_transmit_rate,omitempty"`
	WiFiPHYMode        string  `json:"wifi_phy_mode,omitempty"`
	WiFiSignalQuality  string  `json:"wifi_signal_quality,omitempty"`
	TCPEstablished     int     `json:"tcp_established,omitempty"`
	TCPTimeWait        int     `json:"tcp_time_wait,omitempty"`
	TCPCloseWait       int     `json:"tcp_close_wait,omitempty"`
	// Additional metrics - Phase 2
	DiskUsedPercent    float64 `json:"disk_used_percent,omitempty"`
	DiskFreeGB         float64 `json:"disk_free_gb,omitempty"`
	DiskTotalGB        float64 `json:"disk_total_gb,omitempty"`
	DiskHealth         string  `json:"disk_health,omitempty"`
	OpenFileHandles    int64   `json:"open_file_handles,omitempty"`
	MaxFileHandles     int64   `json:"max_file_handles,omitempty"`
	ExternalDisplays   int     `json:"external_displays,omitempty"`
	TotalDisplays      int     `json:"total_displays,omitempty"`
	BluetoothDevices   int     `json:"bluetooth_devices,omitempty"`
	MeetingActive      bool    `json:"meeting_active,omitempty"`
	CameraInUse        bool    `json:"camera_in_use,omitempty"`
	MicrophoneInUse    bool    `json:"microphone_in_use,omitempty"`
	BrowserTabs        int     `json:"browser_tabs,omitempty"`
	FocusScore         float64 `json:"focus_score,omitempty"`
	// New fields for DB compatibility
	MeetingStatus    string `json:"meeting_status,omitempty"`
	UserActive       bool   `json:"user_active,omitempty"`
	BrowserTabCount  int    `json:"browser_tab_count,omitempty"`
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
	IdleTime       int64     `json:"idle_time"`        // アイドル時間
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
