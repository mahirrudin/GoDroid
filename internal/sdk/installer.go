package sdk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mahirrudin/godroid/internal/archive"
	"github.com/mahirrudin/godroid/internal/http"
	"github.com/mahirrudin/godroid/internal/platform"
	"github.com/mahirrudin/godroid/internal/ui"
)

// CommandLineToolsURLs contains download URLs for command line tools.
// These URLs are updated periodically by Google, using latest known version.
var CommandLineToolsURLs = map[string]string{
	"linux":   "https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip",
	"darwin":  "https://dl.google.com/android/repository/commandlinetools-mac-11076708_latest.zip",
	"windows": "https://dl.google.com/android/repository/commandlinetools-win-11076708_latest.zip",
}

// Installer handles SDK component installation.
type Installer struct {
	sdkPath    string
	httpClient *http.Client
}

// NewInstaller creates a new SDK installer.
func NewInstaller(sdkPath string) *Installer {
	return &Installer{
		sdkPath:    sdkPath,
		httpClient: http.DownloadClient(),
	}
}

// EnsureSDKDirectory creates the SDK directory if it doesn't exist.
func (i *Installer) EnsureSDKDirectory() error {
	return platform.EnsureDir(i.sdkPath)
}

// DownloadCommandLineTools downloads the command line tools for the current OS.
func (i *Installer) DownloadCommandLineTools() (string, error) {
	osName := runtime.GOOS
	url, ok := CommandLineToolsURLs[osName]
	if !ok {
		return "", fmt.Errorf("unsupported OS: %s", osName)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "godroid-cmdline-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	zipPath := filepath.Join(tempDir, "commandlinetools.zip")

	ui.Info("Downloading Android Command Line Tools...")
	ui.Println("URL: %s", url)

	// Download with progress
	err = i.httpClient.Download(url, zipPath, func(current, total int64) {
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
	ui.Success("Download complete")

	return zipPath, nil
}

// InstallCommandLineTools installs command line tools from a zip file.
func (i *Installer) InstallCommandLineTools(zipPath string) error {
	ui.Info("Installing Command Line Tools...")

	// Ensure SDK directory exists
	if err := i.EnsureSDKDirectory(); err != nil {
		return fmt.Errorf("failed to create SDK directory: %w", err)
	}

	// Create cmdline-tools directory
	cmdlineToolsDir := filepath.Join(i.sdkPath, "cmdline-tools")
	if err := platform.EnsureDir(cmdlineToolsDir); err != nil {
		return fmt.Errorf("failed to create cmdline-tools directory: %w", err)
	}

	// Extract zip to temp location first
	tempExtract, err := os.MkdirTemp("", "godroid-extract-")
	if err != nil {
		return fmt.Errorf("failed to create temp extract directory: %w", err)
	}
	defer os.RemoveAll(tempExtract)

	if err := archive.ExtractZip(zipPath, tempExtract); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Move cmdline-tools to latest directory
	srcDir := filepath.Join(tempExtract, "cmdline-tools")
	dstDir := filepath.Join(cmdlineToolsDir, "latest")

	// Remove existing latest if present
	if platform.DirExists(dstDir) {
		os.RemoveAll(dstDir)
	}

	// Move extracted directory
	if err := os.Rename(srcDir, dstDir); err != nil {
		// If rename fails (cross-device), copy instead
		if err := platform.CopyDir(srcDir, dstDir); err != nil {
			return fmt.Errorf("failed to install cmdline-tools: %w", err)
		}
	}

	ui.Success("Command Line Tools installed to %s", dstDir)
	return nil
}

// InstallPlatformTools installs Android SDK Platform Tools using sdkmanager.
func (i *Installer) InstallPlatformTools() error {
	ui.Info("Installing Android SDK Platform Tools...")
	return i.runSDKManager("platform-tools")
}

// InstallEmulator installs Android Emulator using sdkmanager.
func (i *Installer) InstallEmulator() error {
	ui.Info("Installing Android Emulator...")
	return i.runSDKManager("emulator")
}

// InstallSystemImage installs a system image using sdkmanager.
func (i *Installer) InstallSystemImage(imageSpec string) error {
	ui.Info("Installing System Image: %s", imageSpec)
	return i.runSDKManager(imageSpec)
}

// InstallBuildTools installs build tools using sdkmanager.
func (i *Installer) InstallBuildTools(version string) error {
	ui.Info("Installing Build Tools %s...", version)
	return i.runSDKManager(fmt.Sprintf("build-tools;%s", version))
}

// runSDKManager runs sdkmanager to install a package.
func (i *Installer) runSDKManager(pkg string) error {
	sdkmanager := i.getSDKManagerPath()
	if sdkmanager == "" {
		return fmt.Errorf("sdkmanager not found. Install Command Line Tools first")
	}

	// Accept licenses first
	ui.Info("Accepting licenses...")
	licenseCmd := exec.Command(sdkmanager, "--licenses")
	licenseCmd.Env = append(os.Environ(), fmt.Sprintf("ANDROID_HOME=%s", i.sdkPath))

	// Auto-accept licenses by writing "y" to stdin
	licenseCmd.Stdin = strings.NewReader(strings.Repeat("y\n", 20))
	licenseCmd.Run() // Ignore errors, some licenses might already be accepted

	// Install package
	ui.Info("Installing %s...", pkg)
	cmd := exec.Command(sdkmanager, pkg)
	cmd.Env = append(os.Environ(), fmt.Sprintf("ANDROID_HOME=%s", i.sdkPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader("y\n") // Accept any prompts

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sdkmanager failed: %w", err)
	}

	ui.Success("Installed %s", pkg)
	return nil
}

// getSDKManagerPath returns the path to sdkmanager.
func (i *Installer) getSDKManagerPath() string {
	// Check in SDK path
	cmdlineToolsDir := filepath.Join(i.sdkPath, "cmdline-tools", "latest", "bin")

	sdkmanager := "sdkmanager"
	if platform.IsWindows() {
		sdkmanager = "sdkmanager.bat"
	}

	path := filepath.Join(cmdlineToolsDir, sdkmanager)
	if platform.FileExists(path) {
		return path
	}

	// Check in PATH
	if pathSDK := platform.GetCommandPath(sdkmanager); pathSDK != "" {
		return pathSDK
	}

	return ""
}

// ListAvailableSystemImages lists available system images.
func (i *Installer) ListAvailableSystemImages() ([]string, error) {
	sdkmanager := i.getSDKManagerPath()
	if sdkmanager == "" {
		return nil, fmt.Errorf("sdkmanager not found")
	}

	cmd := exec.Command(sdkmanager, "--list")
	cmd.Env = append(os.Environ(), fmt.Sprintf("ANDROID_HOME=%s", i.sdkPath))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	var images []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "system-images;") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				images = append(images, parts[0])
			}
		}
	}

	return images, nil
}

// ListInstalledSystemImages lists installed system images.
func (i *Installer) ListInstalledSystemImages() ([]string, error) {
	sdkmanager := i.getSDKManagerPath()
	if sdkmanager == "" {
		return nil, fmt.Errorf("sdkmanager not found")
	}

	cmd := exec.Command(sdkmanager, "--list_installed")
	cmd.Env = append(os.Environ(), fmt.Sprintf("ANDROID_HOME=%s", i.sdkPath))

	output, err := cmd.Output()
	if err != nil {
		// Try alternative: check filesystem directly
		return i.listInstalledImagesFromDisk()
	}

	var images []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "system-images;") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				images = append(images, parts[0])
			}
		}
	}

	return images, nil
}

// listInstalledImagesFromDisk checks filesystem for installed system images.
func (i *Installer) listInstalledImagesFromDisk() ([]string, error) {
	systemImagesDir := filepath.Join(i.sdkPath, "system-images")
	if !platform.DirExists(systemImagesDir) {
		return nil, nil
	}

	var images []string

	// Walk through system-images directory
	// Structure: system-images/android-XX/target/arch/
	apiDirs, err := os.ReadDir(systemImagesDir)
	if err != nil {
		return nil, err
	}

	for _, apiDir := range apiDirs {
		if !apiDir.IsDir() {
			continue
		}
		apiPath := filepath.Join(systemImagesDir, apiDir.Name())
		targetDirs, err := os.ReadDir(apiPath)
		if err != nil {
			continue
		}
		for _, targetDir := range targetDirs {
			if !targetDir.IsDir() {
				continue
			}
			targetPath := filepath.Join(apiPath, targetDir.Name())
			archDirs, err := os.ReadDir(targetPath)
			if err != nil {
				continue
			}
			for _, archDir := range archDirs {
				if !archDir.IsDir() {
					continue
				}
				// Found an installed image
				pkg := fmt.Sprintf("system-images;%s;%s;%s",
					apiDir.Name(), targetDir.Name(), archDir.Name())
				images = append(images, pkg)
			}
		}
	}

	return images, nil
}

// IsSystemImageInstalled checks if a specific system image is installed.
func (i *Installer) IsSystemImageInstalled(imageSpec string) bool {
	installed, err := i.ListInstalledSystemImages()
	if err != nil {
		return false
	}

	for _, img := range installed {
		if img == imageSpec {
			return true
		}
	}
	return false
}
