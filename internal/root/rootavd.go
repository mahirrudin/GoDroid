package root

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mahirrudin/godroid/internal/archive"
	"github.com/mahirrudin/godroid/internal/http"
	"github.com/mahirrudin/godroid/internal/platform"
	"github.com/mahirrudin/godroid/internal/ui"
)

const rootAVDURL = "https://gitlab.com/newbit/rootAVD/-/archive/master/rootAVD-master.zip"

// RootAVD handles rootAVD operations for patching emulator system images.
type RootAVD struct {
	sdkPath    string
	rootAVDDir string
	httpClient *http.Client
}

// NewRootAVD creates a new rootAVD manager.
func NewRootAVD(sdkPath, workDir string) *RootAVD {
	return &RootAVD{
		sdkPath:    sdkPath,
		rootAVDDir: filepath.Join(workDir, "rootAVD-master"),
		httpClient: http.DownloadClient(),
	}
}

// IsDownloaded checks if rootAVD is already downloaded.
func (r *RootAVD) IsDownloaded() bool {
	var script string
	if platform.IsWindows() {
		script = filepath.Join(r.rootAVDDir, "rootAVD.bat")
	} else {
		script = filepath.Join(r.rootAVDDir, "rootAVD.sh")
	}
	return platform.FileExists(script)
}

// Download downloads rootAVD from GitLab.
func (r *RootAVD) Download() error {
	ui.Info("Downloading rootAVD...")

	// Create parent directory
	workDir := filepath.Dir(r.rootAVDDir)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	zipPath := filepath.Join(workDir, "rootAVD.zip")

	// Download
	err := r.httpClient.Download(rootAVDURL, zipPath, func(current, total int64) {
		if total > 0 {
			bar := ui.ProgressBar(current, total, 40)
			ui.ClearLine()
			ui.Print("\r%s", bar)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	ui.NewLine()

	// Extract
	ui.Info("Extracting rootAVD...")
	if err := archive.ExtractZip(zipPath, workDir); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Cleanup zip
	os.Remove(zipPath)

	// Make scripts executable on Unix
	if !platform.IsWindows() {
		os.Chmod(filepath.Join(r.rootAVDDir, "rootAVD.sh"), 0755)
	}

	ui.Success("rootAVD downloaded to %s", r.rootAVDDir)
	return nil
}

// ListSystemImages lists available system images for patching.
func (r *RootAVD) ListSystemImages() ([]string, error) {
	if !r.IsDownloaded() {
		if err := r.Download(); err != nil {
			return nil, err
		}
	}

	ui.Info("Scanning for system images...")

	// First, try to find system images by scanning the SDK directory directly
	// This is more reliable than parsing rootAVD output
	images := r.scanSystemImagesDir()

	if len(images) > 0 {
		return images, nil
	}

	// Fallback: try rootAVD ListAllAVDs command
	var cmd *exec.Cmd
	if platform.IsWindows() {
		cmd = exec.Command("cmd", "/c", "rootAVD.bat", "ListAllAVDs")
	} else {
		cmd = exec.Command("./rootAVD.sh", "ListAllAVDs")
	}
	cmd.Dir = r.rootAVDDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ANDROID_HOME=%s", r.sdkPath),
		fmt.Sprintf("ANDROID_SDK_ROOT=%s", r.sdkPath),
	)

	output, _ := cmd.CombinedOutput()

	// Parse output to find actual system image paths
	seenImages := make(map[string]bool)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for system-images path pattern
		if !strings.Contains(line, "system-images/android-") && !strings.Contains(line, "system-images\\android-") {
			continue
		}

		// Extract the path
		var imagePath string
		if idx := strings.Index(line, "system-images/"); idx >= 0 {
			imagePath = line[idx:]
		} else if idx := strings.Index(line, "system-images\\"); idx >= 0 {
			imagePath = line[idx:]
		}

		if !strings.HasSuffix(imagePath, "ramdisk.img") {
			continue
		}

		if idx := strings.Index(imagePath, "ramdisk.img"); idx >= 0 {
			imagePath = imagePath[:idx+len("ramdisk.img")]
		}

		// Skip template paths
		if strings.Contains(imagePath, "$API") {
			continue
		}

		if !seenImages[imagePath] {
			seenImages[imagePath] = true
			images = append(images, imagePath)
		}
	}

	return images, nil
}

// scanSystemImagesDir scans the SDK system-images directory for installed images.
func (r *RootAVD) scanSystemImagesDir() []string {
	var images []string

	systemImagesDir := filepath.Join(r.sdkPath, "system-images")
	if !platform.FileExists(systemImagesDir) {
		return images
	}

	// Walk the system-images directory looking for ramdisk.img files
	filepath.Walk(systemImagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "ramdisk.img" {
			// Convert to relative path from SDK root
			relPath, err := filepath.Rel(r.sdkPath, path)
			if err == nil {
				images = append(images, relPath)
			}
		}
		return nil
	})

	return images
}

// PatchSystemImage patches a system image for root access.
func (r *RootAVD) PatchSystemImage(imagePath string) error {
	if !r.IsDownloaded() {
		if err := r.Download(); err != nil {
			return err
		}
	}

	ui.Info("Patching system image: %s", imagePath)
	ui.Warning("This may take several minutes...")

	var cmd *exec.Cmd
	if platform.IsWindows() {
		cmd = exec.Command("cmd", "/c", "rootAVD.bat", imagePath)
	} else {
		cmd = exec.Command("./rootAVD.sh", imagePath)
	}
	cmd.Dir = r.rootAVDDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ANDROID_HOME=%s", r.sdkPath),
		fmt.Sprintf("ANDROID_SDK_ROOT=%s", r.sdkPath),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to patch system image: %w", err)
	}

	ui.Success("System image patched successfully")
	return nil
}

// GetSystemImagePath returns the path to a system image based on API and arch.
func (r *RootAVD) GetSystemImagePath(apiLevel int, arch, target string) string {
	// Build path: $SDK/system-images/android-XX/TARGET/ARCH/ramdisk.img
	return filepath.Join(
		r.sdkPath,
		"system-images",
		fmt.Sprintf("android-%d", apiLevel),
		target,
		arch,
		"ramdisk.img",
	)
}

// InteractiveRoot provides an interactive rooting workflow.
func (r *RootAVD) InteractiveRoot() error {
	ui.Box("Emulator Rooting with rootAVD")
	ui.NewLine()

	ui.Warning("THIS PROCESS INCLUDES MANUAL STEPS!")
	ui.Println("rootAVD will patch the system image. You must:")
	ui.Println("1. Have an emulator already created (API 31, x86_64 or arm64)")
	ui.Println("2. Have Magisk APK installed on the emulator")
	ui.Println("3. Cold boot the emulator after patching")
	ui.Println("4. Complete Magisk setup in the app")
	ui.NewLine()

	// List available system images
	images, err := r.ListSystemImages()
	if err != nil {
		return err
	}

	if len(images) == 0 {
		return fmt.Errorf("no system images found. Install a system image first")
	}

	ui.Println("Available system images:")
	for i, img := range images {
		ui.Println("  %d. %s", i+1, img)
	}
	ui.NewLine()

	// Get user selection
	choice := ui.GetInput("Select image number (or enter full path)")

	var selectedImage string
	var idx int
	if _, err := fmt.Sscanf(choice, "%d", &idx); err == nil && idx >= 1 && idx <= len(images) {
		selectedImage = images[idx-1]
	} else {
		selectedImage = choice
	}

	// Patch the image
	return r.PatchSystemImage(selectedImage)
}

// GetScriptPath returns the path to the rootAVD script.
func (r *RootAVD) GetScriptPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(r.rootAVDDir, "rootAVD.bat")
	}
	return filepath.Join(r.rootAVDDir, "rootAVD.sh")
}
