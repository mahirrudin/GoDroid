// Package device provides device discovery and management for both
// emulators and real Android devices.
package device

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mahirrudin/godroid/internal/platform"
)

// DeviceType represents the type of Android device.
type DeviceType int

const (
	// DeviceTypeEmulator is an Android emulator
	DeviceTypeEmulator DeviceType = iota
	// DeviceTypeUSB is a USB-connected real device
	DeviceTypeUSB
	// DeviceTypeRemote is a wireless/remote ADB device
	DeviceTypeRemote
)

// Device represents an Android device (emulator or real).
type Device struct {
	Serial      string     // ADB serial (e.g., "emulator-5554", "RF8M12345", "192.168.1.100:5555")
	Type        DeviceType // Emulator, USB, Remote
	Model       string     // Device model name
	Product     string     // Product name
	TransportID string     // ADB transport ID
	State       string     // device, offline, unauthorized
}

// TypeString returns a human-readable device type.
func (d *Device) TypeString() string {
	switch d.Type {
	case DeviceTypeEmulator:
		return "Emulator"
	case DeviceTypeUSB:
		return "USB"
	case DeviceTypeRemote:
		return "Remote"
	default:
		return "Unknown"
	}
}

// DisplayName returns a formatted display name for the device.
func (d *Device) DisplayName() string {
	model := d.Model
	if model == "" {
		model = d.Product
	}
	if model == "" {
		model = "Unknown"
	}
	return fmt.Sprintf("%s (%s) [%s]", d.Serial, model, d.TypeString())
}

// IsOnline returns true if the device is online and ready.
func (d *Device) IsOnline() bool {
	return d.State == "device"
}

// Manager handles device discovery and selection.
type Manager struct {
	sdkPath string
	adbPath string
	current *Device
}

// NewManager creates a new device manager.
func NewManager(sdkPath string) *Manager {
	adbPath := filepath.Join(sdkPath, "platform-tools", "adb")
	if platform.IsWindows() {
		adbPath += ".exe"
	}

	return &Manager{
		sdkPath: sdkPath,
		adbPath: adbPath,
	}
}

// getADBPath returns the path to adb, or "adb" if not found in SDK.
func (m *Manager) getADBPath() string {
	if platform.FileExists(m.adbPath) {
		return m.adbPath
	}
	return "adb"
}

// ListDevices returns all connected Android devices.
func (m *Manager) ListDevices() ([]Device, error) {
	adb := m.getADBPath()

	// Run: adb devices -l
	cmd := exec.Command(adb, "devices", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	return m.parseDeviceList(string(output)), nil
}

// parseDeviceList parses the output of "adb devices -l".
func (m *Manager) parseDeviceList(output string) []Device {
	devices := make([]Device, 0)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip header and empty lines
		if line == "" || strings.HasPrefix(line, "List of devices") {
			continue
		}

		// Parse device line: "serial state props..."
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		device := Device{
			Serial: parts[0],
			State:  parts[1],
		}

		// Determine device type from serial
		device.Type = m.detectDeviceType(device.Serial)

		// Parse additional properties (model:xxx product:xxx)
		for _, part := range parts[2:] {
			if strings.HasPrefix(part, "model:") {
				device.Model = strings.TrimPrefix(part, "model:")
			} else if strings.HasPrefix(part, "product:") {
				device.Product = strings.TrimPrefix(part, "product:")
			} else if strings.HasPrefix(part, "transport_id:") {
				device.TransportID = strings.TrimPrefix(part, "transport_id:")
			}
		}

		devices = append(devices, device)
	}

	return devices
}

// detectDeviceType determines the device type from its serial.
func (m *Manager) detectDeviceType(serial string) DeviceType {
	// Emulators typically have "emulator-" prefix
	if strings.HasPrefix(serial, "emulator-") {
		return DeviceTypeEmulator
	}

	// Remote devices have IP:port format
	if strings.Contains(serial, ":") {
		return DeviceTypeRemote
	}

	// Otherwise, it's a USB device
	return DeviceTypeUSB
}

// SelectDevice sets the current active device by serial.
func (m *Manager) SelectDevice(serial string) error {
	devices, err := m.ListDevices()
	if err != nil {
		return err
	}

	for _, d := range devices {
		if d.Serial == serial {
			m.current = &d
			return nil
		}
	}

	return fmt.Errorf("device not found: %s", serial)
}

// GetCurrentDevice returns the currently selected device.
func (m *Manager) GetCurrentDevice() *Device {
	return m.current
}

// GetCurrentSerial returns the serial of the current device, or empty string.
func (m *Manager) GetCurrentSerial() string {
	if m.current != nil {
		return m.current.Serial
	}
	return ""
}

// ConnectRemote connects to a remote device via wireless ADB.
func (m *Manager) ConnectRemote(address string) error {
	// Validate address format (should be ip:port or just ip)
	if !strings.Contains(address, ":") {
		address = address + ":5555" // Default ADB port
	}

	adb := m.getADBPath()
	cmd := exec.Command(adb, "connect", address)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect: %w - %s", err, string(output))
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "connected") || strings.Contains(outputStr, "already connected") {
		return nil
	}

	return fmt.Errorf("connection failed: %s", outputStr)
}

// DisconnectRemote disconnects from a remote device.
func (m *Manager) DisconnectRemote(address string) error {
	adb := m.getADBPath()
	cmd := exec.Command(adb, "disconnect", address)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "disconnected") || strings.Contains(outputStr, "error") {
		// "error" case: device was already disconnected
		return nil
	}

	return nil
}

// DisconnectAll disconnects all remote devices.
func (m *Manager) DisconnectAll() error {
	adb := m.getADBPath()
	cmd := exec.Command(adb, "disconnect")
	_, err := cmd.Output()
	return err
}

// RefreshDevices refreshes the device list and updates current device status.
func (m *Manager) RefreshDevices() ([]Device, error) {
	devices, err := m.ListDevices()
	if err != nil {
		return nil, err
	}

	// Update current device if it's still connected
	if m.current != nil {
		found := false
		for _, d := range devices {
			if d.Serial == m.current.Serial {
				m.current = &d
				found = true
				break
			}
		}
		if !found {
			m.current = nil // Device disconnected
		}
	}

	return devices, nil
}

// HasDevices returns true if any devices are connected.
func (m *Manager) HasDevices() bool {
	devices, err := m.ListDevices()
	if err != nil {
		return false
	}
	return len(devices) > 0
}

// GetFirstDevice returns the first available device.
func (m *Manager) GetFirstDevice() *Device {
	devices, err := m.ListDevices()
	if err != nil || len(devices) == 0 {
		return nil
	}
	return &devices[0]
}

// AutoSelectDevice selects a device automatically if none is selected.
// Prefers emulators, then USB devices, then remote.
func (m *Manager) AutoSelectDevice() error {
	if m.current != nil && m.current.IsOnline() {
		return nil // Already have a valid device
	}

	devices, err := m.ListDevices()
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices connected")
	}

	// Prefer emulators first
	for _, d := range devices {
		if d.Type == DeviceTypeEmulator && d.IsOnline() {
			m.current = &d
			return nil
		}
	}

	// Then USB devices
	for _, d := range devices {
		if d.Type == DeviceTypeUSB && d.IsOnline() {
			m.current = &d
			return nil
		}
	}

	// Finally remote devices
	for _, d := range devices {
		if d.IsOnline() {
			m.current = &d
			return nil
		}
	}

	return fmt.Errorf("no online devices found")
}
