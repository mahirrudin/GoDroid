package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// WindowsPlatform implements Platform for Windows.
type WindowsPlatform struct {
	BasePlatform
}

// Name returns the platform name.
func (p *WindowsPlatform) Name() string {
	return "windows"
}

// DefaultSDKPath returns the default Android SDK path on Windows.
func (p *WindowsPlatform) DefaultSDKPath() string {
	// Check ANDROID_HOME first
	if androidHome := os.Getenv("ANDROID_HOME"); androidHome != "" {
		return androidHome
	}
	// Check ANDROID_SDK_ROOT
	if sdkRoot := os.Getenv("ANDROID_SDK_ROOT"); sdkRoot != "" {
		return sdkRoot
	}
	// Default location (Android Studio default)
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(HomeDirectory(), "AppData", "Local")
	}
	return filepath.Join(localAppData, "Android", "Sdk")
}

// DefaultPythonCmd returns the default Python command on Windows.
func (p *WindowsPlatform) DefaultPythonCmd() string {
	// Check python first (Windows default)
	if CommandExists("python") {
		// Verify it's not the Microsoft Store stub
		output, err := RunCommand("python", "--version")
		if err == nil && strings.Contains(output, "Python") {
			return "python"
		}
	}
	// Try python3
	if CommandExists("python3") {
		return "python3"
	}
	return "python"
}

// GetEnvPath returns the PATH entries.
func (p *WindowsPlatform) GetEnvPath() []string {
	pathEnv := os.Getenv("PATH")
	return strings.Split(pathEnv, ";")
}

// ClearScreen clears the terminal.
func (p *WindowsPlatform) ClearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// GetArchitecture returns the system architecture for Android.
func (p *WindowsPlatform) GetArchitecture() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "arm64-v8a"
	default:
		return "x86_64"
	}
}

// GetShellConfigFiles returns config files (not applicable on Windows).
func (p *WindowsPlatform) GetShellConfigFiles() []string {
	// Windows uses system environment variables instead of shell configs
	return []string{}
}

// GetPathSeparator returns the PATH separator.
func (p *WindowsPlatform) GetPathSeparator() string {
	return ";"
}

// GetPathInstructions returns instructions for adding paths on Windows.
func (p *WindowsPlatform) GetPathInstructions(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	pathStr := strings.Join(paths, ";")
	return fmt.Sprintf(`Add the following to your System Environment Variables:

1. Press Windows + R, type "sysdm.cpl" and press Enter
2. Click "Advanced" tab → "Environment Variables"
3. Under "User variables", find "Path" and click "Edit"
4. Add these paths:
   %s

5. Also add a new variable:
   Name: ANDROID_HOME
   Value: %s

6. Restart your terminal after making changes.
`, pathStr, p.DefaultSDKPath())
}

// IsMicrosoftStorePython checks if the Python is the Microsoft Store stub.
func (p *WindowsPlatform) IsMicrosoftStorePython() bool {
	pythonPath := GetCommandPath("python")
	return strings.Contains(pythonPath, "WindowsApps")
}
