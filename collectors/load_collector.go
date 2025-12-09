// Package collectors provides system metrics collection functions for macOS
package collectors

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// LoadMetrics represents system load average metrics
type LoadMetrics struct {
	Load1      float64 `json:"load_1"`       // 1-minute load average
	Load5      float64 `json:"load_5"`       // 5-minute load average
	Load15     float64 `json:"load_15"`      // 15-minute load average
	LoadNorm1  float64 `json:"load_norm_1"`  // Normalized by CPU count
	LoadNorm5  float64 `json:"load_norm_5"`  // Normalized by CPU count
	LoadNorm15 float64 `json:"load_norm_15"` // Normalized by CPU count
}

// CollectLoadMetrics collects system load average metrics using sysctl
func CollectLoadMetrics() (*LoadMetrics, error) {
	// Get load averages using sysctl vm.loadavg
	cmd := exec.Command("sysctl", "-n", "vm.loadavg")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get load average: %v", err)
	}

	// Output format: "{ 1.23 4.56 7.89 }"
	outputStr := strings.TrimSpace(string(output))
	outputStr = strings.Trim(outputStr, "{ }")
	parts := strings.Fields(outputStr)

	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected load average format: %s", outputStr)
	}

	load1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse load1: %v", err)
	}

	load5, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse load5: %v", err)
	}

	load15, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse load15: %v", err)
	}

	// Get CPU count for normalization
	numCPU, err := GetLogicalCPUCount()
	if err != nil {
		numCPU = 1 // fallback
	}
	cpuFloat := float64(numCPU)

	return &LoadMetrics{
		Load1:      load1,
		Load5:      load5,
		Load15:     load15,
		LoadNorm1:  load1 / cpuFloat,
		LoadNorm5:  load5 / cpuFloat,
		LoadNorm15: load15 / cpuFloat,
	}, nil
}

// GetLogicalCPUCount returns the number of logical CPUs
func GetLogicalCPUCount() (int, error) {
	cmd := exec.Command("sysctl", "-n", "hw.logicalcpu")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetPhysicalCPUCount returns the number of physical CPUs
func GetPhysicalCPUCount() (int, error) {
	cmd := exec.Command("sysctl", "-n", "hw.physicalcpu")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}

	return count, nil
}

