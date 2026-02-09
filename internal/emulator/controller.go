// Package emulator provides Android emulator control functionality.
package emulator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mahirrudin/godroid/internal/platform"
	"github.com/mahirrudin/godroid/internal/ui"
)

// Controller manages emulator and device lifecycle.
type Controller struct {
	sdkPath      string
	deviceSerial string // Target specific device/emulator
	emulatorPath string
	adbPath      string
}

// NewController creates a new controller (will auto-select device).
func NewController(sdkPath string) *Controller {
	c := &Controller{
		sdkPath: sdkPath,
	}
	c.emulatorPath = c.findEmulator()
	c.adbPath = c.findADB()
	return c
}

// NewControllerWithDevice creates a new controller targeting a specific device.
func NewControllerWithDevice(sdkPath, serial string) *Controller {
	c := &Controller{
		sdkPath:      sdkPath,
		deviceSerial: serial,
	}
	c.emulatorPath = c.findEmulator()
	c.adbPath = c.findADB()
	return c
}

// SetDevice sets the target device serial.
func (c *Controller) SetDevice(serial string) {
	c.deviceSerial = serial
}

// GetDevice returns the current device serial.
func (c *Controller) GetDevice() string {
	return c.deviceSerial
}

// findEmulator finds the emulator executable.
func (c *Controller) findEmulator() string {
	emulator := "emulator"
	if platform.IsWindows() {
		emulator = "emulator.exe"
	}

	// Check in SDK emulator directory
	path := filepath.Join(c.sdkPath, "emulator", emulator)
	if platform.FileExists(path) {
		return path
	}

	// Check in PATH
	if pathEmu := platform.GetCommandPath(emulator); pathEmu != "" {
		return pathEmu
	}

	return ""
}

// findADB finds the adb executable.
func (c *Controller) findADB() string {
	adb := "adb"
	if platform.IsWindows() {
		adb = "adb.exe"
	}

	// Check in SDK platform-tools directory
	path := filepath.Join(c.sdkPath, "platform-tools", adb)
	if platform.FileExists(path) {
		return path
	}

	// Check in PATH
	if pathAdb := platform.GetCommandPath(adb); pathAdb != "" {
		return pathAdb
	}

	return ""
}

// IsEmulatorAvailable checks if emulator is available.
func (c *Controller) IsEmulatorAvailable() bool {
	return c.emulatorPath != ""
}

// IsADBAvailable checks if adb is available.
func (c *Controller) IsADBAvailable() bool {
	return c.adbPath != ""
}

// StartEmulator starts an emulator with the given AVD name.
func (c *Controller) StartEmulator(avdName string, coldBoot bool) error {
	if !c.IsEmulatorAvailable() {
		return fmt.Errorf("emulator not found")
	}

	ui.Info("Starting emulator: %s", avdName)

	args := []string{"-avd", avdName}
	if coldBoot {
		args = append(args, "-no-snapshot-load")
		ui.Println("Cold boot enabled")
	}

	// Add writable system for rooting
	args = append(args, "-writable-system")

	cmd := exec.Command(c.emulatorPath, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ANDROID_HOME=%s", c.sdkPath),
		fmt.Sprintf("ANDROID_SDK_ROOT=%s", c.sdkPath),
	)

	// Start emulator in background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start emulator: %w", err)
	}

	ui.Success("Emulator starting in background (PID: %d)", cmd.Process.Pid)
	return nil
}

// StopEmulator stops the running emulator.
func (c *Controller) StopEmulator() error {
	if !c.IsADBAvailable() {
		return fmt.Errorf("adb not found")
	}

	ui.Info("Stopping device: %s", c.deviceSerial)

	// adb emu kill may return error when connection drops, but output shows OK
	output, err := c.ADBCommand("emu", "kill")
	if err != nil {
		// Check if output indicates success despite error
		if strings.Contains(output, "OK") || strings.Contains(output, "killing") {
			ui.Success("Emulator stopped")
			return nil
		}
		return fmt.Errorf("failed to stop emulator: %w", err)
	}

	ui.Success("Emulator stopped")
	return nil
}

// RebootEmulator reboots the emulator.
func (c *Controller) RebootEmulator() error {
	if !c.IsADBAvailable() {
		return fmt.Errorf("adb not found")
	}

	ui.Info("Rebooting device: %s", c.deviceSerial)

	if _, err := c.ADBCommand("reboot"); err != nil {
		return fmt.Errorf("failed to reboot emulator: %w", err)
	}

	ui.Success("Device rebooting...")
	return nil
}

// IsEmulatorRunning checks if an emulator is running.
func (c *Controller) IsEmulatorRunning() bool {
	if !c.IsADBAvailable() {
		return false
	}

	output, err := platform.RunCommand(c.adbPath, "devices")
	if err != nil {
		return false
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "emulator-") && strings.Contains(line, "device") {
			return true
		}
	}

	return false
}

// WaitForEmulator waits for the emulator to be ready.
func (c *Controller) WaitForEmulator(timeout time.Duration) error {
	if !c.IsADBAvailable() {
		return fmt.Errorf("adb not found")
	}

	ui.Info("Waiting for emulator to be ready...")

	deadline := time.Now().Add(timeout)
	spinner := ui.NewSpinner("Waiting for boot")

	for time.Now().Before(deadline) {
		// Check if device is online
		if c.IsEmulatorRunning() {
			// Check if boot completed
			output, err := platform.RunCommand(c.adbPath, "shell", "getprop", "sys.boot_completed")
			if err == nil && strings.TrimSpace(output) == "1" {
				ui.NewLine()
				ui.Success("Emulator is ready")
				return nil
			}
		}

		ui.ClearLine()
		ui.Print("\r%s", spinner.Next())
		time.Sleep(2 * time.Second)
	}

	ui.NewLine()
	return fmt.Errorf("timeout waiting for emulator")
}

// GetEmulatorArch returns the architecture of the running emulator.
func (c *Controller) GetEmulatorArch() (string, error) {
	if !c.IsADBAvailable() {
		return "", fmt.Errorf("adb not found")
	}

	output, err := platform.RunCommand(c.adbPath, "shell", "getprop", "ro.product.cpu.abi")
	if err != nil {
		return "", fmt.Errorf("failed to get architecture: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// GetEmulatorAPI returns the API level of the running emulator.
func (c *Controller) GetEmulatorAPI() (string, error) {
	if !c.IsADBAvailable() {
		return "", fmt.Errorf("adb not found")
	}

	output, err := platform.RunCommand(c.adbPath, "shell", "getprop", "ro.build.version.sdk")
	if err != nil {
		return "", fmt.Errorf("failed to get API level: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// ADBCommand executes an adb command and returns the output.
// If deviceSerial is set, uses -s flag to target specific device.
func (c *Controller) ADBCommand(args ...string) (string, error) {
	if !c.IsADBAvailable() {
		return "", fmt.Errorf("adb not found")
	}

	// If device serial is set, prepend -s flag
	if c.deviceSerial != "" {
		args = append([]string{"-s", c.deviceSerial}, args...)
	}

	return platform.RunCommand(c.adbPath, args...)
}

// ADBShell executes an adb shell command.
func (c *Controller) ADBShell(shellCmd string) (string, error) {
	return c.ADBCommand("shell", shellCmd)
}

// ADBRoot restarts adb with root permissions.
func (c *Controller) ADBRoot() error {
	ui.Info("Restarting adb as root...")

	output, err := c.ADBCommand("root")
	if err != nil {
		return fmt.Errorf("failed to restart as root: %w\n%s", err, output)
	}

	// Wait a bit for adb to restart
	time.Sleep(2 * time.Second)

	ui.Success("ADB running as root")
	return nil
}

// InstallAPK installs an APK on the device.
func (c *Controller) InstallAPK(apkPath string) error {
	ui.Info("Installing APK: %s", filepath.Base(apkPath))

	output, err := c.ADBCommand("install", "-r", apkPath)
	if err != nil {
		return fmt.Errorf("failed to install APK: %w\n%s", err, output)
	}

	ui.Success("APK installed")
	return nil
}

// PushFile pushes a file to the device.
func (c *Controller) PushFile(localPath, remotePath string) error {
	output, err := c.ADBCommand("push", localPath, remotePath)
	if err != nil {
		return fmt.Errorf("failed to push file: %w\n%s", err, output)
	}
	return nil
}

// PullFile pulls a file from the device.
func (c *Controller) PullFile(remotePath, localPath string) error {
	output, err := c.ADBCommand("pull", remotePath, localPath)
	if err != nil {
		return fmt.Errorf("failed to pull file: %w\n%s", err, output)
	}
	return nil
}

// ListAVDs lists available AVDs using the emulator.
func (c *Controller) ListAVDs() ([]string, error) {
	if !c.IsEmulatorAvailable() {
		return nil, fmt.Errorf("emulator not found")
	}

	output, err := platform.RunCommand(c.emulatorPath, "-list-avds")
	if err != nil {
		return nil, fmt.Errorf("failed to list AVDs: %w", err)
	}

	var avds []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			avds = append(avds, line)
		}
	}

	return avds, nil
}

// HasRoot checks if the emulator has root access.
func (c *Controller) HasRoot() bool {
	output, err := c.ADBShell("su -c 'echo root'")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) == "root"
}

// SetProxy sets the emulator proxy.
func (c *Controller) SetProxy(host string, port int) error {
	ui.Info("Setting proxy to %s:%d", host, port)

	// Set global proxy
	_, err := c.ADBShell(fmt.Sprintf("settings put global http_proxy %s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to set proxy: %w", err)
	}

	ui.Success("Proxy set to %s:%d", host, port)
	return nil
}

// ClearProxy clears the emulator proxy.
func (c *Controller) ClearProxy() error {
	ui.Info("Clearing proxy...")

	_, err := c.ADBShell("settings put global http_proxy :0")
	if err != nil {
		return fmt.Errorf("failed to clear proxy: %w", err)
	}

	ui.Success("Proxy cleared")
	return nil
}

// GetDeviceSerial returns the serial of the connected emulator.
func (c *Controller) GetDeviceSerial() (string, error) {
	output, err := c.ADBCommand("devices")
	if err != nil {
		return "", err
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "emulator-") && strings.Contains(line, "device") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				return parts[0], nil
			}
		}
	}

	return "", fmt.Errorf("no emulator found")
}
