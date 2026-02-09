package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// LinuxPlatform implements Platform for Linux.
type LinuxPlatform struct {
	BasePlatform
}

// Name returns the platform name.
func (p *LinuxPlatform) Name() string {
	return "linux"
}

// DefaultSDKPath returns the default Android SDK path on Linux.
func (p *LinuxPlatform) DefaultSDKPath() string {
	// Check ANDROID_HOME first
	if androidHome := os.Getenv("ANDROID_HOME"); androidHome != "" {
		return androidHome
	}
	// Check ANDROID_SDK_ROOT
	if sdkRoot := os.Getenv("ANDROID_SDK_ROOT"); sdkRoot != "" {
		return sdkRoot
	}
	// Default location
	return filepath.Join(HomeDirectory(), "Android", "Sdk")
}

// DefaultPythonCmd returns the default Python command on Linux.
func (p *LinuxPlatform) DefaultPythonCmd() string {
	// Check python3 first
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
func (p *LinuxPlatform) GetEnvPath() []string {
	pathEnv := os.Getenv("PATH")
	return strings.Split(pathEnv, ":")
}

// ClearScreen clears the terminal.
func (p *LinuxPlatform) ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// GetArchitecture returns the system architecture for Android.
func (p *LinuxPlatform) GetArchitecture() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "arm64-v8a"
	default:
		return "x86_64"
	}
}

// GetShellConfigFiles returns shell configuration files for PATH setup.
func (p *LinuxPlatform) GetShellConfigFiles() []string {
	home := HomeDirectory()
	return []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".profile"),
	}
}

// GetPathSeparator returns the PATH separator.
func (p *LinuxPlatform) GetPathSeparator() string {
	return ":"
}

// GetPathInstructions returns instructions for adding paths on Linux.
func (p *LinuxPlatform) GetPathInstructions(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	pathStr := strings.Join(paths, ":")
	return fmt.Sprintf(`Add the following to your ~/.bashrc or ~/.zshrc:

export ANDROID_HOME="%s"
export PATH="$PATH:%s"

Then run: source ~/.bashrc (or source ~/.zshrc)
`, p.DefaultSDKPath(), pathStr)
}
