// Package tools provides external tool management.
package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	httpClient "github.com/mahirrudin/godroid/internal/http"
	"github.com/mahirrudin/godroid/internal/platform"
	"github.com/mahirrudin/godroid/internal/ui"
)

const (
	// ApktoolGitHubAPI is the GitHub API URL for latest release.
	ApktoolGitHubAPI = "https://api.github.com/repos/iBotPeaches/Apktool/releases/latest"
	// ApktoolDownloadURL is the download URL template for apktool.
	ApktoolDownloadURL = "https://github.com/iBotPeaches/Apktool/releases/download/v%s/apktool_%s.jar"
	// ApktoolVersionFile stores the installed version.
	ApktoolVersionFile = "apktool.version"
)

// GitHubRelease represents a GitHub release response.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// Apktool manages apktool installation and usage.
type Apktool struct {
	toolsDir    string
	jarPath     string
	versionPath string
	httpClient  *httpClient.Client
}

// NewApktool creates a new Apktool manager.
func NewApktool(workDir string) *Apktool {
	toolsDir := filepath.Join(workDir, "tools")
	return &Apktool{
		toolsDir:    toolsDir,
		jarPath:     filepath.Join(toolsDir, "apktool.jar"),
		versionPath: filepath.Join(toolsDir, ApktoolVersionFile),
		httpClient:  httpClient.DownloadClient(),
	}
}

// IsInstalled checks if apktool is installed.
func (a *Apktool) IsInstalled() bool {
	return platform.FileExists(a.jarPath)
}

// GetPath returns the path to apktool.jar.
func (a *Apktool) GetPath() string {
	if a.IsInstalled() {
		return a.jarPath
	}
	return ""
}

// GetInstalledVersion returns the currently installed version.
func (a *Apktool) GetInstalledVersion() string {
	if !a.IsInstalled() {
		return ""
	}
	data, err := os.ReadFile(a.versionPath)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

// GetLatestVersion fetches the latest version from GitHub.
func (a *Apktool) GetLatestVersion() (string, error) {
	resp, err := http.Get(ApktoolGitHubAPI)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	// Remove 'v' prefix if present
	version := strings.TrimPrefix(release.TagName, "v")
	return version, nil
}

// CheckUpdate checks if an update is available.
func (a *Apktool) CheckUpdate() (bool, string, string, error) {
	installed := a.GetInstalledVersion()
	latest, err := a.GetLatestVersion()
	if err != nil {
		return false, installed, "", err
	}

	// If version is unknown or different from latest, update is available
	hasUpdate := installed == "unknown" || installed != latest
	return hasUpdate, installed, latest, nil
}

// Download downloads the latest apktool.jar.
func (a *Apktool) Download() error {
	return a.DownloadVersion("")
}

// DownloadVersion downloads a specific version of apktool.jar.
// If version is empty, downloads the latest version.
func (a *Apktool) DownloadVersion(version string) error {
	// Check Java first
	if !IsJavaInstalled() {
		return fmt.Errorf("Java is required but not installed. Please install Java first")
	}

	// Fetch latest version if not specified
	if version == "" {
		ui.Info("Fetching latest apktool version...")
		latest, err := a.GetLatestVersion()
		if err != nil {
			return err
		}
		version = latest
	}

	ui.Info("Downloading apktool v%s...", version)

	// Create tools directory
	if err := os.MkdirAll(a.toolsDir, 0755); err != nil {
		return fmt.Errorf("failed to create tools directory: %w", err)
	}

	// Download apktool
	downloadURL := fmt.Sprintf(ApktoolDownloadURL, version, version)

	err := a.httpClient.Download(downloadURL, a.jarPath, func(current, total int64) {
		if total > 0 {
			bar := ui.ProgressBar(current, total, 40)
			ui.ClearLine()
			ui.Print("\r%s", bar)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to download apktool: %w", err)
	}
	ui.NewLine()

	// Save version info
	if err := os.WriteFile(a.versionPath, []byte(version), 0644); err != nil {
		ui.Warning("Failed to save version info: %v", err)
	}

	ui.Success("Apktool v%s installed to %s", version, a.jarPath)
	return nil
}

// Update updates apktool to the latest version.
func (a *Apktool) Update() error {
	ui.Info("Checking for updates...")

	hasUpdate, installed, latest, err := a.CheckUpdate()
	if err != nil {
		return err
	}

	if !hasUpdate {
		ui.Success("Apktool is already at the latest version (v%s)", installed)
		return nil
	}

	ui.Info("Update available: v%s → v%s", installed, latest)
	return a.DownloadVersion(latest)
}

// Decompile decompiles an APK file.
func (a *Apktool) Decompile(apkPath, outDir string) error {
	if !a.IsInstalled() {
		return fmt.Errorf("apktool is not installed")
	}

	if !IsJavaInstalled() {
		return fmt.Errorf("Java is required but not installed")
	}

	// Clean output directory
	os.RemoveAll(outDir)

	// Run apktool
	cmd := exec.Command("java", "-jar", a.jarPath, "d", "-f", "-o", outDir, apkPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ui.Info("Decompiling %s...", filepath.Base(apkPath))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apktool decompile failed: %w", err)
	}

	ui.Success("Decompiled to %s", outDir)
	return nil
}

// IsJavaInstalled checks if Java is installed and available in PATH.
func IsJavaInstalled() bool {
	cmd := exec.Command("java", "-version")
	err := cmd.Run()
	return err == nil
}

// GetJavaVersion returns the installed Java version.
func GetJavaVersion() string {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	// Parse version from output (usually first line)
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		// Extract version number
		line := lines[0]
		if strings.Contains(line, "version") {
			// Format: java version "X.X.X" or openjdk version "X.X.X"
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end > start {
				return line[start+1 : end]
			}
		}
	}

	return "installed"
}
