package app

import (
	"fmt"
	"os"

	"github.com/mahirrudin/godroid/internal/avd"
	"github.com/mahirrudin/godroid/internal/frida"
	"github.com/mahirrudin/godroid/internal/sdk"
	"github.com/mahirrudin/godroid/internal/tools"
	"github.com/mahirrudin/godroid/internal/ui"
)

func (h *Handler) RunEnvironmentCheck() error {
	ui.Box("Environment Check")
	ui.NewLine()
	ui.Info("Verifying GoDroid Environment...")
	ui.NewLine()

	// Check SDK components
	sdkStatus := h.SDKDetector.DetectAll()

	results := []EnvironmentCheckResult{
		{
			Name:      "Android SDK",
			Installed: sdkStatus.SDKExists,
			Path:      h.Config.SDKPath,
		},
		{
			Name:      "Platform Tools (ADB)",
			Installed: sdkStatus.ADB.Installed,
			Path:      sdkStatus.ADB.Path,
			Version:   sdkStatus.ADB.Version,
			InPath:    sdkStatus.ADB.InPath,
		},
		{
			Name:      "Android Emulator",
			Installed: sdkStatus.Emulator.Installed,
			Path:      sdkStatus.Emulator.Path,
			Version:   sdkStatus.Emulator.Version,
			InPath:    sdkStatus.Emulator.InPath,
		},
		{
			Name:      "SDK Manager",
			Installed: sdkStatus.SDKManager.Installed,
			Path:      sdkStatus.SDKManager.Path,
			InPath:    sdkStatus.SDKManager.InPath,
		},
		{
			Name:      "AVD Manager",
			Installed: sdkStatus.AVDManager.Installed,
			Path:      sdkStatus.AVDManager.Path,
			InPath:    sdkStatus.AVDManager.InPath,
		},
	}

	// Check Frida (check venv first)
	fridaDetector := frida.NewDetectorWithVenv(h.Config.FridaVenvPath)
	fridaStatus := fridaDetector.DetectFridaTools()
	fridaNote := ""
	if !fridaStatus.Installed {
		fridaNote = " (install via 'Environment Setup' → 'Install frida-tools')"
	} else if fridaStatus.InVenv {
		fridaNote = " [venv]"
	}
	results = append(results, EnvironmentCheckResult{
		Name:      "frida-tools" + fridaNote,
		Installed: fridaStatus.Installed,
		Path:      fridaStatus.Path,
		Version:   fridaStatus.Version,
	})

	// Check Python
	pythonStatus := fridaDetector.DetectPython()
	results = append(results, EnvironmentCheckResult{
		Name:      "Python",
		Installed: pythonStatus.Installed,
		Path:      pythonStatus.Path,
		Version:   pythonStatus.Version,
	})

	// Check Java (required for apktool)
	javaInstalled := tools.IsJavaInstalled()
	javaVersion := ""
	if javaInstalled {
		javaVersion = tools.GetJavaVersion()
	}
	results = append(results, EnvironmentCheckResult{
		Name:      "Java (for apktool)",
		Installed: javaInstalled,
		Version:   javaVersion,
	})

	// Display results
	DisplayEnvironmentResults(results)

	return nil
}

func (h *Handler) RunInstallTools() error {
	menu := ui.NewMenu("Environment Setup")
	menu.AddWithDesc("Install Android SDK Command Line Tools", "Download and install sdkmanager, avdmanager")
	menu.AddWithDesc("Install Platform Tools", "Install adb and fastboot")
	menu.AddWithDesc("Install Android Emulator", "Install the Android emulator")
	menu.AddWithDesc("Install System Image", "Download a system image for AVD")
	menu.AddWithDesc("Install frida-tools (venv)", "Create Python venv and install frida-tools")
	menu.AddWithDesc("Install All Pentesting Tools", "Install objection, reflutter, and more")

	choice := menu.Display()
	if choice < 0 {
		return nil
	}

	installer := sdk.NewInstaller(h.Config.SDKPath)

	switch choice {
	case 0:
		// Install command line tools
		zipPath, err := installer.DownloadCommandLineTools()
		if err != nil {
			return err
		}
		defer os.Remove(zipPath)
		return installer.InstallCommandLineTools(zipPath)

	case 1:
		return installer.InstallPlatformTools()

	case 2:
		return installer.InstallEmulator()

	case 3:
		return h.RunInstallSystemImage(installer)

	case 4:
		return h.RunInstallFridaTools()

	case 5:
		return h.RunInstallAllTools()
	}

	return nil
}

func (h *Handler) RunInstallSystemImage(installer *sdk.Installer) error {
	ui.Box("Select System Image")
	ui.NewLine()

	// Show architecture selection first
	arch := SelectArchitecture()

	// Get images for this architecture
	images := avd.GetSystemImagesByArch(arch)

	ui.NewLine()
	ui.Println("Available system images for %s:", arch)
	for i, img := range images {
		ui.Println("  %d. %s", i+1, img.DisplayName)
	}
	ui.NewLine()

	choice := ui.GetInput("Select image number")
	var idx int
	if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(images) {
		return fmt.Errorf("invalid selection")
	}

	selected := images[idx-1]
	return installer.InstallSystemImage(selected.Package)
}

func (h *Handler) RunInstallFridaTools() error {
	ui.Box("Install frida-tools")
	ui.NewLine()

	fridaInstaller := frida.NewInstaller(h.Config.FridaVenvPath)

	// Check if venv exists
	if fridaInstaller.VenvExists() {
		ui.Warning("Virtual environment already exists at %s", h.Config.FridaVenvPath)
		if !ui.Confirm("Reinstall frida-tools?") {
			return nil
		}
	} else {
		// Create venv
		if err := fridaInstaller.CreateVenv(); err != nil {
			return err
		}
	}

	// Install frida-tools
	if err := fridaInstaller.InstallFridaTools(); err != nil {
		return err
	}

	ui.NewLine()
	ui.Success("frida-tools installed successfully!")
	ui.NewLine()
	ui.Println(fridaInstaller.GetActivationInstructions())

	return nil
}

func (h *Handler) RunInstallAllTools() error {
	ui.Box("Install All Pentesting Tools")
	ui.NewLine()

	fridaInstaller := frida.NewInstaller(h.Config.FridaVenvPath)

	// Ensure venv exists
	if !fridaInstaller.VenvExists() {
		if err := fridaInstaller.CreateVenv(); err != nil {
			return err
		}
	}

	// Environment Setup
	toolsList := []struct {
		name    string
		install func() error
	}{
		{"frida-tools", fridaInstaller.InstallFridaTools},
		{"objection", fridaInstaller.InstallObjection},
		{"reflutter", fridaInstaller.InstallReflutter},
	}

	for _, tool := range toolsList {
		ui.NewLine()
		if err := tool.install(); err != nil {
			ui.Error("Failed to install %s: %v", tool.name, err)
		}
	}

	ui.NewLine()
	ui.Success("All tools installed!")
	ui.NewLine()
	ui.Println(fridaInstaller.GetActivationInstructions())

	return nil
}
