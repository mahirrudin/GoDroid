// Package avd provides Android Virtual Device management.
package avd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mahirrudin/godroid/internal/platform"
	"github.com/mahirrudin/godroid/internal/ui"
)

// Manager handles AVD creation and management.
type Manager struct {
	sdkPath        string
	avdHome        string
	avdmanagerPath string
}

// NewManager creates a new AVD manager.
func NewManager(sdkPath string) *Manager {
	m := &Manager{
		sdkPath: sdkPath,
		avdHome: getAVDHome(),
	}
	m.avdmanagerPath = m.findAVDManager()
	return m
}

// getAVDHome returns the AVD home directory.
func getAVDHome() string {
	// Check ANDROID_AVD_HOME first
	if avdHome := os.Getenv("ANDROID_AVD_HOME"); avdHome != "" {
		return avdHome
	}
	// Default location
	return filepath.Join(platform.HomeDirectory(), ".android", "avd")
}

// findAVDManager finds the avdmanager executable.
func (m *Manager) findAVDManager() string {
	// Check in SDK path
	cmdlineToolsDir := filepath.Join(m.sdkPath, "cmdline-tools", "latest", "bin")

	avdmanager := "avdmanager"
	if platform.IsWindows() {
		avdmanager = "avdmanager.bat"
	}

	path := filepath.Join(cmdlineToolsDir, avdmanager)
	if platform.FileExists(path) {
		return path
	}

	// Check in PATH
	if pathAVD := platform.GetCommandPath(avdmanager); pathAVD != "" {
		return pathAVD
	}

	return ""
}

// IsAvailable checks if AVD manager is available.
func (m *Manager) IsAvailable() bool {
	return m.avdmanagerPath != ""
}

// AVDConfig represents configuration for creating an AVD.
type AVDConfig struct {
	Name        string // AVD name
	Device      string // Device profile (e.g., "pixel_6")
	SystemImage string // System image package name
	RAMSize     int    // RAM in MB
	HeapSize    int    // VM heap size in MB
	DataSize    string // Data partition size (e.g., "2G")
}

// DefaultConfig returns a default AVD configuration.
func DefaultConfig() *AVDConfig {
	return &AVDConfig{
		Name:        "GoDroid_Pixel6_API31",
		Device:      "pixel_6",
		SystemImage: "system-images;android-31;google_apis;x86_64",
		RAMSize:     2048,
		HeapSize:    512,
		DataSize:    "2G",
	}
}

// CreateAVD creates a new Android Virtual Device.
func (m *Manager) CreateAVD(config *AVDConfig) error {
	if !m.IsAvailable() {
		return fmt.Errorf("avdmanager not found. Install Command Line Tools first")
	}

	ui.Info("Creating AVD: %s", config.Name)
	ui.Println("Device: %s", config.Device)
	ui.Println("System Image: %s", config.SystemImage)

	// Build command
	args := []string{
		"create", "avd",
		"--name", config.Name,
		"--package", config.SystemImage,
		"--device", config.Device,
		"--force", // Overwrite if exists
	}

	cmd := exec.Command(m.avdmanagerPath, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ANDROID_HOME=%s", m.sdkPath),
		fmt.Sprintf("ANDROID_SDK_ROOT=%s", m.sdkPath),
	)

	// Auto-confirm any prompts
	cmd.Stdin = strings.NewReader("no\n") // Don't create custom hardware profile

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create AVD: %w\n%s", err, string(output))
	}

	// Configure additional settings
	if err := m.configureAVD(config); err != nil {
		ui.Warning("Failed to configure AVD settings: %v", err)
	}

	ui.Success("AVD created: %s", config.Name)
	return nil
}

// configureAVD configures additional AVD settings.
func (m *Manager) configureAVD(config *AVDConfig) error {
	// Find the AVD ini file
	avdIniPath := filepath.Join(m.avdHome, config.Name+".avd", "config.ini")
	if !platform.FileExists(avdIniPath) {
		return nil // AVD not found, skip configuration
	}

	// Read existing config
	data, err := os.ReadFile(avdIniPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	configMap := make(map[string]string)

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			configMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Update settings
	if config.RAMSize > 0 {
		configMap["hw.ramSize"] = fmt.Sprintf("%d", config.RAMSize)
	}
	if config.HeapSize > 0 {
		configMap["vm.heapSize"] = fmt.Sprintf("%d", config.HeapSize)
	}
	if config.DataSize != "" {
		configMap["disk.dataPartition.size"] = config.DataSize
	}

	// Write back
	var newLines []string
	for key, value := range configMap {
		newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
	}

	return os.WriteFile(avdIniPath, []byte(strings.Join(newLines, "\n")), 0644)
}

// ListAVDs lists all available AVDs.
func (m *Manager) ListAVDs() ([]string, error) {
	if !m.IsAvailable() {
		return nil, fmt.Errorf("avdmanager not found")
	}

	cmd := exec.Command(m.avdmanagerPath, "list", "avd", "-c")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ANDROID_HOME=%s", m.sdkPath),
		fmt.Sprintf("ANDROID_SDK_ROOT=%s", m.sdkPath),
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list AVDs: %w", err)
	}

	var avds []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			avds = append(avds, line)
		}
	}

	return avds, nil
}

// DeleteAVD deletes an AVD.
func (m *Manager) DeleteAVD(name string) error {
	if !m.IsAvailable() {
		return fmt.Errorf("avdmanager not found")
	}

	ui.Info("Deleting AVD: %s", name)

	cmd := exec.Command(m.avdmanagerPath, "delete", "avd", "--name", name)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ANDROID_HOME=%s", m.sdkPath),
		fmt.Sprintf("ANDROID_SDK_ROOT=%s", m.sdkPath),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete AVD: %w\n%s", err, string(output))
	}

	ui.Success("AVD deleted: %s", name)
	return nil
}

// AVDExists checks if an AVD exists.
func (m *Manager) AVDExists(name string) bool {
	avds, err := m.ListAVDs()
	if err != nil {
		return false
	}

	for _, avd := range avds {
		if avd == name {
			return true
		}
	}
	return false
}

// ListDevices lists available device definitions.
func (m *Manager) ListDevices() ([]DeviceProfile, error) {
	if !m.IsAvailable() {
		return nil, fmt.Errorf("avdmanager not found")
	}

	cmd := exec.Command(m.avdmanagerPath, "list", "device", "-c")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ANDROID_HOME=%s", m.sdkPath),
		fmt.Sprintf("ANDROID_SDK_ROOT=%s", m.sdkPath),
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	var devices []DeviceProfile
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			devices = append(devices, DeviceProfile{
				ID:   line,
				Name: line,
			})
		}
	}

	return devices, nil
}
