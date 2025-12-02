package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// AutoUpdater handles automatic agent updates
type AutoUpdater struct {
	config         *Config
	currentVersion string
	tempDir        string
	lastCheck      time.Time
}

// NewAutoUpdater creates a new auto updater instance
func NewAutoUpdater(config *Config) *AutoUpdater {
	return &AutoUpdater{
		config:         config,
		currentVersion: GetAgentVersion(),
		tempDir:        filepath.Join(os.TempDir(), "roi-agent-update"),
	}
}

// CheckAndUpdate checks for updates and applies them if available
func (au *AutoUpdater) CheckAndUpdate(remoteConfig *RemoteConfig) error {
	if remoteConfig == nil {
		return nil
	}

	// Skip if no update info available
	if remoteConfig.LatestAgentVersion == "" {
		log.Printf("No update information available in remote config")
		return nil
	}

	// Skip if update mode is disabled
	if remoteConfig.UpdateMode == "disabled" {
		log.Printf("Auto-update is disabled")
		return nil
	}

	// Check if update is needed
	if !au.shouldUpdate(remoteConfig) {
		log.Printf("Agent is up to date (current: %s, latest: %s)",
			au.currentVersion, remoteConfig.LatestAgentVersion)
		return nil
	}

	log.Printf("Update available: %s -> %s", au.currentVersion, remoteConfig.LatestAgentVersion)

	// Check update mode
	if remoteConfig.UpdateMode == "notify" {
		log.Printf("Update mode is 'notify' - skipping automatic update, notifying user")
		// TODO: Show notification to user
		return nil
	}

	// Proceed with auto update (mode is "auto" or update is required)
	if remoteConfig.UpdateRequired {
		log.Printf("REQUIRED update detected - proceeding immediately")
	}

	return au.performUpdate(remoteConfig)
}

// shouldUpdate determines if an update is needed
func (au *AutoUpdater) shouldUpdate(rc *RemoteConfig) bool {
	// Don't update dev builds
	if au.currentVersion == "dev" {
		log.Printf("Dev build detected - skipping update")
		return false
	}

	// Compare versions
	if au.currentVersion == rc.LatestAgentVersion {
		return false
	}

	// Check if current version is older than latest
	return compareVersions(au.currentVersion, rc.LatestAgentVersion) < 0
}

// compareVersions compares two semver versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Pad shorter version with zeros
	for len(parts1) < 3 {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < 3 {
		parts2 = append(parts2, "0")
	}

	for i := 0; i < 3; i++ {
		var n1, n2 int
		fmt.Sscanf(parts1[i], "%d", &n1)
		fmt.Sscanf(parts2[i], "%d", &n2)

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// performUpdate downloads and installs the update
func (au *AutoUpdater) performUpdate(rc *RemoteConfig) error {
	// Check for update URL
	if rc.UpdateURL == "" {
		return fmt.Errorf("no update URL provided")
	}

	// Create temp directory
	if err := os.MkdirAll(au.tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(au.tempDir)

	pkgPath := filepath.Join(au.tempDir, "ROI-Agent-update.pkg")

	// Download PKG
	log.Printf("Downloading update from %s...", rc.UpdateURL)
	if err := au.downloadFile(rc.UpdateURL, pkgPath); err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	log.Printf("Download complete: %s", pkgPath)

	// Verify checksum if provided
	if rc.UpdateChecksum != "" {
		log.Printf("Verifying checksum...")
		if err := au.verifyChecksum(pkgPath, rc.UpdateChecksum); err != nil {
			return fmt.Errorf("checksum verification failed: %v", err)
		}
		log.Printf("Checksum verified")
	}

	// Verify code signature (macOS only)
	if runtime.GOOS == "darwin" {
		log.Printf("Verifying code signature...")
		if err := au.verifyCodeSignature(pkgPath); err != nil {
			return fmt.Errorf("code signature verification failed: %v", err)
		}
		log.Printf("Code signature verified")
	}

	// Install update
	log.Printf("Installing update...")
	if err := au.installPkg(pkgPath); err != nil {
		return fmt.Errorf("installation failed: %v", err)
	}

	log.Printf("Update installed successfully. Agent will restart automatically.")
	return nil
}

// downloadFile downloads a file from URL to destination path
func (au *AutoUpdater) downloadFile(url, destPath string) error {
	client := &http.Client{
		Timeout: 10 * time.Minute, // Allow longer timeout for large downloads
	}

	// Create request with API key
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("ROI-Agent/%s", au.currentVersion))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy with progress logging
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Downloaded %d bytes", written)
	return nil
}

// verifyChecksum verifies SHA256 checksum of a file
func (au *AutoUpdater) verifyChecksum(filePath, expectedChecksum string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// verifyCodeSignature verifies the PKG is properly signed (macOS)
func (au *AutoUpdater) verifyCodeSignature(pkgPath string) error {
	// Use pkgutil to check signature
	cmd := exec.Command("pkgutil", "--check-signature", pkgPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("signature check failed: %s", string(output))
	}

	outputStr := string(output)

	// Check for valid signature indicators
	if !strings.Contains(outputStr, "signed") {
		return fmt.Errorf("package is not signed")
	}

	// Optionally verify specific Developer ID
	// You can add your Developer ID check here
	// if !strings.Contains(outputStr, "Developer ID Installer: YOUR_TEAM_NAME") {
	//     return fmt.Errorf("package not signed by expected developer")
	// }

	log.Printf("Package signature:\n%s", outputStr)
	return nil
}

// installPkg installs the PKG file
func (au *AutoUpdater) installPkg(pkgPath string) error {
	// Note: This requires root privileges
	// The agent should be running as LaunchDaemon (root) for this to work
	cmd := exec.Command("installer", "-pkg", pkgPath, "-target", "/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installer failed: %s (error: %v)", string(output), err)
	}

	log.Printf("Installer output: %s", string(output))

	// The postinstall script in the PKG will handle:
	// 1. Stopping the current agent
	// 2. Installing new files
	// 3. Restarting the agent

	return nil
}

// GetCurrentVersion returns the current agent version
func (au *AutoUpdater) GetCurrentVersion() string {
	return au.currentVersion
}

// ForceUpdate triggers an immediate update check and installation
func (au *AutoUpdater) ForceUpdate(updateURL, checksum string) error {
	rc := &RemoteConfig{
		LatestAgentVersion: "forced",
		UpdateURL:          updateURL,
		UpdateChecksum:     checksum,
		UpdateMode:         "auto",
		UpdateRequired:     true,
	}
	return au.performUpdate(rc)
}

