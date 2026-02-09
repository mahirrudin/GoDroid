package frida

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mahirrudin/godroid/internal/emulator"
	"github.com/mahirrudin/godroid/internal/platform"
	"github.com/mahirrudin/godroid/internal/ui"
)

// Script represents a Frida script.
type Script struct {
	Name     string
	Path     string
	IsCustom bool
}

// App represents an installed application.
type App struct {
	Identifier string // Package name
	Name       string // Display name
	PID        int    // Process ID (0 if not running)
}

// ScriptManager handles Frida script operations.
type ScriptManager struct {
	scriptsDir string
	venvPath   string
	emu        *emulator.Controller
}

// NewScriptManager creates a new script manager.
func NewScriptManager(scriptsDir, venvPath string, emu *emulator.Controller) *ScriptManager {
	return &ScriptManager{
		scriptsDir: scriptsDir,
		venvPath:   venvPath,
		emu:        emu,
	}
}

// getFridaPath returns the path to frida executable.
func (sm *ScriptManager) getFridaPath() string {
	if sm.venvPath != "" {
		var frida string
		if platform.IsWindows() {
			frida = filepath.Join(sm.venvPath, "Scripts", "frida.exe")
		} else {
			frida = filepath.Join(sm.venvPath, "bin", "frida")
		}
		if platform.FileExists(frida) {
			return frida
		}
	}
	// Fallback to system PATH
	return "frida"
}

// getFridaPsPath returns the path to frida-ps executable.
func (sm *ScriptManager) getFridaPsPath() string {
	if sm.venvPath != "" {
		var fridaPs string
		if platform.IsWindows() {
			fridaPs = filepath.Join(sm.venvPath, "Scripts", "frida-ps.exe")
		} else {
			fridaPs = filepath.Join(sm.venvPath, "bin", "frida-ps")
		}
		if platform.FileExists(fridaPs) {
			return fridaPs
		}
	}
	// Fallback to system PATH
	return "frida-ps"
}

// getDeviceArgs returns frida args for targeting the selected device.
// Returns "-D <serial>" if a device is selected, otherwise "-U" for USB.
func (sm *ScriptManager) getDeviceArgs() []string {
	if sm.emu != nil {
		serial := sm.emu.GetDevice()
		if serial != "" {
			return []string{"-D", serial}
		}
	}
	return []string{"-U"}
}

// DefaultScripts returns the list of built-in scripts.
var DefaultScripts = []Script{
	{Name: "BYPASS-SSLPINNING", Path: "BYPASS-SSLPINNING.js", IsCustom: false},
	{Name: "BYPASS-ROOT", Path: "BYPASS-ROOT.js", IsCustom: false},
	{Name: "BYPASS-ROOT-SSLPINNING", Path: "BYPASS-ROOT-SSLPINNING.js", IsCustom: false},
}

// ListScripts lists all available scripts (default + custom).
func (sm *ScriptManager) ListScripts() []Script {
	scripts := make([]Script, 0)

	// Add default scripts
	for _, s := range DefaultScripts {
		fullPath := filepath.Join(sm.scriptsDir, s.Path)
		if platform.FileExists(fullPath) {
			scripts = append(scripts, Script{
				Name:     s.Name,
				Path:     fullPath,
				IsCustom: false,
			})
		}
	}

	// List custom scripts
	if platform.DirExists(sm.scriptsDir) {
		entries, err := os.ReadDir(sm.scriptsDir)
		if err == nil {
			defaultNames := make(map[string]bool)
			for _, s := range DefaultScripts {
				defaultNames[s.Path] = true
			}

			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".js") {
					if !defaultNames[entry.Name()] {
						scripts = append(scripts, Script{
							Name:     strings.TrimSuffix(entry.Name(), ".js"),
							Path:     filepath.Join(sm.scriptsDir, entry.Name()),
							IsCustom: true,
						})
					}
				}
			}
		}
	}

	return scripts
}

// AddCustomScript adds a custom Frida script.
func (sm *ScriptManager) AddCustomScript(name, code string) error {
	// Ensure scripts directory exists
	if err := os.MkdirAll(sm.scriptsDir, 0755); err != nil {
		return fmt.Errorf("failed to create scripts directory: %w", err)
	}

	// Validate name
	if name == "" {
		return fmt.Errorf("script name cannot be empty")
	}

	// Add .js extension if not present
	if !strings.HasSuffix(name, ".js") {
		name += ".js"
	}

	// Check for reserved names
	for _, s := range DefaultScripts {
		if s.Path == name {
			return fmt.Errorf("cannot overwrite default script: %s", name)
		}
	}

	// Write script
	scriptPath := filepath.Join(sm.scriptsDir, name)
	if err := os.WriteFile(scriptPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to save script: %w", err)
	}

	ui.Success("Script saved: %s", scriptPath)
	return nil
}

// DeleteCustomScript deletes a custom script.
func (sm *ScriptManager) DeleteCustomScript(name string) error {
	// Check if it's a default script
	for _, s := range DefaultScripts {
		if s.Name == name || s.Path == name {
			return fmt.Errorf("cannot delete default script: %s", name)
		}
	}

	// Add .js extension if not present
	if !strings.HasSuffix(name, ".js") {
		name += ".js"
	}

	scriptPath := filepath.Join(sm.scriptsDir, name)
	if !platform.FileExists(scriptPath) {
		return fmt.Errorf("script not found: %s", name)
	}

	if err := os.Remove(scriptPath); err != nil {
		return fmt.Errorf("failed to delete script: %w", err)
	}

	ui.Success("Script deleted: %s", name)
	return nil
}

// RunScript runs a Frida script on a target application.
func (sm *ScriptManager) RunScript(scriptPath, packageName string) error {
	if !platform.FileExists(scriptPath) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	ui.Info("Running script %s on %s...", filepath.Base(scriptPath), packageName)

	// frida -D <device> -f <package> -l <script>
	args := sm.getDeviceArgs()
	args = append(args, "-f", packageName, "-l", scriptPath)
	cmd := exec.Command(sm.getFridaPath(), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("frida execution failed: %w", err)
	}

	return nil
}

// SpawnAndAttach spawns an app and attaches Frida with a script.
func (sm *ScriptManager) SpawnAndAttach(scriptPath, packageName string) error {
	if !platform.FileExists(scriptPath) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	ui.Info("Spawning %s with script %s...", packageName, filepath.Base(scriptPath))

	// frida -D <device> -f <package> -l <script> --no-pause
	args := sm.getDeviceArgs()
	args = append(args, "-f", packageName, "-l", scriptPath, "--no-pause")
	cmd := exec.Command(sm.getFridaPath(), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("frida execution failed: %w", err)
	}

	return nil
}

// AttachToRunning attaches Frida to a running process.
func (sm *ScriptManager) AttachToRunning(scriptPath, packageName string) error {
	if !platform.FileExists(scriptPath) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	ui.Info("Attaching to %s with script %s...", packageName, filepath.Base(scriptPath))

	// frida -D <device> <package> -l <script>
	args := sm.getDeviceArgs()
	args = append(args, packageName, "-l", scriptPath)
	cmd := exec.Command(sm.getFridaPath(), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("frida execution failed: %w", err)
	}

	return nil
}

// ListInstalledApps lists installed applications using frida-ps.
func (sm *ScriptManager) ListInstalledApps() ([]App, error) {
	ui.Info("Listing installed applications...")

	// frida-ps -D <device> -ai (device, all apps, with identifiers)
	deviceArgs := sm.getDeviceArgs()
	args := append(deviceArgs, "-ai")
	output, err := platform.RunCommand(sm.getFridaPsPath(), args...)
	if err != nil {
		return nil, fmt.Errorf("frida-ps failed: %w", err)
	}

	apps := make([]App, 0)
	lines := strings.Split(output, "\n")

	// Skip header lines
	for i, line := range lines {
		if i < 2 { // Skip "PID  Name  Identifier" and separator
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse: PID  Name  Identifier
		// or:    -    Name  Identifier (for non-running apps)
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			app := App{}

			// Last part is always identifier
			app.Identifier = parts[len(parts)-1]

			// First part might be PID or "-"
			if parts[0] != "-" {
				fmt.Sscanf(parts[0], "%d", &app.PID)
			}

			// Everything in between is the name
			if len(parts) > 2 {
				app.Name = strings.Join(parts[1:len(parts)-1], " ")
			}

			apps = append(apps, app)
		}
	}

	return apps, nil
}

// ListRunningProcesses lists running processes using frida-ps.
func (sm *ScriptManager) ListRunningProcesses() ([]App, error) {
	deviceArgs := sm.getDeviceArgs()
	output, err := platform.RunCommand(sm.getFridaPsPath(), deviceArgs...)
	if err != nil {
		return nil, fmt.Errorf("frida-ps failed: %w", err)
	}

	apps := make([]App, 0)
	lines := strings.Split(output, "\n")

	// Skip header
	for i, line := range lines {
		if i < 2 {
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse: PID  Name
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			app := App{}
			fmt.Sscanf(parts[0], "%d", &app.PID)
			app.Name = strings.Join(parts[1:], " ")
			apps = append(apps, app)
		}
	}

	return apps, nil
}
