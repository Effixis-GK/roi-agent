// Package collectors provides system metrics collection functions for macOS
package collectors

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FileHandleMetrics represents file descriptor usage
type FileHandleMetrics struct {
	OpenFiles     int64 `json:"open_files"`      // Currently open files
	MaxFiles      int64 `json:"max_files"`       // Maximum allowed files
	UsedPercent   float64 `json:"used_percent"`  // Percentage used
}

// CollectFileHandleMetrics collects file descriptor usage
func CollectFileHandleMetrics() (*FileHandleMetrics, error) {
	metrics := &FileHandleMetrics{}

	// Get current open files
	cmd := exec.Command("sysctl", "-n", "kern.num_files")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get open files: %v", err)
	}
	metrics.OpenFiles, _ = strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)

	// Get max files
	cmd = exec.Command("sysctl", "-n", "kern.maxfiles")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get max files: %v", err)
	}
	metrics.MaxFiles, _ = strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)

	if metrics.MaxFiles > 0 {
		metrics.UsedPercent = float64(metrics.OpenFiles) / float64(metrics.MaxFiles) * 100
	}

	return metrics, nil
}

// DisplayInfo represents external display information
type DisplayInfo struct {
	DisplayID   string `json:"display_id"`
	Name        string `json:"name"`
	Resolution  string `json:"resolution"`   // e.g., "2560x1440"
	RefreshRate int    `json:"refresh_rate"` // Hz
	IsBuiltIn   bool   `json:"is_built_in"`
	IsMain      bool   `json:"is_main"`
	Rotation    int    `json:"rotation"`     // degrees
}

// DisplayMetrics represents display configuration metrics
type DisplayMetrics struct {
	TotalDisplays    int            `json:"total_displays"`
	ExternalDisplays int            `json:"external_displays"`
	Displays         []*DisplayInfo `json:"displays"`
}

// CollectDisplayMetrics collects information about connected displays
func CollectDisplayMetrics() (*DisplayMetrics, error) {
	metrics := &DisplayMetrics{
		Displays: make([]*DisplayInfo, 0),
	}

	// Use system_profiler to get display info
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get display info: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var currentDisplay *DisplayInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Detect new display section
		if strings.HasSuffix(line, ":") && !strings.Contains(line, "Resolution") && 
			!strings.Contains(line, "Main Display") && !strings.Contains(line, "Mirror") &&
			!strings.Contains(line, "Online") && !strings.Contains(line, "Rotation") {
			if currentDisplay != nil {
				metrics.Displays = append(metrics.Displays, currentDisplay)
				metrics.TotalDisplays++
				if !currentDisplay.IsBuiltIn {
					metrics.ExternalDisplays++
				}
			}
			currentDisplay = &DisplayInfo{
				Name: strings.TrimSuffix(line, ":"),
			}
		}

		if currentDisplay == nil {
			continue
		}

		// Parse display properties
		if strings.HasPrefix(line, "Resolution:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				res := strings.TrimSpace(parts[1])
				// Format: "2560 x 1440 (QHD/WQHD - Wide Quad High Definition) @ 60 Hz"
				if idx := strings.Index(res, "@"); idx > 0 {
					resOnly := strings.TrimSpace(res[:idx])
					currentDisplay.Resolution = strings.ReplaceAll(resOnly, " x ", "x")
					
					// Extract refresh rate
					hzPart := strings.TrimSpace(res[idx+1:])
					hzPart = strings.TrimSuffix(hzPart, " Hz")
					hzPart = strings.TrimSpace(hzPart)
					currentDisplay.RefreshRate, _ = strconv.Atoi(hzPart)
				} else {
					currentDisplay.Resolution = strings.ReplaceAll(res, " x ", "x")
				}
			}
		} else if strings.Contains(line, "Main Display: Yes") {
			currentDisplay.IsMain = true
		} else if strings.Contains(line, "Built-In") || strings.Contains(line, "Internal") {
			currentDisplay.IsBuiltIn = true
		} else if strings.HasPrefix(line, "Rotation:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				rotStr := strings.TrimSpace(parts[1])
				rotStr = strings.TrimSuffix(rotStr, "°")
				currentDisplay.Rotation, _ = strconv.Atoi(rotStr)
			}
		}
	}

	// Don't forget the last display
	if currentDisplay != nil {
		metrics.Displays = append(metrics.Displays, currentDisplay)
		metrics.TotalDisplays++
		if !currentDisplay.IsBuiltIn {
			metrics.ExternalDisplays++
		}
	}

	return metrics, nil
}

// BluetoothDeviceInfo represents a connected Bluetooth device
type BluetoothDeviceInfo struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	Type       string `json:"type"`        // e.g., "Keyboard", "Mouse", "Headphones"
	Connected  bool   `json:"connected"`
	BatteryLevel int  `json:"battery_level,omitempty"` // 0-100, if available
}

// BluetoothMetrics represents Bluetooth connectivity metrics
type BluetoothMetrics struct {
	Enabled          bool                   `json:"enabled"`
	Discoverable     bool                   `json:"discoverable"`
	ConnectedDevices int                    `json:"connected_devices"`
	Devices          []*BluetoothDeviceInfo `json:"devices"`
}

// CollectBluetoothMetrics collects Bluetooth device information
func CollectBluetoothMetrics() (*BluetoothMetrics, error) {
	metrics := &BluetoothMetrics{
		Devices: make([]*BluetoothDeviceInfo, 0),
	}

	// Use system_profiler to get Bluetooth info
	cmd := exec.Command("system_profiler", "SPBluetoothDataType")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Bluetooth info: %v", err)
	}

	outputStr := string(output)
	
	// Check if Bluetooth is enabled
	if strings.Contains(outputStr, "State: On") {
		metrics.Enabled = true
	}

	if strings.Contains(outputStr, "Discoverable: Yes") {
		metrics.Discoverable = true
	}

	// Parse connected devices
	lines := strings.Split(outputStr, "\n")
	inConnectedSection := false
	var currentDevice *BluetoothDeviceInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "Connected:") {
			inConnectedSection = true
			continue
		}

		if !inConnectedSection {
			continue
		}

		// Device name (ends with ":")
		if strings.HasSuffix(line, ":") && !strings.Contains(line, "Address") &&
			!strings.Contains(line, "Type") && !strings.Contains(line, "Battery") {
			if currentDevice != nil && currentDevice.Name != "" {
				metrics.Devices = append(metrics.Devices, currentDevice)
				metrics.ConnectedDevices++
			}
			currentDevice = &BluetoothDeviceInfo{
				Name:      strings.TrimSuffix(line, ":"),
				Connected: true,
			}
		}

		if currentDevice == nil {
			continue
		}

		// Parse device properties
		if strings.HasPrefix(line, "Address:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentDevice.Address = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Type:") || strings.HasPrefix(line, "Minor Type:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentDevice.Type = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Battery Level") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				battStr := strings.TrimSpace(parts[1])
				battStr = strings.TrimSuffix(battStr, "%")
				currentDevice.BatteryLevel, _ = strconv.Atoi(battStr)
			}
		}
	}

	// Don't forget the last device
	if currentDevice != nil && currentDevice.Name != "" {
		metrics.Devices = append(metrics.Devices, currentDevice)
		metrics.ConnectedDevices++
	}

	return metrics, nil
}

// ContextSwitchMetrics represents CPU context switch statistics
type ContextSwitchMetrics struct {
	ContextSwitches uint64 `json:"context_switches"`
	Interrupts      uint64 `json:"interrupts"`
}

// CollectContextSwitchMetrics collects context switch statistics
func CollectContextSwitchMetrics() (*ContextSwitchMetrics, error) {
	// Use vm_stat for some of this info on macOS
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	metrics := &ContextSwitchMetrics{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimSuffix(line, ".")
		
		if strings.Contains(line, "Pageins:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				metrics.Interrupts, _ = strconv.ParseUint(parts[1], 10, 64)
			}
		}
	}

	// Get context switches using sysctl (if available)
	cmd = exec.Command("sysctl", "-n", "vm.stats.sys.v_swtch")
	output, err = cmd.Output()
	if err == nil {
		metrics.ContextSwitches, _ = strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	}

	return metrics, nil
}

