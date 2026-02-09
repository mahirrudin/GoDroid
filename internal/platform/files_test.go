package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "copyfile-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create source file with content
	srcPath := filepath.Join(tempDir, "source.txt")
	srcContent := []byte("test file content for copy")
	err = os.WriteFile(srcPath, srcContent, 0644)
	if err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Copy the file
	dstPath := filepath.Join(tempDir, "destination.txt")
	err = CopyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination exists
	if !FileExists(dstPath) {
		t.Error("destination file does not exist")
	}

	// Verify content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(dstContent) != string(srcContent) {
		t.Errorf("content mismatch: expected %q, got %q", srcContent, dstContent)
	}

	// Verify permissions are preserved
	srcInfo, _ := os.Stat(srcPath)
	dstInfo, _ := os.Stat(dstPath)
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("permission mismatch: expected %v, got %v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyFile_NonExistent(t *testing.T) {
	err := CopyFile("/nonexistent/source.txt", "/tmp/dest.txt")
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}

func TestCopyDir(t *testing.T) {
	// Create temp directory with structure
	tempDir, err := os.MkdirTemp("", "copydir-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "source")

	// Create nested structure
	os.MkdirAll(filepath.Join(srcDir, "level1", "level2"), 0755)
	os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("root file"), 0644)
	os.WriteFile(filepath.Join(srcDir, "level1", "level1.txt"), []byte("level1 file"), 0644)
	os.WriteFile(filepath.Join(srcDir, "level1", "level2", "level2.txt"), []byte("level2 file"), 0644)

	// Copy directory
	dstDir := filepath.Join(tempDir, "destination")
	err = CopyDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	// Verify structure
	files := []string{
		"root.txt",
		"level1/level1.txt",
		"level1/level2/level2.txt",
	}

	for _, f := range files {
		path := filepath.Join(dstDir, f)
		if !FileExists(path) {
			t.Errorf("expected file %s to exist", path)
		}
	}

	// Verify content of a file
	content, err := os.ReadFile(filepath.Join(dstDir, "level1", "level2", "level2.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content) != "level2 file" {
		t.Errorf("content mismatch: expected 'level2 file', got '%s'", content)
	}
}

func TestCopyDir_NonExistent(t *testing.T) {
	err := CopyDir("/nonexistent/source", "/tmp/dest")
	if err == nil {
		t.Error("expected error for non-existent source directory")
	}
}
