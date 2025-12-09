//go:build darwin
// +build darwin

// Package collectors provides system metrics collection functions for macOS
package collectors

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// WiFiMetrics represents WiFi connection metrics
type WiFiMetrics struct {
	SSID         string  `json:"ssid"`          // Network name
	BSSID        string  `json:"bssid"`         // Access point MAC address
	RSSI         int     `json:"rssi"`          // Signal strength in dBm (typically -30 to -90)
	Noise        int     `json:"noise"`         // Noise level in dBm
	Channel      int     `json:"channel"`       // WiFi channel
	TransmitRate float64 `json:"transmit_rate"` // Transmit rate in Mbps
	PHYMode      string  `json:"phy_mode"`      // 802.11a/b/g/n/ac/ax
	Security     string  `json:"security"`      // Security type (WPA2, etc.)
	MACAddress   string  `json:"mac_address"`   // Interface MAC address
	Connected    bool    `json:"connected"`     // Whether WiFi is connected
}

// CollectWiFiMetrics collects WiFi information using macOS airport utility
func CollectWiFiMetrics() (*WiFiMetrics, error) {
	// macOS airport utility path
	airportPath := "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"

	// Get WiFi info using airport -I
	cmd := exec.Command(airportPath, "-I")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get WiFi info: %v", err)
	}

	outputStr := string(output)
	metrics := &WiFiMetrics{
		Connected: true,
	}

	// Parse airport output
	// Format:
	//      agrCtlRSSI: -45
	//      agrExtRSSI: 0
	//      agrCtlNoise: -90
	//      ...
	//      SSID: MyNetwork
	//      BSSID: aa:bb:cc:dd:ee:ff
	//      channel: 36
	//      ...

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "agrCtlRSSI":
			metrics.RSSI, _ = strconv.Atoi(value)
		case "agrCtlNoise":
			metrics.Noise, _ = strconv.Atoi(value)
		case "SSID":
			metrics.SSID = value
		case "BSSID":
			metrics.BSSID = value
		case "channel":
			// Channel might be like "36,80" (channel,width)
			channelParts := strings.Split(value, ",")
			if len(channelParts) > 0 {
				metrics.Channel, _ = strconv.Atoi(channelParts[0])
			}
		case "lastTxRate":
			metrics.TransmitRate, _ = strconv.ParseFloat(value, 64)
		case "link auth":
			metrics.Security = value
		}
	}

	// Determine PHY mode based on channel and transmit rate
	metrics.PHYMode = determinePHYMode(metrics.Channel, metrics.TransmitRate)

	// Get MAC address
	metrics.MACAddress, _ = getWiFiMACAddress()

	// Check if actually connected
	if metrics.SSID == "" || metrics.RSSI == 0 {
		metrics.Connected = false
	}

	return metrics, nil
}

// determinePHYMode determines the PHY mode based on channel and transmit rate
func determinePHYMode(channel int, txRate float64) string {
	// This is a simplified heuristic
	// In reality, you'd need to query the specific PHY mode from the system

	// 5GHz channels
	is5GHz := channel >= 36

	if txRate > 1000 {
		return "802.11ax" // WiFi 6
	} else if txRate > 400 {
		return "802.11ac" // WiFi 5
	} else if txRate > 54 {
		return "802.11n" // WiFi 4
	} else if is5GHz {
		return "802.11a"
	} else {
		return "802.11g"
	}
}

// getWiFiMACAddress gets the MAC address of the WiFi interface
func getWiFiMACAddress() (string, error) {
	// Get the primary WiFi interface (usually en0 or en1)
	cmd := exec.Command("networksetup", "-listallhardwareports")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Find Wi-Fi interface and its MAC address
	lines := strings.Split(string(output), "\n")
	foundWiFi := false
	for _, line := range lines {
		if strings.Contains(line, "Wi-Fi") || strings.Contains(line, "AirPort") {
			foundWiFi = true
			continue
		}
		if foundWiFi && strings.HasPrefix(line, "Ethernet Address:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
		if foundWiFi && strings.HasPrefix(line, "Device:") {
			foundWiFi = false // Reset for next hardware port
		}
	}

	return "", fmt.Errorf("WiFi interface not found")
}

// GetWiFiSignalQuality returns a human-readable signal quality based on RSSI
func GetWiFiSignalQuality(rssi int) string {
	switch {
	case rssi >= -50:
		return "Excellent"
	case rssi >= -60:
		return "Good"
	case rssi >= -70:
		return "Fair"
	case rssi >= -80:
		return "Poor"
	default:
		return "Very Poor"
	}
}

// CollectWiFiNetworkList scans for available WiFi networks
func CollectWiFiNetworkList() ([]map[string]interface{}, error) {
	airportPath := "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"

	cmd := exec.Command(airportPath, "-s")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to scan WiFi networks: %v", err)
	}

	var networks []map[string]interface{}
	lines := strings.Split(string(output), "\n")

	// Skip header line
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Parse network info using regex for better handling
		// Format: SSID BSSID RSSI CHANNEL HT CC SECURITY
		re := regexp.MustCompile(`^\s*(.+?)\s+([0-9a-f:]+)\s+(-?\d+)\s+(\d+)`)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 5 {
			rssi, _ := strconv.Atoi(matches[3])
			channel, _ := strconv.Atoi(matches[4])
			networks = append(networks, map[string]interface{}{
				"ssid":    strings.TrimSpace(matches[1]),
				"bssid":   matches[2],
				"rssi":    rssi,
				"channel": channel,
			})
		}
	}

	return networks, nil
}

