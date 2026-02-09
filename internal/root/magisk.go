// Package root provides Magisk and rootAVD integration for emulator rooting.
package root

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mahirrudin/godroid/internal/emulator"
	"github.com/mahirrudin/godroid/internal/http"
	"github.com/mahirrudin/godroid/internal/ui"
)

// Magisk handles Magisk download and installation.
type Magisk struct {
	emu        *emulator.Controller
	httpClient *http.Client
}

// NewMagisk creates a new Magisk manager.
func NewMagisk(emu *emulator.Controller) *Magisk {
	return &Magisk{
		emu:        emu,
		httpClient: http.DefaultClient(),
	}
}

// MagiskVersion contains version information.
type MagiskVersion struct {
	Tag      string
	Filename string
	URL      string
}

// GetLatestVersion gets the latest Magisk version from GitHub.
func (m *Magisk) GetLatestVersion() (*MagiskVersion, error) {
	ui.Info("Checking for latest Magisk version...")

	release, err := m.httpClient.GetLatestGitHubRelease("topjohnwu/Magisk")
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	// Find the APK asset
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".apk") && strings.Contains(asset.Name, "Magisk") {
			return &MagiskVersion{
				Tag:      release.TagName,
				Filename: asset.Name,
				URL:      asset.BrowserDownloadURL,
			}, nil
		}
	}

	return nil, fmt.Errorf("Magisk APK not found in release")
}

// DownloadMagisk downloads the latest Magisk APK.
func (m *Magisk) DownloadMagisk() (string, error) {
	version, err := m.GetLatestVersion()
	if err != nil {
		// Fallback to known version
		ui.Warning("Failed to get latest version, using fallback v29.0")
		version = &MagiskVersion{
			Tag:      "v29.0",
			Filename: "Magisk-v29.0.apk",
			URL:      "https://github.com/topjohnwu/Magisk/releases/download/v29.0/Magisk-v29.0.apk",
		}
	}

	ui.Println("Version: %s", version.Tag)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "godroid-magisk-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	apkPath := filepath.Join(tempDir, version.Filename)

	ui.Info("Downloading %s...", version.Filename)

	downloadClient := http.DownloadClient()
	err = downloadClient.Download(version.URL, apkPath, func(current, total int64) {
		if total > 0 {
			bar := ui.ProgressBar(current, total, 40)
			ui.ClearLine()
			ui.Print("\r%s", bar)
		}
	})
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download: %w", err)
	}
	ui.NewLine()

	ui.Success("Downloaded %s", version.Filename)
	return apkPath, nil
}

// InstallMagisk installs Magisk APK on the emulator.
func (m *Magisk) InstallMagisk(apkPath string) error {
	return m.emu.InstallAPK(apkPath)
}

// VerifyRoot checks if the emulator is rooted.
func (m *Magisk) VerifyRoot() error {
	ui.Info("Verifying root access...")

	output, err := m.emu.ADBShell("su -c 'echo Root granted'")
	if err != nil {
		return fmt.Errorf("failed to execute su: %w", err)
	}

	if !strings.Contains(output, "Root granted") {
		return fmt.Errorf("root access not available: %s", output)
	}

	ui.Success("Root access verified")
	return nil
}

// DownloadAndInstall downloads and installs Magisk.
func (m *Magisk) DownloadAndInstall() error {
	// Check if emulator is running
	if !m.emu.IsEmulatorRunning() {
		return fmt.Errorf("emulator is not running")
	}

	// Download Magisk
	apkPath, err := m.DownloadMagisk()
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Dir(apkPath))

	// Install Magisk
	return m.InstallMagisk(apkPath)
}

// IsMagiskInstalled checks if Magisk is installed on the emulator.
func (m *Magisk) IsMagiskInstalled() bool {
	// Check for Magisk app
	output, err := m.emu.ADBShell("pm list packages | grep -i magisk")
	if err != nil {
		return false
	}
	return strings.Contains(output, "magisk")
}

// GetMagiskVersion gets the installed Magisk version.
func (m *Magisk) GetMagiskVersion() (string, error) {
	output, err := m.emu.ADBShell("su -c 'magisk -v' 2>/dev/null")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// LaunchMagiskApp launches the Magisk app on the emulator.
func (m *Magisk) LaunchMagiskApp() error {
	ui.Info("Launching Magisk app...")

	// Try different package names
	packages := []string{
		"com.topjohnwu.magisk",
		"io.github.vvb2060.magisk", // Magisk Delta
	}

	for _, pkg := range packages {
		_, err := m.emu.ADBShell(fmt.Sprintf("monkey -p %s -c android.intent.category.LAUNCHER 1", pkg))
		if err == nil {
			ui.Success("Magisk app launched")
			return nil
		}
	}

	return fmt.Errorf("could not launch Magisk app - is it installed?")
}

// SetupMagisk configures initial Magisk setup.
func (m *Magisk) SetupMagisk() error {
	ui.Info("Setting up Magisk...")

	// Enable Zygisk
	m.emu.ADBShell("su -c 'magisk --sqlite \"REPLACE INTO settings (key,value) VALUES (\\\"zygisk\\\",1);\"'")

	// Enable deny list
	m.emu.ADBShell("su -c 'magisk --sqlite \"REPLACE INTO settings (key,value) VALUES (\\\"denylist\\\",1);\"'")

	ui.Success("Magisk configured")
	return nil
}
