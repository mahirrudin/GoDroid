package app

import (
	"fmt"

	"github.com/mahirrudin/godroid/internal/apk"
	"github.com/mahirrudin/godroid/internal/tools"
	"github.com/mahirrudin/godroid/internal/ui"
)

func (h *Handler) RunAPKAnalyzer() error {
	apktoolMgr := tools.NewApktool(h.Config.WorkDir)
	apktoolStatus := ""
	if apktoolMgr.IsInstalled() {
		apktoolStatus = " ✓"
	}

	menu := ui.NewMenu("APK Analyzer")
	menu.AddWithDesc("Apktool Status"+apktoolStatus, "Check/install apktool and Java")
	menu.AddWithDesc("Detect Framework Only", "Quick framework detection (no decompilation)")
	menu.AddWithDesc("Analyze APK File", "Full analysis (framework + SSL pinning + root detection)")
	menu.AddWithDesc("Manage Decompiled Files", "View and delete decompiled APK directories")

	choice := menu.Display()
	if choice < 0 {
		return nil
	}

	switch choice {
	case 0:
		return h.RunApktoolStatus(apktoolMgr)
	case 1:
		return h.RunFrameworkDetection()
	case 2:
		return h.RunFullAPKAnalysis(apktoolMgr)
	case 3:
		return h.RunManageDecompiled()
	}

	return nil
}

func (h *Handler) RunFullAPKAnalysis(apktoolMgr *tools.Apktool) error {
	ui.Box("APK Analysis")
	ui.NewLine()

	apkPath := ui.GetInput("Enter APK file path")
	if apkPath == "" {
		return fmt.Errorf("no APK path provided")
	}

	analyzer := apk.NewAnalyzer(h.Config.WorkDir)

	// Set apktool path if available
	if apktoolMgr.IsInstalled() {
		analyzer.SetApktoolPath(apktoolMgr.GetPath())
	} else {
		ui.Warning("Apktool not installed - SSL pinning and root detection will be skipped")
		ui.Tip("Install apktool from APK Analyzer → Install Apktool")
		ui.NewLine()
	}

	result, err := analyzer.AnalyzeAPK(apkPath)
	if err != nil {
		return err
	}

	ui.NewLine()
	analyzer.PrintResult(result)
	return nil
}

func (h *Handler) RunFrameworkDetection() error {
	ui.Box("Framework Detection")
	ui.NewLine()

	apkPath := ui.GetInput("Enter APK file path")
	if apkPath == "" {
		return fmt.Errorf("no APK path provided")
	}

	analyzer := apk.NewAnalyzer(h.Config.WorkDir)
	frameworks, err := analyzer.DetectFrameworks(apkPath)
	if err != nil {
		return err
	}

	ui.NewLine()
	ui.Println("📦 Detected Frameworks:")
	for _, fw := range frameworks {
		ui.Success("  • %s", fw)
	}
	return nil
}

func (h *Handler) RunManageDecompiled() error {
	ui.Box("Manage Decompiled Files")
	ui.NewLine()

	analyzer := apk.NewAnalyzer(h.Config.WorkDir)
	dirs, err := analyzer.ListDecompiledDirs()
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		ui.Println("No decompiled APK directories found.")
		return nil
	}

	ui.Println("📂 Decompiled APK directories:")
	for i, dir := range dirs {
		ui.Println("  %d. %s", i+1, dir)
	}
	ui.NewLine()

	menu := ui.NewMenu("Actions")
	menu.AddWithDesc("Delete specific directory", "Remove one decompiled APK")
	menu.AddWithDesc("Delete all directories", "Remove all decompiled APKs")

	choice := menu.Display()
	if choice < 0 {
		return nil
	}

	switch choice {
	case 0:
		// Delete specific
		input := ui.GetInput("Enter directory number to delete")
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(dirs) {
			return fmt.Errorf("invalid selection")
		}
		dirName := dirs[idx-1]
		if err := analyzer.DeleteDecompiledDir(dirName); err != nil {
			return err
		}
		ui.Success("Deleted: %s", dirName)

	case 1:
		// Delete all
		confirm := ui.GetInputDefault("Delete ALL decompiled directories? (y/n)", "n")
		if confirm == "y" || confirm == "Y" {
			if err := analyzer.DeleteAllDecompiledDirs(); err != nil {
				return err
			}
			ui.Success("All decompiled directories deleted")
		} else {
			ui.Println("Cancelled")
		}
	}

	return nil
}

func (h *Handler) RunApktoolStatus(apktoolMgr *tools.Apktool) error {
	ui.Box("Apktool Status")
	ui.NewLine()

	// Check Java first
	if tools.IsJavaInstalled() {
		javaVersion := tools.GetJavaVersion()
		ui.Success("Java: Installed (v%s)", javaVersion)
	} else {
		ui.Warning("Java: Not installed")
		ui.Tip("Apktool requires Java to run")
	}
	ui.NewLine()

	// Check Apktool
	if apktoolMgr.IsInstalled() {
		installedVersion := apktoolMgr.GetInstalledVersion()
		ui.Success("Apktool: Installed (v%s)", installedVersion)
		ui.Println("  Path: %s", apktoolMgr.GetPath())

		// Check for updates
		ui.NewLine()
		ui.Info("Checking for updates...")
		hasUpdate, _, latestVersion, err := apktoolMgr.CheckUpdate()
		if err != nil {
			ui.Warning("Could not check for updates: %v", err)
		} else if hasUpdate {
			ui.Warning("Update available: v%s → v%s", installedVersion, latestVersion)
			confirm := ui.GetInputDefault("Update to latest version? (y/n)", "y")
			if confirm == "y" || confirm == "Y" {
				return apktoolMgr.Update()
			}
		} else {
			ui.Success("Apktool is up to date")
		}
	} else {
		ui.Warning("Apktool: Not installed")
		ui.NewLine()

		// Ask user if they want to install
		if !tools.IsJavaInstalled() {
			ui.Error("Cannot install apktool without Java")
			ui.Tip("Please install Java (JDK or JRE) first")
			return nil
		}

		// Fetch and show latest version
		ui.Info("Fetching latest version...")
		latestVersion, err := apktoolMgr.GetLatestVersion()
		if err != nil {
			ui.Warning("Could not fetch version: %v", err)
			latestVersion = "latest"
		}

		confirm := ui.GetInputDefault(fmt.Sprintf("Download and install apktool v%s? (y/n)", latestVersion), "y")
		if confirm == "y" || confirm == "Y" {
			return apktoolMgr.Download()
		}
	}

	return nil
}
