// Package platform provides cross-platform OS detection and abstractions.
package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Platform defines the interface for OS-specific operations.
type Platform interface {
	Name() string
	DefaultSDKPath() string
	DefaultPythonCmd() string
	IsCommandAvailable(cmd string) bool
	GetEnvPath() []string
	ClearScreen()
	HomeDir() string
	GetArchitecture() string
	GetShellConfigFiles() []string
	GetPathSeparator() string
	GetPathInstructions(paths []string) string
}

// BasePlatform provides common implementations for Platform interface.
type BasePlatform struct{}

// HomeDir returns the user's home directory.
func (p *BasePlatform) HomeDir() string {
	return HomeDirectory()
}

// IsCommandAvailable checks if a command is available in PATH.
func (p *BasePlatform) IsCommandAvailable(cmd string) bool {
	return CommandExists(cmd)
}

// Current returns the platform implementation for the current OS.
func Current() Platform {
	switch runtime.GOOS {
	case "windows":
		return &WindowsPlatform{}
	case "darwin":
		return &DarwinPlatform{}
	default:
		return &LinuxPlatform{}
	}
}

// GetOS returns the current operating system name.
func GetOS() string {
	return runtime.GOOS
}

// GetArch returns the current CPU architecture.
func GetArch() string {
	return runtime.GOARCH
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsDarwin returns true if running on macOS.
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if running on Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// HomeDirectory returns the user's home directory.
func HomeDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// CommandExists checks if a command is available in PATH.
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetCommandPath returns the full path to a command if it exists.
func GetCommandPath(cmd string) string {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return ""
	}
	return path
}

// RunCommand executes a command and returns output.
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// RunCommandInDir executes a command in a specific directory.
func RunCommandInDir(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// ExpandPath expands ~ to home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		return filepath.Join(HomeDirectory(), path[1:])
	}
	return path
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
