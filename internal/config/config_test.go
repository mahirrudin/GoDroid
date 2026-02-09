package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Check default values are set
	if cfg.BurpAddress != "127.0.0.1:8080" {
		t.Errorf("expected BurpAddress '127.0.0.1:8080', got '%s'", cfg.BurpAddress)
	}

	if cfg.DefaultAVD != "GoDroid_Pixel6_API31" {
		t.Errorf("expected DefaultAVD 'GoDroid_Pixel6_API31', got '%s'", cfg.DefaultAVD)
	}

	if cfg.DefaultDevice != "pixel_6" {
		t.Errorf("expected DefaultDevice 'pixel_6', got '%s'", cfg.DefaultDevice)
	}

	// Work directory should contain .godroid
	if cfg.WorkDir == "" {
		t.Error("WorkDir is empty")
	}
}

func TestConfig_GetRootAVDDir(t *testing.T) {
	cfg := &Config{
		WorkDir: "/home/test/.godroid",
	}

	expected := "/home/test/.godroid/rootAVD"
	result := cfg.GetRootAVDDir()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestConfig_GetTempDir(t *testing.T) {
	cfg := &Config{
		WorkDir: "/home/test/.godroid",
	}

	expected := "/home/test/.godroid/tmp"
	result := cfg.GetTempDir()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Create a temp directory for test config
	tempDir, err := os.MkdirTemp("", "config-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config with custom values
	cfg := &Config{
		SDKPath:       "/custom/sdk",
		BurpAddress:   "192.168.1.1:9090",
		DefaultAVD:    "TestAVD",
		DefaultDevice: "pixel_7",
		WorkDir:       tempDir,
		ScriptsDir:    filepath.Join(tempDir, "scripts"),
		FridaVenvPath: filepath.Join(tempDir, "venv"),
	}

	// We can't easily test Save/Load since configPath() uses platform.HomeDirectory()
	// But we can test EnsureDirectories
	err = cfg.EnsureDirectories()
	if err != nil {
		t.Fatalf("EnsureDirectories failed: %v", err)
	}

	// Verify directories were created
	dirs := []string{cfg.WorkDir, cfg.ScriptsDir}
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %s was not created: %v", dir, err)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}
