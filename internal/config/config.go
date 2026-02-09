// Package config provides configuration management for GoDroid.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mahirrudin/godroid/internal/platform"
)

// Config represents the application configuration.
type Config struct {
	// SDK Configuration
	SDKPath string `json:"sdk_path"`

	// Frida Configuration
	FridaVenvPath string `json:"frida_venv_path"`
	ScriptsDir    string `json:"scripts_dir"`

	// Emulator Configuration
	DefaultAVD         string `json:"default_avd"`
	DefaultDevice      string `json:"default_device"`
	DefaultSystemImage string `json:"default_system_image"`
	DefaultArch        string `json:"default_arch"`

	// Burp Configuration
	BurpAddress string `json:"burp_address"`

	// Working Directory
	WorkDir string `json:"work_dir"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	p := platform.Current()
	homeDir := p.HomeDir()

	workDir := filepath.Join(homeDir, ".godroid")

	return &Config{
		SDKPath:            p.DefaultSDKPath(),
		FridaVenvPath:      filepath.Join(workDir, "venv"),
		ScriptsDir:         filepath.Join(workDir, "scripts"),
		DefaultAVD:         "GoDroid_Pixel6_API31",
		DefaultDevice:      "pixel_6",
		DefaultSystemImage: "system-images;android-31;google_apis;x86_64",
		DefaultArch:        p.GetArchitecture(),
		BurpAddress:        "127.0.0.1:8080",
		WorkDir:            workDir,
	}
}

// configPath returns the path to the config file.
func configPath() string {
	homeDir := platform.HomeDirectory()
	return filepath.Join(homeDir, ".godroid", "config.json")
}

// Load loads configuration from file, or returns default if not exists.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves the configuration to file.
func (c *Config) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(configPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath(), data, 0644)
}

// EnsureDirectories creates necessary directories.
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		c.WorkDir,
		c.ScriptsDir,
		filepath.Dir(c.FridaVenvPath),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetRootAVDDir returns the directory for rootAVD.
func (c *Config) GetRootAVDDir() string {
	return filepath.Join(c.WorkDir, "rootAVD")
}

// GetTempDir returns a temp directory for downloads.
func (c *Config) GetTempDir() string {
	return filepath.Join(c.WorkDir, "tmp")
}
