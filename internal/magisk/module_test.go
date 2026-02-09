package magisk

import "testing"

func TestBuildMagiskFridaURL(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{
			version:  "17.6.2",
			expected: "https://github.com/ViRb3/magisk-frida/releases/download/17.6.2-1/MagiskFrida-17.6.2-1.zip",
		},
		{
			version:  "16.0.0",
			expected: "https://github.com/ViRb3/magisk-frida/releases/download/16.0.0-1/MagiskFrida-16.0.0-1.zip",
		},
		{
			version:  "15.2.1",
			expected: "https://github.com/ViRb3/magisk-frida/releases/download/15.2.1-1/MagiskFrida-15.2.1-1.zip",
		},
	}

	for _, tt := range tests {
		result := BuildMagiskFridaURL(tt.version)
		if result != tt.expected {
			t.Errorf("BuildMagiskFridaURL(%q) = %q, expected %q",
				tt.version, result, tt.expected)
		}
	}
}

func TestNewModuleManager(t *testing.T) {
	venvPath := "/home/test/.godroid/venv"
	manager := NewModuleManager(nil, venvPath)

	if manager == nil {
		t.Fatal("NewModuleManager returned nil")
	}

	if manager.venvPath != venvPath {
		t.Errorf("expected venvPath %q, got %q", venvPath, manager.venvPath)
	}

	if manager.httpClient == nil {
		t.Error("httpClient is nil")
	}
}

func TestGetFridaPath_Venv(t *testing.T) {
	manager := &ModuleManager{
		venvPath: "/nonexistent/venv",
	}

	// Should fallback to "frida" since venv path doesn't exist
	path := manager.getFridaPath()
	if path != "frida" {
		t.Errorf("expected fallback to 'frida', got %q", path)
	}
}
