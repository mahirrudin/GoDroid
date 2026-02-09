// Package archive provides archive and compression extraction utilities.
package archive

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

// ExtractZip extracts a zip file to the destination directory.
// It includes protection against zip slip vulnerabilities.
func ExtractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		// Security check for zip slip vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open archived file: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}
	}

	return nil
}

// ExtractXZ extracts an xz compressed file.
// The extracted file will be made executable (0755).
func ExtractXZ(xzPath, destPath string) error {
	xzFile, err := os.Open(xzPath)
	if err != nil {
		return fmt.Errorf("failed to open xz file: %w", err)
	}
	defer xzFile.Close()

	reader, err := xz.NewReader(xzFile)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("failed to extract xz content: %w", err)
	}

	// Make executable
	return os.Chmod(destPath, 0755)
}

// ExtractGzip extracts a gzip compressed file.
// The extracted file will be made executable (0755).
func ExtractGzip(gzPath, destPath string) error {
	gzFile, err := os.Open(gzPath)
	if err != nil {
		return fmt.Errorf("failed to open gzip file: %w", err)
	}
	defer gzFile.Close()

	reader, err := gzip.NewReader(gzFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("failed to extract gzip content: %w", err)
	}

	return os.Chmod(destPath, 0755)
}
