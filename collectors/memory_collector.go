// Package collectors provides system metrics collection functions for macOS
package collectors

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ExtendedMemoryMetrics represents extended memory information including swap
type ExtendedMemoryMetrics struct {
	TotalBytes       uint64  `json:"total_bytes"`        // Total physical memory
	UsedBytes        uint64  `json:"used_bytes"`         // Used physical memory
	FreeBytes        uint64  `json:"free_bytes"`         // Free physical memory
	WiredBytes       uint64  `json:"wired_bytes"`        // Wired/locked memory
	ActiveBytes      uint64  `json:"active_bytes"`       // Active pages
	InactiveBytes    uint64  `json:"inactive_bytes"`     // Inactive pages
	CompressedBytes  uint64  `json:"compressed_bytes"`   // Compressed memory
	UsedPercent      float64 `json:"used_percent"`       // Memory usage percentage
	SwapTotalBytes   uint64  `json:"swap_total_bytes"`   // Total swap space
	SwapUsedBytes    uint64  `json:"swap_used_bytes"`    // Used swap space
	SwapFreeBytes    uint64  `json:"swap_free_bytes"`    // Free swap space
	SwapUsedPercent  float64 `json:"swap_used_percent"`  // Swap usage percentage
	PageInsPerSec    uint64  `json:"page_ins_per_sec"`   // Page ins (from vm_stat)
	PageOutsPerSec   uint64  `json:"page_outs_per_sec"`  // Page outs (from vm_stat)
	MemoryPressure   string  `json:"memory_pressure"`    // Memory pressure level
}

// CollectExtendedMemoryMetrics collects detailed memory metrics including swap
func CollectExtendedMemoryMetrics() (*ExtendedMemoryMetrics, error) {
	metrics := &ExtendedMemoryMetrics{}

	// Get total memory
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get total memory: %v", err)
	}
	metrics.TotalBytes, _ = strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)

	// Get page size
	cmd = exec.Command("sysctl", "-n", "hw.pagesize")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get page size: %v", err)
	}
	pageSize, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)

	// Get vm_stat for detailed memory info
	cmd = exec.Command("vm_stat")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get vm_stat: %v", err)
	}

	var freePages, activePages, inactivePages, speculativePages uint64
	var wiredPages, compressedPages, pageIns, pageOuts uint64

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Remove trailing period if present
		line = strings.TrimSuffix(line, ".")

		if strings.Contains(line, "Pages free:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				freePages, _ = strconv.ParseUint(parts[2], 10, 64)
			}
		} else if strings.Contains(line, "Pages active:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				activePages, _ = strconv.ParseUint(parts[2], 10, 64)
			}
		} else if strings.Contains(line, "Pages inactive:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				inactivePages, _ = strconv.ParseUint(parts[2], 10, 64)
			}
		} else if strings.Contains(line, "Pages speculative:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				speculativePages, _ = strconv.ParseUint(parts[2], 10, 64)
			}
		} else if strings.Contains(line, "Pages wired down:") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				wiredPages, _ = strconv.ParseUint(parts[3], 10, 64)
			}
		} else if strings.Contains(line, "Pages occupied by compressor:") {
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				compressedPages, _ = strconv.ParseUint(parts[4], 10, 64)
			}
		} else if strings.Contains(line, "Pageins:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pageIns, _ = strconv.ParseUint(parts[1], 10, 64)
			}
		} else if strings.Contains(line, "Pageouts:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pageOuts, _ = strconv.ParseUint(parts[1], 10, 64)
			}
		}
	}

	// Calculate memory values
	metrics.FreeBytes = (freePages + speculativePages) * pageSize
	metrics.ActiveBytes = activePages * pageSize
	metrics.InactiveBytes = inactivePages * pageSize
	metrics.WiredBytes = wiredPages * pageSize
	metrics.CompressedBytes = compressedPages * pageSize
	metrics.UsedBytes = metrics.TotalBytes - metrics.FreeBytes - metrics.InactiveBytes
	metrics.PageInsPerSec = pageIns
	metrics.PageOutsPerSec = pageOuts

	if metrics.TotalBytes > 0 {
		metrics.UsedPercent = float64(metrics.UsedBytes) / float64(metrics.TotalBytes) * 100
	}

	// Get swap info using sysctl
	if err := collectSwapInfo(metrics); err != nil {
		// Swap info collection failed, but continue with memory info
		metrics.SwapTotalBytes = 0
		metrics.SwapUsedBytes = 0
		metrics.SwapFreeBytes = 0
		metrics.SwapUsedPercent = 0
	}

	// Get memory pressure
	metrics.MemoryPressure = getMemoryPressure()

	return metrics, nil
}

// collectSwapInfo collects swap usage information
func collectSwapInfo(metrics *ExtendedMemoryMetrics) error {
	// Use sysctl vm.swapusage
	cmd := exec.Command("sysctl", "-n", "vm.swapusage")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Output format: "total = 2048.00M  used = 512.00M  free = 1536.00M"
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Fields(outputStr)

	for i := 0; i < len(parts); i++ {
		if parts[i] == "total" && i+2 < len(parts) && parts[i+1] == "=" {
			metrics.SwapTotalBytes = parseMemorySize(parts[i+2])
		} else if parts[i] == "used" && i+2 < len(parts) && parts[i+1] == "=" {
			metrics.SwapUsedBytes = parseMemorySize(parts[i+2])
		} else if parts[i] == "free" && i+2 < len(parts) && parts[i+1] == "=" {
			metrics.SwapFreeBytes = parseMemorySize(parts[i+2])
		}
	}

	if metrics.SwapTotalBytes > 0 {
		metrics.SwapUsedPercent = float64(metrics.SwapUsedBytes) / float64(metrics.SwapTotalBytes) * 100
	}

	return nil
}

// parseMemorySize parses a memory size string like "2048.00M" or "1.50G"
func parseMemorySize(sizeStr string) uint64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if len(sizeStr) == 0 {
		return 0
	}

	// Get the unit (last character)
	unit := sizeStr[len(sizeStr)-1]
	valueStr := sizeStr[:len(sizeStr)-1]

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0
	}

	switch unit {
	case 'K', 'k':
		return uint64(value * 1024)
	case 'M', 'm':
		return uint64(value * 1024 * 1024)
	case 'G', 'g':
		return uint64(value * 1024 * 1024 * 1024)
	case 'T', 't':
		return uint64(value * 1024 * 1024 * 1024 * 1024)
	default:
		return uint64(value)
	}
}

// getMemoryPressure returns the current memory pressure level
func getMemoryPressure() string {
	// Use memory_pressure command if available
	cmd := exec.Command("memory_pressure")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	outputStr := strings.ToLower(string(output))
	if strings.Contains(outputStr, "critical") {
		return "critical"
	} else if strings.Contains(outputStr, "warning") {
		return "warning"
	} else if strings.Contains(outputStr, "normal") {
		return "normal"
	}

	return "unknown"
}

// GetMemorySummary returns a simplified memory summary
type MemorySummary struct {
	TotalMB       int64   `json:"total_mb"`
	UsedMB        int64   `json:"used_mb"`
	FreeMB        int64   `json:"free_mb"`
	UsedPercent   float64 `json:"used_percent"`
	SwapTotalMB   int64   `json:"swap_total_mb"`
	SwapUsedMB    int64   `json:"swap_used_mb"`
	SwapFreeMB    int64   `json:"swap_free_mb"`
	SwapPercent   float64 `json:"swap_percent"`
	Pressure      string  `json:"pressure"`
}

// GetMemorySummary returns a simplified memory summary in MB
func GetMemorySummary() (*MemorySummary, error) {
	metrics, err := CollectExtendedMemoryMetrics()
	if err != nil {
		return nil, err
	}

	const mb = 1024 * 1024
	return &MemorySummary{
		TotalMB:       int64(metrics.TotalBytes / mb),
		UsedMB:        int64(metrics.UsedBytes / mb),
		FreeMB:        int64(metrics.FreeBytes / mb),
		UsedPercent:   metrics.UsedPercent,
		SwapTotalMB:   int64(metrics.SwapTotalBytes / mb),
		SwapUsedMB:    int64(metrics.SwapUsedBytes / mb),
		SwapFreeMB:    int64(metrics.SwapFreeBytes / mb),
		SwapPercent:   metrics.SwapUsedPercent,
		Pressure:      metrics.MemoryPressure,
	}, nil
}

