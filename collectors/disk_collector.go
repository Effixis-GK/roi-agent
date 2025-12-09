// Package collectors provides system metrics collection functions for macOS
package collectors

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// DiskUsageMetrics represents disk space usage metrics
type DiskUsageMetrics struct {
	Device      string  `json:"device"`       // e.g., "/dev/disk1s1"
	MountPoint  string  `json:"mount_point"`  // e.g., "/"
	FileSystem  string  `json:"file_system"`  // e.g., "apfs"
	TotalGB     float64 `json:"total_gb"`     // Total space in GB
	UsedGB      float64 `json:"used_gb"`      // Used space in GB
	FreeGB      float64 `json:"free_gb"`      // Free space in GB
	UsedPercent float64 `json:"used_percent"` // Usage percentage
	InodesTotal uint64  `json:"inodes_total,omitempty"`
	InodesUsed  uint64  `json:"inodes_used,omitempty"`
	InodesFree  uint64  `json:"inodes_free,omitempty"`
}

// DiskSummary represents overall disk usage summary
type DiskSummary struct {
	TotalDisks       int                   `json:"total_disks"`
	TotalSpaceGB     float64               `json:"total_space_gb"`
	TotalUsedGB      float64               `json:"total_used_gb"`
	TotalFreeGB      float64               `json:"total_free_gb"`
	OverallUsedPct   float64               `json:"overall_used_pct"`
	CriticalDisks    int                   `json:"critical_disks"`  // >90% usage
	WarningDisks     int                   `json:"warning_disks"`   // >80% usage
	Disks            []*DiskUsageMetrics   `json:"disks"`
}

// CollectDiskUsageMetrics collects disk space usage for all mounted filesystems
func CollectDiskUsageMetrics() (*DiskSummary, error) {
	// Use df to get disk usage
	cmd := exec.Command("df", "-k")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %v", err)
	}

	summary := &DiskSummary{
		Disks: make([]*DiskUsageMetrics, 0),
	}

	lines := strings.Split(string(output), "\n")

	// Skip header line
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 6 {
			continue
		}

		device := parts[0]
		mountPoint := parts[len(parts)-1]

		// Skip virtual filesystems
		if strings.HasPrefix(device, "map") ||
			strings.HasPrefix(device, "devfs") ||
			strings.HasPrefix(mountPoint, "/System/Volumes/VM") ||
			strings.HasPrefix(mountPoint, "/System/Volumes/Preboot") ||
			strings.HasPrefix(mountPoint, "/System/Volumes/Update") {
			continue
		}

		// Parse sizes (in KB)
		totalKB, _ := strconv.ParseFloat(parts[1], 64)
		usedKB, _ := strconv.ParseFloat(parts[2], 64)
		freeKB, _ := strconv.ParseFloat(parts[3], 64)

		// Convert to GB
		totalGB := totalKB / (1024 * 1024)
		usedGB := usedKB / (1024 * 1024)
		freeGB := freeKB / (1024 * 1024)

		// Calculate percentage
		usedPct := 0.0
		if totalKB > 0 {
			usedPct = (usedKB / totalKB) * 100
		}

		disk := &DiskUsageMetrics{
			Device:      device,
			MountPoint:  mountPoint,
			TotalGB:     totalGB,
			UsedGB:      usedGB,
			FreeGB:      freeGB,
			UsedPercent: usedPct,
		}

		summary.Disks = append(summary.Disks, disk)
		summary.TotalDisks++
		summary.TotalSpaceGB += totalGB
		summary.TotalUsedGB += usedGB
		summary.TotalFreeGB += freeGB

		if usedPct >= 90 {
			summary.CriticalDisks++
		} else if usedPct >= 80 {
			summary.WarningDisks++
		}
	}

	if summary.TotalSpaceGB > 0 {
		summary.OverallUsedPct = (summary.TotalUsedGB / summary.TotalSpaceGB) * 100
	}

	return summary, nil
}

// GetRootDiskUsage returns disk usage for the root filesystem
func GetRootDiskUsage() (*DiskUsageMetrics, error) {
	summary, err := CollectDiskUsageMetrics()
	if err != nil {
		return nil, err
	}

	for _, disk := range summary.Disks {
		if disk.MountPoint == "/" {
			return disk, nil
		}
	}

	return nil, fmt.Errorf("root disk not found")
}

// GetDiskHealth returns overall disk health status
func GetDiskHealth(summary *DiskSummary) string {
	if summary.CriticalDisks > 0 {
		return "critical"
	} else if summary.WarningDisks > 0 {
		return "warning"
	}
	return "healthy"
}

