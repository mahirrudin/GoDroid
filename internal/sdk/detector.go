// Package sdk provides Android SDK detection and installation functionality.
package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mahirrudin/godroid/internal/platform"
)

// ToolStatus represents the status of an SDK tool.
type ToolStatus struct {
	Name      string
	Installed bool
	Path      string
	Version   string
	InPath    bool
}

// SDKStatus represents the overall SDK status.
type SDKStatus struct {
	SDKPath       string
	SDKExists     bool
	PlatformTools ToolStatus
	Emulator      ToolStatus
	AVDManager    ToolStatus
	SDKManager    ToolStatus
	ADB           ToolStatus
}

// Detector handles SDK detection.
type Detector struct {
	platform platform.Platform
	sdkPath  string
}

// NewDetector creates a new SDK detector.
func NewDetector() *Detector {
	p := platform.Current()
	return &Detector{
		platform: p,
		sdkPath:  p.DefaultSDKPath(),
	}
}

// NewDetectorWithPath creates a detector with a custom SDK path.
func NewDetectorWithPath(sdkPath string) *Detector {
	return &Detector{
		platform: platform.Current(),
		sdkPath:  sdkPath,
	}
}

// SetSDKPath sets the SDK path.
func (d *Detector) SetSDKPath(path string) {
	d.sdkPath = path
}

// GetSDKPath returns the current SDK path.
func (d *Detector) GetSDKPath() string {
	return d.sdkPath
}

// DetectAll detects all SDK components and returns the overall status.
func (d *Detector) DetectAll() *SDKStatus {
	status := &SDKStatus{
		SDKPath:   d.sdkPath,
		SDKExists: platform.DirExists(d.sdkPath),
	}

	status.PlatformTools = d.DetectPlatformTools()
	status.Emulator = d.DetectEmulator()
	status.AVDManager = d.DetectAVDManager()
	status.SDKManager = d.DetectSDKManager()
	status.ADB = d.DetectADB()

	return status
}

// DetectPlatformTools detects Android SDK Platform Tools.
func (d *Detector) DetectPlatformTools() ToolStatus {
	status := ToolStatus{Name: "Platform Tools"}

	// Check in SDK path
	platformToolsDir := filepath.Join(d.sdkPath, "platform-tools")
	if platform.DirExists(platformToolsDir) {
		status.Installed = true
		status.Path = platformToolsDir
	}

	// Check if adb is in PATH
	if adbPath := platform.GetCommandPath("adb"); adbPath != "" {
		status.InPath = true
		if !status.Installed {
			status.Installed = true
			status.Path = filepath.Dir(adbPath)
		}
		// Get version
		if output, err := platform.RunCommand("adb", "--version"); err == nil {
			lines := strings.Split(output, "\n")
			if len(lines) > 0 {
				// Parse "Android Debug Bridge version X.X.X"
				parts := strings.Fields(lines[0])
				if len(parts) >= 5 {
					status.Version = parts[4]
				}
			}
		}
	}

	return status
}

// DetectEmulator detects Android Emulator.
func (d *Detector) DetectEmulator() ToolStatus {
	status := ToolStatus{Name: "Android Emulator"}

	// Check in SDK path
	emulatorDir := filepath.Join(d.sdkPath, "emulator")
	if platform.DirExists(emulatorDir) {
		status.Installed = true
		status.Path = emulatorDir
	}

	// Check if emulator is in PATH
	emulatorCmd := "emulator"
	if platform.IsWindows() {
		emulatorCmd = "emulator.exe"
	}

	if emulatorPath := platform.GetCommandPath(emulatorCmd); emulatorPath != "" {
		status.InPath = true
		if !status.Installed {
			status.Installed = true
			status.Path = filepath.Dir(emulatorPath)
		}
		// Get version
		if output, err := platform.RunCommand(emulatorCmd, "-version"); err == nil {
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if strings.Contains(line, "version") {
					parts := strings.Fields(line)
					for i, p := range parts {
						if p == "version" && i+1 < len(parts) {
							status.Version = parts[i+1]
							break
						}
					}
					break
				}
			}
		}
	}

	return status
}

// DetectAVDManager detects AVD Manager.
func (d *Detector) DetectAVDManager() ToolStatus {
	status := ToolStatus{Name: "AVD Manager"}

	// AVD Manager is part of cmdline-tools
	cmdlineToolsDir := d.findCmdlineTools()
	if cmdlineToolsDir != "" {
		avdmanager := filepath.Join(cmdlineToolsDir, "bin", "avdmanager")
		if platform.IsWindows() {
			avdmanager += ".bat"
		}
		if platform.FileExists(avdmanager) {
			status.Installed = true
			status.Path = avdmanager
		}
	}

	// Check if avdmanager is in PATH
	avdCmd := "avdmanager"
	if platform.IsWindows() {
		avdCmd = "avdmanager.bat"
	}

	if avdPath := platform.GetCommandPath(avdCmd); avdPath != "" {
		status.InPath = true
		if !status.Installed {
			status.Installed = true
			status.Path = avdPath
		}
	}

	return status
}

// DetectSDKManager detects SDK Manager.
func (d *Detector) DetectSDKManager() ToolStatus {
	status := ToolStatus{Name: "SDK Manager"}

	// SDK Manager is part of cmdline-tools
	cmdlineToolsDir := d.findCmdlineTools()
	if cmdlineToolsDir != "" {
		sdkmanager := filepath.Join(cmdlineToolsDir, "bin", "sdkmanager")
		if platform.IsWindows() {
			sdkmanager += ".bat"
		}
		if platform.FileExists(sdkmanager) {
			status.Installed = true
			status.Path = sdkmanager
		}
	}

	// Check if sdkmanager is in PATH
	sdkCmd := "sdkmanager"
	if platform.IsWindows() {
		sdkCmd = "sdkmanager.bat"
	}

	if sdkPath := platform.GetCommandPath(sdkCmd); sdkPath != "" {
		status.InPath = true
		if !status.Installed {
			status.Installed = true
			status.Path = sdkPath
		}
	}

	return status
}

// DetectADB specifically detects adb.
func (d *Detector) DetectADB() ToolStatus {
	status := ToolStatus{Name: "ADB"}

	adbCmd := "adb"
	if platform.IsWindows() {
		adbCmd = "adb.exe"
	}

	// Check in SDK path
	adbPath := filepath.Join(d.sdkPath, "platform-tools", adbCmd)
	if platform.FileExists(adbPath) {
		status.Installed = true
		status.Path = adbPath
	}

	// Check if adb is in PATH
	if pathAdb := platform.GetCommandPath(adbCmd); pathAdb != "" {
		status.InPath = true
		if !status.Installed {
			status.Installed = true
			status.Path = pathAdb
		}
		// Get version
		if output, err := platform.RunCommand(adbCmd, "version"); err == nil {
			lines := strings.Split(output, "\n")
			if len(lines) > 0 {
				parts := strings.Fields(lines[0])
				if len(parts) >= 5 {
					status.Version = parts[4]
				}
			}
		}
	}

	return status
}

// findCmdlineTools finds the cmdline-tools directory.
func (d *Detector) findCmdlineTools() string {
	cmdlineToolsBase := filepath.Join(d.sdkPath, "cmdline-tools")
	if !platform.DirExists(cmdlineToolsBase) {
		return ""
	}

	// Check for 'latest' first
	latestDir := filepath.Join(cmdlineToolsBase, "latest")
	if platform.DirExists(latestDir) {
		return latestDir
	}

	// Look for versioned directories
	entries, err := os.ReadDir(cmdlineToolsBase)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(cmdlineToolsBase, entry.Name())
		}
	}

	return ""
}

// GetMissingPaths returns paths that should be added to PATH.
func (d *Detector) GetMissingPaths() []string {
	var missing []string

	status := d.DetectAll()

	if status.PlatformTools.Installed && !status.PlatformTools.InPath {
		missing = append(missing, status.PlatformTools.Path)
	}

	if status.Emulator.Installed && !status.Emulator.InPath {
		missing = append(missing, status.Emulator.Path)
	}

	// Add cmdline-tools/latest/bin if SDK manager is installed but not in path
	if status.SDKManager.Installed && !status.SDKManager.InPath {
		binDir := filepath.Dir(status.SDKManager.Path)
		missing = append(missing, binDir)
	}

	return missing
}

// GeneratePathInstructions generates OS-specific instructions for PATH setup.
func (d *Detector) GeneratePathInstructions() string {
	missing := d.GetMissingPaths()
	if len(missing) == 0 {
		return ""
	}

	p := platform.Current()

	switch p.Name() {
	case "linux":
		return d.generateLinuxPathInstructions(missing)
	case "darwin":
		return d.generateDarwinPathInstructions(missing)
	case "windows":
		return d.generateWindowsPathInstructions(missing)
	default:
		return ""
	}
}

func (d *Detector) generateLinuxPathInstructions(paths []string) string {
	pathStr := strings.Join(paths, ":")
	return fmt.Sprintf(`Add the following to your ~/.bashrc or ~/.zshrc:

export ANDROID_HOME="%s"
export PATH="$PATH:%s"

Then run: source ~/.bashrc (or source ~/.zshrc)
`, d.sdkPath, pathStr)
}

func (d *Detector) generateDarwinPathInstructions(paths []string) string {
	pathStr := strings.Join(paths, ":")
	return fmt.Sprintf(`Add the following to your ~/.zshrc (or ~/.bash_profile):

export ANDROID_HOME="%s"
export PATH="$PATH:%s"

Then run: source ~/.zshrc
`, d.sdkPath, pathStr)
}

func (d *Detector) generateWindowsPathInstructions(paths []string) string {
	pathStr := strings.Join(paths, ";")
	return fmt.Sprintf(`Add to your System Environment Variables:

1. Press Windows + R, type "sysdm.cpl" and press Enter
2. Click "Advanced" tab → "Environment Variables"
3. Under "User variables":
   - Add new variable: ANDROID_HOME = %s
   - Edit "Path" and add: %s
4. Restart your terminal
`, d.sdkPath, pathStr)
}
