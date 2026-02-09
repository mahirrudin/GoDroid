package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// DarwinPlatform implements Platform for macOS.
type DarwinPlatform struct {
	BasePlatform
}

// Name returns the platform name.
func (p *DarwinPlatform) Name() string {
	return "darwin"
}

// DefaultSDKPath returns the default Android SDK path on macOS.
func (p *DarwinPlatform) DefaultSDKPath() string {
	// Check ANDROID_HOME first
	if androidHome := os.Getenv("ANDROID_HOME"); androidHome != "" {
		return androidHome
	}
	// Check ANDROID_SDK_ROOT
	if sdkRoot := os.Getenv("ANDROID_SDK_ROOT"); sdkRoot != "" {
		return sdkRoot
	}
	// Default location (Android Studio default)
	return filepath.Join(HomeDirectory(), "Library", "Android", "sdk")
}

// DefaultPythonCmd returns the default Python command on macOS.
func (p *DarwinPlatform) DefaultPythonCmd() string {
	// Check python3 first (preferred on macOS)
	if CommandExists("python3") {
		return "python3"
	}
	// Fallback to python
	if CommandExists("python") {
		return "python"
	}
	return "python3"
}

// GetEnvPath returns the PATH entries.
func (p *DarwinPlatform) GetEnvPath() []string {
	pathEnv := os.Getenv("PATH")
	return strings.Split(pathEnv, ":")
}

// ClearScreen clears the terminal.
func (p *DarwinPlatform) ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// GetArchitecture returns the system architecture for Android.
func (p *DarwinPlatform) GetArchitecture() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		// Apple Silicon Macs
		return "arm64-v8a"
	default:
		return "x86_64"
	}
}

// GetShellConfigFiles returns shell configuration files for PATH setup.
func (p *DarwinPlatform) GetShellConfigFiles() []string {
	home := HomeDirectory()
	return []string{
		filepath.Join(home, ".zshrc"), // Default shell on macOS Catalina+
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".bashrc"),
	}
}

// GetPathSeparator returns the PATH separator.
func (p *DarwinPlatform) GetPathSeparator() string {
	return ":"
}

// GetPathInstructions returns instructions for adding paths on macOS.
func (p *DarwinPlatform) GetPathInstructions(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	pathStr := strings.Join(paths, ":")
	return fmt.Sprintf(`Add the following to your ~/.zshrc (or ~/.bash_profile if using bash):

export ANDROID_HOME="%s"
export PATH="$PATH:%s"

Then run: source ~/.zshrc
`, p.DefaultSDKPath(), pathStr)
}

// IsAppleSilicon returns true if running on Apple Silicon.
func (p *DarwinPlatform) IsAppleSilicon() bool {
	return runtime.GOARCH == "arm64"
}
