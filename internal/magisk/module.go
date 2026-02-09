// Package magisk provides Magisk module management functionality.
package magisk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mahirrudin/godroid/internal/emulator"
	"github.com/mahirrudin/godroid/internal/http"
	"github.com/mahirrudin/godroid/internal/platform"
	"github.com/mahirrudin/godroid/internal/ui"
)

// ModuleManager handles Magisk module installation.
type ModuleManager struct {
	emu        *emulator.Controller
	venvPath   string
	httpClient *http.Client
}

// NewModuleManager creates a new module manager.
func NewModuleManager(emu *emulator.Controller, venvPath string) *ModuleManager {
	return &ModuleManager{
		emu:        emu,
		venvPath:   venvPath,
		httpClient: http.DownloadClient(),
	}
}

// GetFridaClientVersion gets the frida version from the venv.
func (m *ModuleManager) GetFridaClientVersion() (string, error) {
	fridaPath := m.getFridaPath()
	output, err := platform.RunCommand(fridaPath, "--version")
	if err != nil {
		return "", fmt.Errorf("frida not installed. Run 'Environment Setup' → 'Install frida-tools' first")
	}
	return strings.TrimSpace(output), nil
}

// getFridaPath returns the path to frida executable.
func (m *ModuleManager) getFridaPath() string {
	if m.venvPath != "" {
		var frida string
		if platform.IsWindows() {
			frida = filepath.Join(m.venvPath, "Scripts", "frida.exe")
		} else {
			frida = filepath.Join(m.venvPath, "bin", "frida")
		}
		if platform.FileExists(frida) {
			return frida
		}
	}
	return "frida"
}

// BuildMagiskFridaURL constructs the download URL for MagiskFrida.
// Format: https://github.com/ViRb3/magisk-frida/releases/download/{VERSION}-1/MagiskFrida-{VERSION}-1.zip
func BuildMagiskFridaURL(version string) string {
	return fmt.Sprintf(
		"https://github.com/ViRb3/magisk-frida/releases/download/%s-1/MagiskFrida-%s-1.zip",
		version, version,
	)
}

// InstallMagiskFrida installs the MagiskFrida module matching the frida-tools version.
func (m *ModuleManager) InstallMagiskFrida() error {
	ui.Info("Installing MagiskFrida module...")

	// Check emulator running
	if !m.emu.IsEmulatorRunning() {
		return fmt.Errorf("emulator is not running")
	}

	// Check root
	if !m.emu.HasRoot() {
		return fmt.Errorf("root access required")
	}

	// Get frida client version
	version, err := m.GetFridaClientVersion()
	if err != nil {
		return err
	}
	ui.Println("Frida client version: %s", version)

	// Build download URL
	downloadURL := BuildMagiskFridaURL(version)
	ui.Info("Download URL: %s", downloadURL)

	// Download and install
	return m.downloadAndInstallModule("MagiskFrida", downloadURL)
}

// InstallZygiskSSLUnpinning installs the Zygisk SSL Unpinning module.
func (m *ModuleManager) InstallZygiskSSLUnpinning() error {
	ui.Info("Installing Zygisk SSL Unpinning module...")

	// Check emulator running
	if !m.emu.IsEmulatorRunning() {
		return fmt.Errorf("emulator is not running")
	}

	// Check root
	if !m.emu.HasRoot() {
		return fmt.Errorf("root access required")
	}

	// Get latest release from GitHub
	ui.Info("Fetching latest release...")
	release, err := m.httpClient.GetLatestGitHubRelease("m0szy/Zygisk-SSL-Unpinning")
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	// Find the zip asset
	_, downloadURL, err := release.FindAsset(".zip")
	if err != nil {
		return fmt.Errorf("failed to find module zip: %w", err)
	}

	ui.Println("Latest version: %s", release.TagName)

	return m.downloadAndInstallModule("Zygisk-SSL-Unpinning", downloadURL)
}

// downloadAndInstallModule downloads and installs a Magisk module.
func (m *ModuleManager) downloadAndInstallModule(name, downloadURL string) error {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "godroid-magisk-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	modulePath := filepath.Join(tempDir, name+".zip")

	// Download with progress
	ui.Info("Downloading %s...", name)
	err = m.httpClient.Download(downloadURL, modulePath, func(current, total int64) {
		if total > 0 {
			bar := ui.ProgressBar(current, total, 40)
			ui.ClearLine()
			ui.Print("\r%s", bar)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	ui.NewLine()

	// Push to emulator
	remotePath := fmt.Sprintf("/data/local/tmp/%s.zip", name)
	if err := m.emu.PushFile(modulePath, remotePath); err != nil {
		return fmt.Errorf("failed to push module: %w", err)
	}

	// Install via Magisk
	ui.Info("Installing module via Magisk...")
	_, err = m.emu.ADBShell(fmt.Sprintf("su -c 'magisk --install-module %s'", remotePath))
	if err != nil {
		return fmt.Errorf("failed to install module: %w", err)
	}

	// Clean up remote file
	m.emu.ADBShell(fmt.Sprintf("rm %s", remotePath))

	ui.Success("%s module installed", name)
	ui.Warning("Reboot the emulator for the module to take effect")
	return nil
}
