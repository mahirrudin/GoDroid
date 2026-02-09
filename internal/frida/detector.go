// Package frida provides frida-tools detection, installation, and management.
package frida

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mahirrudin/godroid/internal/platform"
	"github.com/mahirrudin/godroid/internal/ui"
)

// Detector handles frida-tools detection.
type Detector struct {
	pythonCmd string
	venvPath  string
}

// NewDetector creates a new frida detector.
func NewDetector() *Detector {
	p := platform.Current()
	return &Detector{
		pythonCmd: p.DefaultPythonCmd(),
	}
}

// NewDetectorWithVenv creates a new frida detector that checks venv first.
func NewDetectorWithVenv(venvPath string) *Detector {
	p := platform.Current()
	return &Detector{
		pythonCmd: p.DefaultPythonCmd(),
		venvPath:  venvPath,
	}
}

// FridaStatus represents the status of frida-tools.
type FridaStatus struct {
	Installed bool
	Version   string
	Path      string
	InVenv    bool
	VenvPath  string
}

// PythonStatus represents the status of Python.
type PythonStatus struct {
	Installed bool
	Version   string
	Path      string
}

// DetectFridaTools detects if frida-tools is installed.
func (d *Detector) DetectFridaTools() FridaStatus {
	status := FridaStatus{}

	// First, check in our venv if configured
	if d.venvPath != "" {
		fridaPath := d.getFridaPathInVenv()
		if fridaPath != "" {
			status.Installed = true
			status.Path = fridaPath
			status.InVenv = true
			status.VenvPath = d.venvPath

			// Get version from venv frida
			output, err := platform.RunCommand(fridaPath, "--version")
			if err == nil {
				status.Version = strings.TrimSpace(output)
			}
			return status
		}
	}

	// Check if frida command exists in system PATH
	fridaPath := platform.GetCommandPath("frida")
	if fridaPath == "" {
		return status
	}

	status.Installed = true
	status.Path = fridaPath

	// Check if it's in a venv
	if strings.Contains(fridaPath, "venv") || strings.Contains(fridaPath, ".venv") {
		status.InVenv = true
		// Try to extract venv path
		parts := strings.Split(fridaPath, string(filepath.Separator))
		for i, part := range parts {
			if part == "venv" || part == ".venv" || strings.Contains(part, "venv") {
				status.VenvPath = filepath.Join(parts[:i+1]...)
				break
			}
		}
	}

	// Get version
	output, err := platform.RunCommand("frida", "--version")
	if err == nil {
		status.Version = strings.TrimSpace(output)
	}

	return status
}

// getFridaPathInVenv returns the path to frida in the venv if it exists.
func (d *Detector) getFridaPathInVenv() string {
	if d.venvPath == "" {
		return ""
	}

	var frida string
	if platform.IsWindows() {
		frida = filepath.Join(d.venvPath, "Scripts", "frida.exe")
	} else {
		frida = filepath.Join(d.venvPath, "bin", "frida")
	}

	if platform.FileExists(frida) {
		return frida
	}
	return ""
}

// DetectPython detects Python installation.
func (d *Detector) DetectPython() PythonStatus {
	status := PythonStatus{}

	// Check python3 first, then python
	for _, pyCmd := range []string{"python3", "python"} {
		pyPath := platform.GetCommandPath(pyCmd)
		if pyPath != "" {
			// Verify it's a real Python (not Windows Store stub)
			output, err := platform.RunCommand(pyCmd, "--version")
			if err == nil && strings.Contains(output, "Python") {
				status.Installed = true
				status.Path = pyPath
				// Parse version from "Python X.Y.Z"
				parts := strings.Fields(output)
				if len(parts) >= 2 {
					status.Version = parts[1]
				}
				return status
			}
		}
	}

	return status
}

// Installer handles frida-tools installation with venv.
type Installer struct {
	pythonCmd    string
	venvPath     string
	fridaVersion string // Empty for latest
}

// NewInstaller creates a new frida installer.
func NewInstaller(venvPath string) *Installer {
	p := platform.Current()
	return &Installer{
		pythonCmd: p.DefaultPythonCmd(),
		venvPath:  venvPath,
	}
}

// SetFridaVersion sets a specific frida version to install.
func (i *Installer) SetFridaVersion(version string) {
	i.fridaVersion = version
}

// CreateVenv creates a Python virtual environment.
func (i *Installer) CreateVenv() error {
	ui.Info("Creating Python virtual environment: %s", i.venvPath)

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(i.venvPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Create venv
	cmd := exec.Command(i.pythonCmd, "-m", "venv", i.venvPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create venv: %w\n%s", err, string(output))
	}

	ui.Success("Virtual environment created")
	return nil
}

// InstallFridaTools installs frida-tools in the venv.
func (i *Installer) InstallFridaTools() error {
	pip := i.getPipPath()
	if pip == "" {
		return fmt.Errorf("pip not found in venv")
	}

	// Upgrade pip first
	ui.Info("Upgrading pip...")
	upgradeCmd := exec.Command(pip, "install", "--upgrade", "pip")
	upgradeCmd.Run() // Ignore errors

	// Install frida-tools
	ui.Info("Installing frida-tools...")

	fridaPkg := "frida-tools"
	if i.fridaVersion != "" {
		fridaPkg = fmt.Sprintf("frida-tools==%s", i.fridaVersion)
	}

	cmd := exec.Command(pip, "install", fridaPkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install frida-tools: %w", err)
	}

	ui.Success("frida-tools installed")
	return nil
}

// InstallObjection installs objection in the venv.
func (i *Installer) InstallObjection() error {
	pip := i.getPipPath()
	if pip == "" {
		return fmt.Errorf("pip not found in venv")
	}

	ui.Info("Installing objection...")

	cmd := exec.Command(pip, "install", "objection")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install objection: %w", err)
	}

	ui.Success("objection installed")
	return nil
}

// InstallReflutter installs reflutter in the venv.
func (i *Installer) InstallReflutter() error {
	pip := i.getPipPath()
	if pip == "" {
		return fmt.Errorf("pip not found in venv")
	}

	ui.Info("Installing reflutter...")

	cmd := exec.Command(pip, "install", "reflutter")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install reflutter: %w", err)
	}

	ui.Success("reflutter installed")
	return nil
}

// getPipPath returns the path to pip in the venv.
func (i *Installer) getPipPath() string {
	var pip string
	if platform.IsWindows() {
		pip = filepath.Join(i.venvPath, "Scripts", "pip.exe")
	} else {
		pip = filepath.Join(i.venvPath, "bin", "pip")
	}

	if platform.FileExists(pip) {
		return pip
	}
	return ""
}

// GetActivationInstructions returns instructions for activating the venv.
func (i *Installer) GetActivationInstructions() string {
	if platform.IsWindows() {
		return fmt.Sprintf(`To activate the virtual environment, run:
  %s\Scripts\activate.bat

Or in PowerShell:
  %s\Scripts\Activate.ps1
`, i.venvPath, i.venvPath)
	}

	return fmt.Sprintf(`To activate the virtual environment, run:
  source %s/bin/activate
`, i.venvPath)
}

// GetFridaPath returns the path to frida in the venv.
func (i *Installer) GetFridaPath() string {
	var frida string
	if platform.IsWindows() {
		frida = filepath.Join(i.venvPath, "Scripts", "frida.exe")
	} else {
		frida = filepath.Join(i.venvPath, "bin", "frida")
	}

	if platform.FileExists(frida) {
		return frida
	}
	return ""
}

// GetFridaPsPath returns the path to frida-ps in the venv.
func (i *Installer) GetFridaPsPath() string {
	var fridaPs string
	if platform.IsWindows() {
		fridaPs = filepath.Join(i.venvPath, "Scripts", "frida-ps.exe")
	} else {
		fridaPs = filepath.Join(i.venvPath, "bin", "frida-ps")
	}

	if platform.FileExists(fridaPs) {
		return fridaPs
	}
	return ""
}

// GetVenvPath returns the venv path.
func (i *Installer) GetVenvPath() string {
	return i.venvPath
}

// VenvExists checks if the venv exists.
func (i *Installer) VenvExists() bool {
	return platform.DirExists(i.venvPath) && i.getPipPath() != ""
}
