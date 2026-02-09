// Package apk provides APK analysis functionality.
package apk

import (
	"archive/zip"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mahirrudin/godroid/internal/ui"
)

// AnalysisResult contains the results of APK analysis.
type AnalysisResult struct {
	APKPath          string
	DecompiledDir    string
	Frameworks       []string
	SSLPinning       []PatternMatch
	RootDetection    []PatternMatch
	HasSmaliAnalysis bool
}

// PatternMatch represents a detected pattern match.
type PatternMatch struct {
	Category string
	Pattern  string
	File     string
	Line     int
}

// Analyzer provides APK analysis functionality.
type Analyzer struct {
	workDir     string
	apktoolPath string
}

// NewAnalyzer creates a new APK analyzer.
func NewAnalyzer(workDir string) *Analyzer {
	return &Analyzer{
		workDir: workDir,
	}
}

// SetApktoolPath sets the path to apktool.jar.
func (a *Analyzer) SetApktoolPath(path string) {
	a.apktoolPath = path
}

// AnalyzeAPK performs full analysis on an APK file.
func (a *Analyzer) AnalyzeAPK(apkPath string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		APKPath: apkPath,
	}

	// Check if APK exists
	if _, err := os.Stat(apkPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("APK file not found: %s", apkPath)
	}

	// Detect frameworks (no decompilation needed)
	ui.Info("Detecting frameworks...")
	frameworks, err := a.DetectFrameworks(apkPath)
	if err != nil {
		ui.Warning("Framework detection failed: %v", err)
	} else {
		result.Frameworks = frameworks
	}

	// Try smali analysis if apktool is available
	if a.apktoolPath != "" {
		ui.Info("Decompiling APK for deep analysis...")
		smaliDir, err := a.decompileAPK(apkPath)
		if err != nil {
			ui.Warning("Decompilation failed: %v", err)
			ui.Tip("SSL pinning and root detection require apktool + Java")
		} else {
			result.HasSmaliAnalysis = true
			result.DecompiledDir = smaliDir

			// Detect SSL pinning
			ui.Info("Detecting SSL pinning patterns...")
			result.SSLPinning = a.scanSmaliForPatterns(smaliDir, SSLPinningPatterns)

			// Detect root detection
			ui.Info("Detecting root detection libraries...")
			result.RootDetection = a.scanSmaliForPatterns(smaliDir, RootDetectionPatterns)
		}
	} else {
		ui.Warning("Apktool not configured - skipping SSL pinning and root detection analysis")
		ui.Tip("Install apktool from Tools menu for deep analysis")
	}

	return result, nil
}

// DetectFrameworks detects frameworks by scanning APK file structure.
func (a *Analyzer) DetectFrameworks(apkPath string) ([]string, error) {
	reader, err := zip.OpenReader(apkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open APK: %w", err)
	}
	defer reader.Close()

	detected := make(map[string]bool)

	for _, file := range reader.File {
		for framework, patterns := range FrameworkPatterns {
			for _, pattern := range patterns {
				if matchesPattern(file.Name, pattern) {
					detected[framework] = true
					break
				}
			}
		}
	}

	var frameworks []string
	for framework := range detected {
		frameworks = append(frameworks, framework)
	}

	if len(frameworks) == 0 {
		frameworks = append(frameworks, "Native Android")
	}

	return frameworks, nil
}

// matchesPattern checks if a filename matches a pattern (supports * and /).
func matchesPattern(filename, pattern string) bool {
	// Simple pattern matching
	if strings.Contains(pattern, "*") {
		// Handle wildcard patterns like "lib/*/libflutter.so"
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(filename, parts[0]) && strings.HasSuffix(filename, parts[1])
		}
	}

	// Check for exact match or contains
	if strings.HasSuffix(pattern, "/") {
		// Directory pattern
		return strings.HasPrefix(filename, pattern) || strings.Contains(filename, "/"+pattern)
	}

	return strings.Contains(filename, pattern)
}

// decompileAPK decompiles an APK using apktool.
func (a *Analyzer) decompileAPK(apkPath string) (string, error) {
	// Create output directory
	baseName := strings.TrimSuffix(filepath.Base(apkPath), ".apk")
	outDir := filepath.Join(a.workDir, "decompiled", baseName)

	// Clean up previous decompilation
	os.RemoveAll(outDir)

	// Run apktool: java -jar apktool.jar d -f -o outDir apkPath
	cmd := exec.Command("java", "-jar", a.apktoolPath, "d", "-f", "-o", outDir, apkPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("apktool failed: %w", err)
	}

	return outDir, nil
}

// scanSmaliForPatterns scans smali files for patterns.
func (a *Analyzer) scanSmaliForPatterns(smaliDir string, patterns map[string][]string) []PatternMatch {
	var matches []PatternMatch

	// Walk through smali directories
	filepath.Walk(smaliDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Only process smali files
		if !strings.HasSuffix(path, ".smali") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		relPath, _ := filepath.Rel(smaliDir, path)

		// Check each pattern
		for category, patternList := range patterns {
			for _, pattern := range patternList {
				for lineNum, line := range lines {
					if strings.Contains(line, pattern) {
						matches = append(matches, PatternMatch{
							Category: category,
							Pattern:  pattern,
							File:     relPath,
							Line:     lineNum + 1,
						})
					}
				}
			}
		}

		return nil
	})

	return matches
}

// PrintResult prints the analysis result.
func (a *Analyzer) PrintResult(result *AnalysisResult) {
	ui.Box("APK Analysis Result")
	ui.NewLine()

	ui.Println("APK: %s", result.APKPath)
	ui.NewLine()

	// Frameworks
	ui.Println("📦 Detected Frameworks:")
	if len(result.Frameworks) == 0 {
		ui.Println("  None detected (Native Android)")
	} else {
		for _, fw := range result.Frameworks {
			ui.Success("  • %s", fw)
		}
	}
	ui.NewLine()

	// SSL Pinning
	ui.Println("🔒 SSL Pinning Detection:")
	if !result.HasSmaliAnalysis {
		ui.Warning("  Skipped (requires apktool)")
	} else if len(result.SSLPinning) == 0 {
		ui.Println("  No SSL pinning detected")
	} else {
		// Group by category
		categories := make(map[string]int)
		for _, match := range result.SSLPinning {
			categories[match.Category]++
		}
		for cat, count := range categories {
			ui.Warning("  • %s (%d matches)", cat, count)
		}
	}
	ui.NewLine()

	// Root Detection
	ui.Println("🔓 Root Detection Libraries:")
	if !result.HasSmaliAnalysis {
		ui.Warning("  Skipped (requires apktool)")
	} else if len(result.RootDetection) == 0 {
		ui.Println("  No root detection libraries detected")
	} else {
		// Group by category
		categories := make(map[string]int)
		for _, match := range result.RootDetection {
			categories[match.Category]++
		}
		for cat, count := range categories {
			ui.Warning("  • %s (%d matches)", cat, count)
		}
	}

	// Show decompiled directory info
	if result.DecompiledDir != "" {
		ui.NewLine()
		ui.Println("📂 Decompiled Directory:")
		ui.Println("  %s", result.DecompiledDir)
		ui.Tip("Use 'Manage Decompiled Files' to delete when done")
	}
}

// GetDecompiledDir returns the path to decompiled directory for an APK.
func (a *Analyzer) GetDecompiledDir(apkPath string) string {
	baseName := strings.TrimSuffix(filepath.Base(apkPath), ".apk")
	return filepath.Join(a.workDir, "decompiled", baseName)
}

// ListDecompiledDirs lists all decompiled directories.
func (a *Analyzer) ListDecompiledDirs() ([]string, error) {
	decompDir := filepath.Join(a.workDir, "decompiled")
	if _, err := os.Stat(decompDir); os.IsNotExist(err) {
		return nil, nil
	}

	var dirs []string
	entries, err := os.ReadDir(decompDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

// DeleteDecompiledDir deletes a decompiled directory.
func (a *Analyzer) DeleteDecompiledDir(name string) error {
	path := filepath.Join(a.workDir, "decompiled", name)
	return os.RemoveAll(path)
}

// DeleteAllDecompiledDirs deletes all decompiled directories.
func (a *Analyzer) DeleteAllDecompiledDirs() error {
	path := filepath.Join(a.workDir, "decompiled")
	return os.RemoveAll(path)
}
