package app

import (
	"github.com/mahirrudin/godroid/internal/ui"
)

// SelectArchitecture prompts user to select architecture.
func SelectArchitecture() string {
	items := []string{
		"x86_64 (Intel/AMD - Recommended for most systems)",
		"arm64-v8a (ARM - For Apple Silicon Macs or ARM systems)",
	}

	choice := ui.SelectFromList("Select Architecture", items)
	if choice == 0 {
		return "x86_64"
	}
	return "arm64-v8a"
}

// EnvironmentCheckResult represents the result of an environment check.
type EnvironmentCheckResult struct {
	Name      string
	Installed bool
	Path      string
	Version   string
	InPath    bool
}

// DisplayEnvironmentResults shows environment check results.
func DisplayEnvironmentResults(results []EnvironmentCheckResult) {
	ui.Box("Environment Check Results")
	ui.NewLine()

	allPassed := true
	for _, r := range results {
		var status, color string
		if r.Installed {
			status = string(ui.IconCheck)
			color = ui.BrightGreen
		} else {
			status = string(ui.IconCross)
			color = ui.BrightRed
			allPassed = false
		}

		ui.ColorPrint(color, "[%s] ", status)
		ui.Print("%s", r.Name)

		if r.Version != "" {
			ui.ColorPrint(ui.Dim, " (%s)", r.Version)
		}

		// Show path for installed items
		if r.Installed && r.Path != "" {
			ui.ColorPrint(ui.Dim, " - %s", r.Path)
		}

		ui.NewLine()
	}

	ui.NewLine()
	if allPassed {
		ui.Success("All prerequisites are installed!")
	} else {
		ui.Warning("Some prerequisites are missing. Use 'Environment Setup' to fix.")
	}
}
