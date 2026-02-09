package main

import (
	"os"

	"github.com/mahirrudin/godroid/cmd/godroid/app"
	"github.com/mahirrudin/godroid/internal/config"
	"github.com/mahirrudin/godroid/internal/device"
	"github.com/mahirrudin/godroid/internal/emulator"
	"github.com/mahirrudin/godroid/internal/sdk"
	"github.com/mahirrudin/godroid/internal/ui"
)

func main() {
	// Initialize UI
	ui.Init()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		ui.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Ensure directories exist
	if err := cfg.EnsureDirectories(); err != nil {
		ui.Error("Failed to create directories: %v", err)
		os.Exit(1)
	}

	// Initialize components
	sdkDetector := sdk.NewDetectorWithPath(cfg.SDKPath)
	deviceMgr := device.NewManager(cfg.SDKPath)
	emuCtrl := emulator.NewController(cfg.SDKPath)

	// Try to auto-select a device
	_ = deviceMgr.AutoSelectDevice()

	// Create and initialize app handler
	handler := app.NewHandler(cfg, sdkDetector, deviceMgr, emuCtrl)

	// Initialize default scripts if they don't exist
	handler.InitDefaultScripts()

	// Show startup animation
	ui.DisplayStartupAnimation()

	// Run main menu
	handler.RunMain()
}
