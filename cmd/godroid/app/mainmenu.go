package app

import (
	"github.com/mahirrudin/godroid/internal/ui"
)

// RunMain displays the main menu and handles navigation.
func (h *Handler) RunMain() {
	for {
		ui.ClearScreen()
		ui.DisplayBanner()
		h.DisplayDeviceStatus()

		menu := ui.NewMenu("Main Menu")
		menu.AddWithDesc("Environment Check", "Verify all prerequisites are installed")
		menu.AddWithDesc("Environment Setup", "Install SDK, frida-tools, and dependencies")
		menu.AddWithDesc("Device Management", "Select and manage connected devices or emulator")
		menu.AddWithDesc("AVD Management", "Create and manage Android Virtual Devices")
		menu.AddWithDesc("AVD Root Emulator", "Root emulator using Magisk and rootAVD")
		menu.AddWithDesc("Rooted Device Configure", "Install Frida server and Burp certificate")
		menu.AddWithDesc("Frida Management", "Run Frida scripts and manage apps")
		menu.AddWithDesc("APK Management", "Analyze APK for frameworks and security features")

		choice := menu.DisplayInline()
		if choice < 0 {
			ui.Println("Goodbye! 👋")
			return
		}

		ui.ClearScreen()
		ui.DisplayCompactBanner()

		var err error
		switch choice {
		case 0:
			err = h.RunEnvironmentCheck()
		case 1:
			err = h.RunInstallTools()
		case 2:
			err = h.RunDeviceManagement()
		case 3:
			err = h.RunAVDSetup()
		case 4:
			err = h.RunRootEmulator()
		case 5:
			err = h.RunConfigureEmulator()
		case 6:
			err = h.RunFridaTools()
		case 7:
			err = h.RunAPKAnalyzer()
		}

		if err != nil {
			ui.Error("Error: %v", err)
		}
		ui.WaitForEnter()
	}
}
