package apk

import "testing"

func TestMatchesPattern_Exact(t *testing.T) {
	tests := []struct {
		filename string
		pattern  string
		expected bool
	}{
		// Exact matches
		{"assets/flutter_assets/", "assets/flutter_assets/", true},
		{"lib/arm64-v8a/libflutter.so", "lib/*/libflutter.so", true},
		{"lib/x86_64/libflutter.so", "lib/*/libflutter.so", true},

		// Directory patterns
		{"assets/www/index.html", "assets/www/", true},
		{"kotlin/reflect/reflect.kotlin_builtins", "kotlin/", true},

		// Non-matches
		{"assets/data/file.txt", "assets/www/", false},
		{"lib/arm64/other.so", "lib/*/libflutter.so", false},
	}

	for _, tt := range tests {
		result := matchesPattern(tt.filename, tt.pattern)
		if result != tt.expected {
			t.Errorf("matchesPattern(%q, %q) = %v, expected %v",
				tt.filename, tt.pattern, result, tt.expected)
		}
	}
}

func TestMatchesPattern_Wildcard(t *testing.T) {
	tests := []struct {
		filename string
		pattern  string
		expected bool
	}{
		{"lib/arm64-v8a/libapp.so", "lib/*/libapp.so", true},
		{"lib/armeabi-v7a/libapp.so", "lib/*/libapp.so", true},
		{"lib/x86/libapp.so", "lib/*/libapp.so", true},
		{"lib/libapp.so", "lib/*/libapp.so", true}, // Matches: prefix "lib/" and suffix "/libapp.so"
	}

	for _, tt := range tests {
		result := matchesPattern(tt.filename, tt.pattern)
		if result != tt.expected {
			t.Errorf("matchesPattern(%q, %q) = %v, expected %v",
				tt.filename, tt.pattern, result, tt.expected)
		}
	}
}

func TestMatchesPattern_Directory(t *testing.T) {
	tests := []struct {
		filename string
		pattern  string
		expected bool
	}{
		// Directory prefix patterns (ending with /)
		{"assemblies/Mono.Android.dll", "assemblies/", true},
		{"assemblies/", "assemblies/", true},
		{"assets/bin/Data/level0", "assets/bin/Data/", true},

		// Should not match non-prefixes without subdirectory
		{"assemblies-other/file.dll", "assemblies/", false},
	}

	for _, tt := range tests {
		result := matchesPattern(tt.filename, tt.pattern)
		if result != tt.expected {
			t.Errorf("matchesPattern(%q, %q) = %v, expected %v",
				tt.filename, tt.pattern, result, tt.expected)
		}
	}
}

func TestNewAnalyzer(t *testing.T) {
	workDir := "/tmp/test-workdir"
	analyzer := NewAnalyzer(workDir)

	if analyzer == nil {
		t.Fatal("NewAnalyzer returned nil")
	}

	if analyzer.workDir != workDir {
		t.Errorf("expected workDir %q, got %q", workDir, analyzer.workDir)
	}
}

func TestAnalyzer_GetDecompiledDir(t *testing.T) {
	workDir := "/tmp/godroid-test"
	analyzer := NewAnalyzer(workDir)

	result := analyzer.GetDecompiledDir("/path/to/myapp.apk")
	expected := "/tmp/godroid-test/decompiled/myapp"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestAnalyzer_SetApktoolPath(t *testing.T) {
	analyzer := NewAnalyzer("/tmp")

	customPath := "/custom/apktool.jar"
	analyzer.SetApktoolPath(customPath)

	if analyzer.apktoolPath != customPath {
		t.Errorf("expected apktoolPath %q, got %q", customPath, analyzer.apktoolPath)
	}
}
