package frida

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mahirrudin/godroid/internal/archive"
	"github.com/mahirrudin/godroid/internal/emulator"
	"github.com/mahirrudin/godroid/internal/http"
	"github.com/mahirrudin/godroid/internal/platform"
	"github.com/mahirrudin/godroid/internal/ui"
)

// Server handles Frida server operations on the emulator.
type Server struct {
	emu        *emulator.Controller
	venvPath   string
	httpClient *http.Client
}

// NewServer creates a new Frida server manager.
func NewServer(emu *emulator.Controller, venvPath string) *Server {
	return &Server{
		emu:        emu,
		venvPath:   venvPath,
		httpClient: http.DownloadClient(),
	}
}

// getFridaPath returns the path to frida executable.
func (s *Server) getFridaPath() string {
	if s.venvPath != "" {
		var frida string
		if platform.IsWindows() {
			frida = filepath.Join(s.venvPath, "Scripts", "frida.exe")
		} else {
			frida = filepath.Join(s.venvPath, "bin", "frida")
		}
		if platform.FileExists(frida) {
			return frida
		}
	}
	// Fallback to system PATH
	return "frida"
}

// GetFridaVersion returns the installed frida-tools version.
func (s *Server) GetFridaVersion() (string, error) {
	fridaPath := s.getFridaPath()
	output, err := platform.RunCommand(fridaPath, "--version")
	if err != nil {
		if s.venvPath != "" {
			return "", fmt.Errorf("frida not installed in venv (%s). Use 'Environment Setup' → 'Install frida-tools' first", s.venvPath)
		}
		return "", fmt.Errorf("frida not installed or not in PATH")
	}
	return strings.TrimSpace(output), nil
}

// DownloadServer downloads the Frida server for the emulator architecture.
func (s *Server) DownloadServer(version, arch string) (string, error) {
	// Map Android architecture names to Frida's naming convention
	fridaArch := mapAndroidArchToFrida(arch)
	ui.Info("Downloading Frida server %s for %s...", version, fridaArch)

	// Construct download URL
	// Format: frida-server-VERSION-android-ARCH.xz
	fileName := fmt.Sprintf("frida-server-%s-android-%s.xz", version, fridaArch)
	downloadURL := fmt.Sprintf("https://github.com/frida/frida/releases/download/%s/%s", version, fileName)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "godroid-frida-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	xzPath := filepath.Join(tempDir, fileName)

	// Download with progress
	err = s.httpClient.Download(downloadURL, xzPath, func(current, total int64) {
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

	// Extract xz file
	ui.Info("Extracting Frida server...")
	serverPath := filepath.Join(tempDir, "frida-server")

	if err := archive.ExtractXZ(xzPath, serverPath); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to extract: %w", err)
	}

	ui.Success("Frida server downloaded")
	return serverPath, nil
}

// mapAndroidArchToFrida converts Android architecture names to Frida's naming convention.
// Android SDK uses: arm64-v8a, armeabi-v7a, x86, x86_64
// Frida uses: arm64, arm, x86, x86_64
func mapAndroidArchToFrida(androidArch string) string {
	archMap := map[string]string{
		"arm64-v8a":   "arm64",
		"armeabi-v7a": "arm",
		"armeabi":     "arm",
		"x86":         "x86",
		"x86_64":      "x86_64",
	}

	if fridaArch, ok := archMap[androidArch]; ok {
		return fridaArch
	}
	// Return as-is if no mapping found (might already be in Frida format)
	return androidArch
}

// InstallServer installs the Frida server on the emulator.
func (s *Server) InstallServer(localPath string) error {
	ui.Info("Installing Frida server to emulator...")

	// Push to /data/local/tmp/
	remotePath := "/data/local/tmp/frida-server"

	if err := s.emu.PushFile(localPath, remotePath); err != nil {
		return fmt.Errorf("failed to push server: %w", err)
	}

	// Make executable
	if _, err := s.emu.ADBShell(fmt.Sprintf("chmod +x %s", remotePath)); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	ui.Success("Frida server installed at %s", remotePath)
	return nil
}

// StartServer starts the Frida server on the emulator.
func (s *Server) StartServer() error {
	ui.Info("Starting Frida server...")

	// Check if running
	if s.IsServerRunning() {
		ui.Warning("Frida server is already running")
		return nil
	}

	// Start in background with su
	cmd := "su -c 'nohup /data/local/tmp/frida-server > /dev/null 2>&1 &'"
	if _, err := s.emu.ADBShell(cmd); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	ui.Success("Frida server started")
	return nil
}

// StopServer stops the Frida server on the emulator.
func (s *Server) StopServer() error {
	ui.Info("Stopping Frida server...")

	// Kill frida-server process
	cmd := "su -c 'pkill -f frida-server'"
	s.emu.ADBShell(cmd) // Ignore errors

	ui.Success("Frida server stopped")
	return nil
}

// IsServerRunning checks if Frida server is running.
func (s *Server) IsServerRunning() bool {
	output, err := s.emu.ADBShell("ps -A | grep frida-server")
	if err != nil {
		return false
	}
	return strings.Contains(output, "frida-server")
}

// IsServerInstalled checks if Frida server is installed on the emulator.
func (s *Server) IsServerInstalled() bool {
	output, err := s.emu.ADBShell("ls /data/local/tmp/frida-server 2>/dev/null")
	if err != nil {
		return false
	}
	return strings.Contains(output, "frida-server")
}

// DownloadAndInstall downloads and installs the Frida server matching the frida-tools version.
func (s *Server) DownloadAndInstall() error {
	// Check if emulator is running
	if !s.emu.IsEmulatorRunning() {
		return fmt.Errorf("emulator is not running")
	}

	// Get frida version
	version, err := s.GetFridaVersion()
	if err != nil {
		return err
	}

	// Get emulator architecture
	arch, err := s.emu.GetEmulatorArch()
	if err != nil {
		return fmt.Errorf("failed to get emulator architecture: %w", err)
	}

	ui.Println("Frida version: %s", version)
	ui.Println("Emulator architecture: %s", arch)

	// Download server
	serverPath, err := s.DownloadServer(version, arch)
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Dir(serverPath))

	// Install server
	return s.InstallServer(serverPath)
}
