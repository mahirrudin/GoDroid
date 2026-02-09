package archive

import (
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZip_Valid(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "archive-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test zip file
	zipPath := filepath.Join(tempDir, "test.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	// Add a file to the zip
	fileWriter, err := zipWriter.Create("test.txt")
	if err != nil {
		t.Fatalf("failed to create file in zip: %v", err)
	}
	_, err = fileWriter.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("failed to write to zip: %v", err)
	}

	// Add a directory with a file
	fileWriter, err = zipWriter.Create("subdir/nested.txt")
	if err != nil {
		t.Fatalf("failed to create nested file in zip: %v", err)
	}
	_, err = fileWriter.Write([]byte("nested content"))
	if err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}

	zipWriter.Close()
	zipFile.Close()

	// Extract the zip
	extractDir := filepath.Join(tempDir, "extracted")
	err = ExtractZip(zipPath, extractDir)
	if err != nil {
		t.Fatalf("ExtractZip failed: %v", err)
	}

	// Verify extracted files
	content, err := os.ReadFile(filepath.Join(extractDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(content))
	}

	nestedContent, err := os.ReadFile(filepath.Join(extractDir, "subdir", "nested.txt"))
	if err != nil {
		t.Fatalf("failed to read nested file: %v", err)
	}
	if string(nestedContent) != "nested content" {
		t.Errorf("expected 'nested content', got '%s'", string(nestedContent))
	}
}

func TestExtractZip_InvalidPath(t *testing.T) {
	err := ExtractZip("/nonexistent/path.zip", "/tmp/output")
	if err == nil {
		t.Error("expected error for non-existent zip file")
	}
}

func TestExtractZip_ZipSlipProtection(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "archive-zipslip-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a malicious zip file with path traversal
	zipPath := filepath.Join(tempDir, "malicious.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	// Attempt to create a file with path traversal
	fileWriter, err := zipWriter.Create("../../../etc/passwd")
	if err != nil {
		t.Fatalf("failed to create file in zip: %v", err)
	}
	_, err = fileWriter.Write([]byte("malicious content"))
	if err != nil {
		t.Fatalf("failed to write to zip: %v", err)
	}

	zipWriter.Close()
	zipFile.Close()

	// Extract should fail due to zip slip protection
	extractDir := filepath.Join(tempDir, "extracted")
	err = ExtractZip(zipPath, extractDir)
	if err == nil {
		t.Error("expected error for zip slip attempt, but got none")
	}
}

func TestExtractGzip_Valid(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "archive-gzip-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a gzip file
	gzPath := filepath.Join(tempDir, "test.gz")
	gzFile, err := os.Create(gzPath)
	if err != nil {
		t.Fatalf("failed to create gz file: %v", err)
	}

	gzWriter := gzip.NewWriter(gzFile)
	_, err = gzWriter.Write([]byte("gzip test content"))
	if err != nil {
		t.Fatalf("failed to write gzip content: %v", err)
	}
	gzWriter.Close()
	gzFile.Close()

	// Extract gzip
	destPath := filepath.Join(tempDir, "output")
	err = ExtractGzip(gzPath, destPath)
	if err != nil {
		t.Fatalf("ExtractGzip failed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "gzip test content" {
		t.Errorf("expected 'gzip test content', got '%s'", string(content))
	}

	// Verify file permissions (should be executable)
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("failed to stat output file: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("expected output file to be executable")
	}
}

func TestExtractGzip_InvalidPath(t *testing.T) {
	err := ExtractGzip("/nonexistent/path.gz", "/tmp/output")
	if err == nil {
		t.Error("expected error for non-existent gzip file")
	}
}

func TestExtractXZ_InvalidPath(t *testing.T) {
	err := ExtractXZ("/nonexistent/path.xz", "/tmp/output")
	if err == nil {
		t.Error("expected error for non-existent xz file")
	}
}
