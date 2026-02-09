package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetOS(t *testing.T) {
	os := GetOS()
	if os != runtime.GOOS {
		t.Errorf("expected %s, got %s", runtime.GOOS, os)
	}
}

func TestGetArch(t *testing.T) {
	arch := GetArch()
	if arch != runtime.GOARCH {
		t.Errorf("expected %s, got %s", runtime.GOARCH, arch)
	}
}

func TestIsWindows(t *testing.T) {
	expected := runtime.GOOS == "windows"
	if IsWindows() != expected {
		t.Errorf("IsWindows() = %v, expected %v", IsWindows(), expected)
	}
}

func TestIsDarwin(t *testing.T) {
	expected := runtime.GOOS == "darwin"
	if IsDarwin() != expected {
		t.Errorf("IsDarwin() = %v, expected %v", IsDarwin(), expected)
	}
}

func TestIsLinux(t *testing.T) {
	expected := runtime.GOOS == "linux"
	if IsLinux() != expected {
		t.Errorf("IsLinux() = %v, expected %v", IsLinux(), expected)
	}
}

func TestHomeDirectory(t *testing.T) {
	home := HomeDirectory()
	if home == "" {
		t.Error("HomeDirectory returned empty string")
	}

	// Should match os.UserHomeDir
	expected, err := os.UserHomeDir()
	if err == nil && home != expected {
		t.Errorf("expected %s, got %s", expected, home)
	}
}

func TestFileExists(t *testing.T) {
	// Create a temp file
	tempFile, err := os.CreateTemp("", "platform-test-")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	// Test existing file
	if !FileExists(tempPath) {
		t.Errorf("FileExists(%s) = false, expected true", tempPath)
	}

	// Test non-existent file
	if FileExists("/nonexistent/path/to/file.txt") {
		t.Error("FileExists returned true for non-existent file")
	}
}

func TestDirExists(t *testing.T) {
	// Create a temp directory
	tempDir, err := os.MkdirTemp("", "platform-test-dir-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test existing directory
	if !DirExists(tempDir) {
		t.Errorf("DirExists(%s) = false, expected true", tempDir)
	}

	// Test non-existent directory
	if DirExists("/nonexistent/path/to/directory") {
		t.Error("DirExists returned true for non-existent directory")
	}

	// Test file (should return false for files)
	tempFile, _ := os.CreateTemp(tempDir, "file-")
	tempFile.Close()
	if DirExists(tempFile.Name()) {
		t.Error("DirExists returned true for a file")
	}
}

func TestExpandPath(t *testing.T) {
	home := HomeDirectory()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/documents", filepath.Join(home, "documents")},
		{"~/.config", filepath.Join(home, ".config")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		result := ExpandPath(tt.input)
		if result != tt.expected {
			t.Errorf("ExpandPath(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestEnsureDir(t *testing.T) {
	// Create a temp base directory
	tempBase, err := os.MkdirTemp("", "platform-ensure-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempBase)

	// Test creating nested directory
	nestedDir := filepath.Join(tempBase, "level1", "level2", "level3")
	err = EnsureDir(nestedDir)
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	if !DirExists(nestedDir) {
		t.Error("EnsureDir did not create the directory")
	}

	// Test calling again (should not error)
	err = EnsureDir(nestedDir)
	if err != nil {
		t.Errorf("EnsureDir failed on existing dir: %v", err)
	}
}

func TestCommandExists(t *testing.T) {
	// Test a command that should exist on all platforms
	commonCmds := []string{"go"}

	// At least one should exist
	found := false
	for _, cmd := range commonCmds {
		if CommandExists(cmd) {
			found = true
			break
		}
	}

	if !found {
		t.Skip("No common commands found, skipping CommandExists test")
	}

	// Test non-existent command
	if CommandExists("nonexistent-command-xyz-123") {
		t.Error("CommandExists returned true for non-existent command")
	}
}

func TestGetCommandPath(t *testing.T) {
	// Test non-existent command
	path := GetCommandPath("nonexistent-command-xyz-123")
	if path != "" {
		t.Errorf("GetCommandPath returned non-empty for non-existent command: %s", path)
	}

	// Test existing command (go should exist since we're running Go tests)
	goPath := GetCommandPath("go")
	if goPath == "" {
		t.Skip("go command not in PATH, skipping test")
	}

	if !FileExists(goPath) {
		t.Errorf("GetCommandPath returned path that doesn't exist: %s", goPath)
	}
}
