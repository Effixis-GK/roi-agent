// Package collectors provides system metrics collection functions for macOS
package collectors

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// NetworkIOMetrics represents network I/O statistics for an interface
type NetworkIOMetrics struct {
	Interface   string `json:"interface"`    // Interface name (en0, en1, etc.)
	BytesRecv   uint64 `json:"bytes_recv"`   // Total bytes received
	BytesSent   uint64 `json:"bytes_sent"`   // Total bytes sent
	PacketsRecv uint64 `json:"packets_recv"` // Total packets received
	PacketsSent uint64 `json:"packets_sent"` // Total packets sent
	ErrorsIn    uint64 `json:"errors_in"`    // Input errors
	ErrorsOut   uint64 `json:"errors_out"`   // Output errors
	DropsIn     uint64 `json:"drops_in"`     // Input drops
	DropsOut    uint64 `json:"drops_out"`    // Output drops
}

// NetworkIOSummary represents aggregated network I/O for all interfaces
type NetworkIOSummary struct {
	TotalBytesRecv   uint64              `json:"total_bytes_recv"`
	TotalBytesSent   uint64              `json:"total_bytes_sent"`
	TotalPacketsRecv uint64              `json:"total_packets_recv"`
	TotalPacketsSent uint64              `json:"total_packets_sent"`
	Interfaces       []*NetworkIOMetrics `json:"interfaces"`
}

// CollectNetworkIOMetrics collects network I/O statistics for all interfaces
func CollectNetworkIOMetrics() (*NetworkIOSummary, error) {
	// Use netstat -ib to get interface statistics
	cmd := exec.Command("netstat", "-ib")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get network stats: %v", err)
	}

	summary := &NetworkIOSummary{
		Interfaces: make([]*NetworkIOMetrics, 0),
	}

	lines := strings.Split(string(output), "\n")

	// Skip header line
	// Format: Name  Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		// We need at least: Name Network Ipkts Ierrs Ibytes Opkts Oerrs Obytes
		if len(parts) < 10 {
			continue
		}

		interfaceName := parts[0]

		// Skip loopback and virtual interfaces
		if strings.HasPrefix(interfaceName, "lo") ||
			strings.HasPrefix(interfaceName, "gif") ||
			strings.HasPrefix(interfaceName, "stf") ||
			strings.HasPrefix(interfaceName, "utun") ||
			strings.HasPrefix(interfaceName, "bridge") ||
			strings.HasPrefix(interfaceName, "awdl") ||
			strings.HasPrefix(interfaceName, "llw") {
			continue
		}

		// Only count interfaces with actual network addresses (IPv4/IPv6)
		network := parts[2]
		if network == "Link" {
			// This is the link-level line, parse it for bytes
			metrics := &NetworkIOMetrics{
				Interface: interfaceName,
			}

			// Parse fields based on netstat -ib output
			// Ipkts Ierrs Ibytes Opkts Oerrs Obytes Coll
			if len(parts) >= 11 {
				metrics.PacketsRecv, _ = strconv.ParseUint(parts[4], 10, 64)
				metrics.ErrorsIn, _ = strconv.ParseUint(parts[5], 10, 64)
				metrics.BytesRecv, _ = strconv.ParseUint(parts[6], 10, 64)
				metrics.PacketsSent, _ = strconv.ParseUint(parts[7], 10, 64)
				metrics.ErrorsOut, _ = strconv.ParseUint(parts[8], 10, 64)
				metrics.BytesSent, _ = strconv.ParseUint(parts[9], 10, 64)
			}

			// Check if this interface already exists (dedupe)
			exists := false
			for _, existing := range summary.Interfaces {
				if existing.Interface == interfaceName {
					exists = true
					break
				}
			}

			if !exists && (metrics.BytesRecv > 0 || metrics.BytesSent > 0) {
				summary.Interfaces = append(summary.Interfaces, metrics)
				summary.TotalBytesRecv += metrics.BytesRecv
				summary.TotalBytesSent += metrics.BytesSent
				summary.TotalPacketsRecv += metrics.PacketsRecv
				summary.TotalPacketsSent += metrics.PacketsSent
			}
		}
	}

	return summary, nil
}

// CollectNetworkIOForInterface collects network I/O for a specific interface
func CollectNetworkIOForInterface(interfaceName string) (*NetworkIOMetrics, error) {
	summary, err := CollectNetworkIOMetrics()
	if err != nil {
		return nil, err
	}

	for _, iface := range summary.Interfaces {
		if iface.Interface == interfaceName {
			return iface, nil
		}
	}

	return nil, fmt.Errorf("interface %s not found", interfaceName)
}

// GetPrimaryNetworkInterface returns the name of the primary network interface
func GetPrimaryNetworkInterface() (string, error) {
	// Get the default route to determine primary interface
	cmd := exec.Command("route", "-n", "get", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get default route: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "interface:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("primary interface not found")
}

// TCPConnectionStats represents TCP connection statistics
type TCPConnectionStats struct {
	Established int `json:"established"`
	TimeWait    int `json:"time_wait"`
	CloseWait   int `json:"close_wait"`
	Listen      int `json:"listen"`
	SynSent     int `json:"syn_sent"`
	SynRecv     int `json:"syn_recv"`
	FinWait1    int `json:"fin_wait_1"`
	FinWait2    int `json:"fin_wait_2"`
	Closing     int `json:"closing"`
	LastAck     int `json:"last_ack"`
}

// CollectTCPConnectionStats collects TCP connection statistics
func CollectTCPConnectionStats() (*TCPConnectionStats, error) {
	cmd := exec.Command("netstat", "-an", "-p", "tcp")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get TCP stats: %v", err)
	}

	stats := &TCPConnectionStats{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Active") || strings.HasPrefix(line, "Proto") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 6 {
			continue
		}

		// Last field is the state
		state := strings.ToUpper(parts[len(parts)-1])

		switch state {
		case "ESTABLISHED":
			stats.Established++
		case "TIME_WAIT":
			stats.TimeWait++
		case "CLOSE_WAIT":
			stats.CloseWait++
		case "LISTEN":
			stats.Listen++
		case "SYN_SENT":
			stats.SynSent++
		case "SYN_RECEIVED", "SYN_RECV":
			stats.SynRecv++
		case "FIN_WAIT_1":
			stats.FinWait1++
		case "FIN_WAIT_2":
			stats.FinWait2++
		case "CLOSING":
			stats.Closing++
		case "LAST_ACK":
			stats.LastAck++
		}
	}

	return stats, nil
}

