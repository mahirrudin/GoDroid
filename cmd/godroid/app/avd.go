package app

import (
	"fmt"
	"time"

	"github.com/mahirrudin/godroid/internal/avd"
	"github.com/mahirrudin/godroid/internal/sdk"
	"github.com/mahirrudin/godroid/internal/ui"
)

func (h *Handler) RunAVDSetup() error {
	menu := ui.NewMenu("AVD Management")
	menu.AddWithDesc("Create New AVD", "Create a new Android Virtual Device")
	menu.AddWithDesc("List AVDs", "Show all available AVDs")
	menu.AddWithDesc("Start Emulator", "Start an AVD")
	menu.AddWithDesc("Stop Emulator", "Stop the running emulator")
	menu.AddWithDesc("Restart Emulator", "Reboot the running emulator")
	menu.AddWithDesc("Delete AVD", "Remove an AVD")

	choice := menu.Display()
	if choice < 0 {
		return nil
	}

	avdMgr := avd.NewManager(h.Config.SDKPath)

	switch choice {
	case 0:
		return h.RunCreateAVD(avdMgr)
	case 1:
		return h.RunListAVDs(avdMgr)
	case 2:
		return h.RunStartEmulator()
	case 3:
		if err := h.EnsureDevice(); err != nil {
			return err
		}
		return h.EmuCtrl.StopEmulator()
	case 4:
		if err := h.EnsureDevice(); err != nil {
			return err
		}
		return h.EmuCtrl.RebootEmulator()
	case 5:
		return h.RunDeleteAVD(avdMgr)
	}

	return nil
}

func (h *Handler) RunCreateAVD(avdMgr *avd.Manager) error {
	ui.Box("Create New AVD")
	ui.NewLine()

	// Get AVD name
	defaultName := h.Config.DefaultAVD
	name := ui.GetInputDefault("AVD Name", defaultName)

	// Select architecture
	arch := SelectArchitecture()

	// Select system image
	images := avd.GetSystemImagesByArch(arch)
	ui.NewLine()
	ui.Println("Available system images:")
	for i, img := range images {
		ui.Println("  %d. %s", i+1, img.DisplayName)
	}
	ui.NewLine()

	choice := ui.GetInputDefault("Select image", "1")
	var idx int
	fmt.Sscanf(choice, "%d", &idx)
	if idx < 1 || idx > len(images) {
		idx = 1
	}
	selectedImage := images[idx-1]

	// Check if system image is installed
	installer := sdk.NewInstaller(h.Config.SDKPath)
	if !installer.IsSystemImageInstalled(selectedImage.Package) {
		ui.Warning("System image not installed: %s", selectedImage.Package)
		ui.NewLine()

		confirm := ui.GetInputDefault("Install this system image now? (y/n)", "y")
		if confirm == "y" || confirm == "Y" {
			ui.NewLine()
			if err := installer.InstallSystemImage(selectedImage.Package); err != nil {
				return fmt.Errorf("failed to install system image: %w", err)
			}
			ui.NewLine()
		} else {
			return fmt.Errorf("system image not installed. Install it first using 'Environment Setup' → 'Install System Image'")
		}
	}

	// Create AVD config
	avdConfig := &avd.AVDConfig{
		Name:        name,
		Device:      h.Config.DefaultDevice,
		SystemImage: selectedImage.Package,
		RAMSize:     2048,
		HeapSize:    512,
		DataSize:    "2G",
	}

	return avdMgr.CreateAVD(avdConfig)
}

func (h *Handler) RunListAVDs(avdMgr *avd.Manager) error {
	ui.Box("Available AVDs")
	ui.NewLine()

	avds, err := avdMgr.ListAVDs()
	if err != nil {
		return err
	}

	if len(avds) == 0 {
		ui.Warning("No AVDs found. Create one using 'Create New AVD'")
		return nil
	}

	for i, name := range avds {
		ui.Println("  %d. %s", i+1, name)
	}

	return nil
}

func (h *Handler) RunStartEmulator() error {
	ui.Box("Start Emulator")
	ui.NewLine()

	// List available AVDs
	avds, err := h.EmuCtrl.ListAVDs()
	if err != nil {
		return err
	}

	if len(avds) == 0 {
		return fmt.Errorf("no AVDs found. Create one first")
	}

	ui.Println("Available AVDs:")
	for i, name := range avds {
		ui.Println("  %d. %s", i+1, name)
	}
	ui.NewLine()

	choice := ui.GetInputDefault("Select AVD", "1")
	var idx int
	fmt.Sscanf(choice, "%d", &idx)
	if idx < 1 || idx > len(avds) {
		return fmt.Errorf("invalid selection")
	}

	selectedAVD := avds[idx-1]

	coldBoot := ui.Confirm("Cold boot (fresh start)?")

	if err := h.EmuCtrl.StartEmulator(selectedAVD, coldBoot); err != nil {
		return err
	}

	// Wait for emulator
	return h.EmuCtrl.WaitForEmulator(5 * time.Minute)
}

func (h *Handler) RunDeleteAVD(avdMgr *avd.Manager) error {
	avds, err := avdMgr.ListAVDs()
	if err != nil {
		return err
	}

	if len(avds) == 0 {
		ui.Warning("No AVDs to delete")
		return nil
	}

	ui.Println("Available AVDs:")
	for i, name := range avds {
		ui.Println("  %d. %s", i+1, name)
	}
	ui.NewLine()

	choice := ui.GetInput("Select AVD to delete")
	var idx int
	fmt.Sscanf(choice, "%d", &idx)
	if idx < 1 || idx > len(avds) {
		return fmt.Errorf("invalid selection")
	}

	if !ui.Confirm(fmt.Sprintf("Delete AVD '%s'?", avds[idx-1])) {
		return nil
	}

	return avdMgr.DeleteAVD(avds[idx-1])
}
