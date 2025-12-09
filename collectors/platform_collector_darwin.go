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

// PlatformInfo represents platform/OS information
type PlatformInfo struct {
	OSName          string `json:"os_name"`           // e.g., "macOS"
	OSVersion       string `json:"os_version"`        // e.g., "14.2.1"
	OSBuildVersion  string `json:"os_build_version"`  // e.g., "23C71"
	KernelName      string `json:"kernel_name"`       // e.g., "Darwin"
	KernelVersion   string `json:"kernel_version"`    // e.g., "23.2.0"
	KernelRelease   string `json:"kernel_release"`    // Full kernel release string
	Hostname        string `json:"hostname"`          // Machine hostname
	Machine         string `json:"machine"`           // e.g., "arm64" or "x86_64"
	Model           string `json:"model"`             // e.g., "MacBookPro18,3"
	ModelName       string `json:"model_name"`        // e.g., "MacBook Pro (14-inch, 2021)"
	IsAppleSilicon  bool   `json:"is_apple_silicon"`  // true if ARM-based Mac
	IsRosetta       bool   `json:"is_rosetta"`        // true if running under Rosetta 2
	SerialNumber    string `json:"serial_number"`     // Machine serial number
	CPUBrand        string `json:"cpu_brand"`         // CPU brand string
	PhysicalCPUs    int    `json:"physical_cpus"`     // Number of physical CPUs
	LogicalCPUs     int    `json:"logical_cpus"`      // Number of logical CPUs
	MemoryGB        int    `json:"memory_gb"`         // Total memory in GB
}

// CollectPlatformInfo collects platform and hardware information
func CollectPlatformInfo() (*PlatformInfo, error) {
	info := &PlatformInfo{
		OSName: "macOS",
	}

	// Get OS version using sw_vers
	if version, err := getCommandOutput("sw_vers", "-productVersion"); err == nil {
		info.OSVersion = version
	}

	if buildVersion, err := getCommandOutput("sw_vers", "-buildVersion"); err == nil {
		info.OSBuildVersion = buildVersion
	}

	// Get kernel info using uname
	if kernelName, err := getCommandOutput("uname", "-s"); err == nil {
		info.KernelName = kernelName
	}

	if kernelVersion, err := getCommandOutput("uname", "-r"); err == nil {
		info.KernelVersion = kernelVersion
	}

	if kernelRelease, err := getCommandOutput("uname", "-v"); err == nil {
		info.KernelRelease = kernelRelease
	}

	if hostname, err := getCommandOutput("hostname"); err == nil {
		info.Hostname = hostname
	}

	if machine, err := getCommandOutput("uname", "-m"); err == nil {
		info.Machine = machine
		info.IsAppleSilicon = strings.Contains(strings.ToLower(machine), "arm")
	}

	// Get model info using sysctl
	if model, err := getCommandOutput("sysctl", "-n", "hw.model"); err == nil {
		info.Model = model
	}

	// Get model name from system_profiler
	if modelName, err := getModelName(); err == nil {
		info.ModelName = modelName
	}

	// Check if running under Rosetta 2
	if isRosetta, err := checkRosetta(); err == nil {
		info.IsRosetta = isRosetta
	}

	// Get CPU brand
	if cpuBrand, err := getCommandOutput("sysctl", "-n", "machdep.cpu.brand_string"); err == nil {
		info.CPUBrand = cpuBrand
	} else {
		// On Apple Silicon, use different approach
		info.CPUBrand = getAppleSiliconCPUName()
	}

	// Get CPU counts
	if physicalCPUs, err := getCommandOutput("sysctl", "-n", "hw.physicalcpu"); err == nil {
		info.PhysicalCPUs, _ = strconv.Atoi(physicalCPUs)
	}

	if logicalCPUs, err := getCommandOutput("sysctl", "-n", "hw.logicalcpu"); err == nil {
		info.LogicalCPUs, _ = strconv.Atoi(logicalCPUs)
	}

	// Get memory
	if memsize, err := getCommandOutput("sysctl", "-n", "hw.memsize"); err == nil {
		if bytes, err := strconv.ParseUint(memsize, 10, 64); err == nil {
			info.MemoryGB = int(bytes / (1024 * 1024 * 1024))
		}
	}

	// Get serial number (may require admin privileges)
	if serial, err := getSerialNumber(); err == nil {
		info.SerialNumber = serial
	}

	return info, nil
}

// getCommandOutput executes a command and returns trimmed output
func getCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getModelName gets the human-readable model name from system_profiler
func getModelName() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Model Name:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("model name not found")
}

// checkRosetta checks if the process is running under Rosetta 2
func checkRosetta() (bool, error) {
	cmd := exec.Command("sysctl", "-n", "sysctl.proc_translated")
	output, err := cmd.Output()
	if err != nil {
		// ENOENT means the key doesn't exist (not running under Rosetta)
		return false, nil
	}

	value := strings.TrimSpace(string(output))
	return value == "1", nil
}

// getAppleSiliconCPUName returns the CPU name for Apple Silicon
func getAppleSiliconCPUName() string {
	// Try to get chip info from system_profiler
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "Apple Silicon"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chip:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return "Apple Silicon"
}

// getSerialNumber gets the machine serial number
func getSerialNumber() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Serial Number") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("serial number not found")
}

// HardwareMetrics represents hardware-related metrics
type HardwareMetrics struct {
	CPUTemperature    float64 `json:"cpu_temperature,omitempty"`     // CPU temperature in Celsius
	FanSpeed          int     `json:"fan_speed,omitempty"`           // Fan speed in RPM
	BatteryHealth     int     `json:"battery_health,omitempty"`      // Battery health percentage
	BatteryCycleCount int     `json:"battery_cycle_count,omitempty"` // Battery cycle count
}

// CollectHardwareMetrics collects hardware metrics (temperature, fan, battery)
func CollectHardwareMetrics() (*HardwareMetrics, error) {
	metrics := &HardwareMetrics{}

	// Get battery info from system_profiler
	cmd := exec.Command("system_profiler", "SPPowerDataType")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			
			if strings.Contains(line, "Cycle Count:") {
				re := regexp.MustCompile(`(\d+)`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 2 {
					metrics.BatteryCycleCount, _ = strconv.Atoi(matches[1])
				}
			}
			
			if strings.Contains(line, "Condition:") {
				// Map condition to percentage
				if strings.Contains(line, "Normal") {
					metrics.BatteryHealth = 100
				} else if strings.Contains(line, "Replace Soon") {
					metrics.BatteryHealth = 60
				} else if strings.Contains(line, "Replace Now") {
					metrics.BatteryHealth = 30
				} else if strings.Contains(line, "Service Battery") {
					metrics.BatteryHealth = 10
				}
			}
		}
	}

	// Note: CPU temperature and fan speed require third-party tools or
	// specific frameworks on macOS. The standard APIs don't expose these.
	// Consider using osx-cpu-temp or similar tools if needed.

	return metrics, nil
}

// GetMacOSCodename returns the marketing name for the macOS version
func GetMacOSCodename(version string) string {
	// Extract major version
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return "Unknown"
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "Unknown"
	}

	codenames := map[int]string{
		15: "Sequoia",
		14: "Sonoma",
		13: "Ventura",
		12: "Monterey",
		11: "Big Sur",
		10: "Catalina", // Note: macOS 10.15 is Catalina
	}

	if codename, ok := codenames[major]; ok {
		return codename
	}

	return "Unknown"
}

