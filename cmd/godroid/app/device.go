package app

import (
	"fmt"

	"github.com/mahirrudin/godroid/internal/device"
	"github.com/mahirrudin/godroid/internal/ui"
)

// RunDeviceManagement handles device selection and remote connections.
func (h *Handler) RunDeviceManagement() error {
	for {
		ui.ClearScreen()
		ui.DisplayCompactBanner()

		// Show current device
		if d := h.DeviceMgr.GetCurrentDevice(); d != nil {
			ui.ColorPrintln(ui.BrightGreen, "📱 Current: %s", d.DisplayName())
		} else {
			ui.ColorPrintln(ui.BrightYellow, "📱 No device selected")
		}
		ui.NewLine()

		menu := ui.NewMenu("Device Management")
		menu.AddWithDesc("List Connected Devices", "Show all available devices (emulator/USB/remote)")
		menu.AddWithDesc("Select Active Device", "Choose which device to work with")
		menu.AddWithDesc("Connect Remote Device", "Connect to a device via wireless ADB")
		menu.AddWithDesc("Disconnect Remote Device", "Disconnect from a remote device")
		menu.AddWithDesc("Refresh Devices", "Refresh the device list")

		choice := menu.Display()
		if choice < 0 {
			return nil
		}

		ui.NewLine()

		var err error
		switch choice {
		case 0:
			err = h.RunListDevices()
		case 1:
			err = h.RunSelectDevice()
		case 2:
			err = h.RunConnectRemote()
		case 3:
			err = h.RunDisconnectRemote()
		case 4:
			_, err = h.DeviceMgr.RefreshDevices()
			if err == nil {
				ui.Success("Device list refreshed")
			}
		}

		if err != nil {
			ui.Error("Error: %v", err)
		}
		ui.WaitForEnter()
	}
}

func (h *Handler) RunListDevices() error {
	ui.Box("Connected Devices")
	ui.NewLine()

	devices, err := h.DeviceMgr.ListDevices()
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		ui.Warning("No devices connected")
		ui.NewLine()
		ui.Println("To connect a device:")
		ui.Println("  • Start an emulator using 'AVD Setup' → 'Start Emulator'")
		ui.Println("  • Connect a USB device with USB debugging enabled")
		ui.Println("  • Connect wirelessly using 'Connect Remote Device'")
		return nil
	}

	current := h.DeviceMgr.GetCurrentSerial()
	for i, d := range devices {
		marker := "  "
		if d.Serial == current {
			marker = "→ "
		}

		status := ui.BrightGreen
		statusText := "online"
		if !d.IsOnline() {
			status = ui.BrightRed
			statusText = d.State
		}

		ui.ColorPrint(status, "%s%d. ", marker, i+1)
		ui.Println("%s", d.DisplayName())
		ui.ColorPrintln(ui.Dim, "      Status: %s", statusText)
	}

	return nil
}

func (h *Handler) RunSelectDevice() error {
	devices, err := h.DeviceMgr.ListDevices()
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices connected")
	}

	ui.Println("Available devices:")
	for i, d := range devices {
		ui.Println("  %d. %s", i+1, d.DisplayName())
	}
	ui.NewLine()

	choice := ui.GetInput("Select device number")
	var idx int
	if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(devices) {
		return fmt.Errorf("invalid selection")
	}

	selected := devices[idx-1]
	if err := h.DeviceMgr.SelectDevice(selected.Serial); err != nil {
		return err
	}

	// Update the controller to use this device
	h.EmuCtrl.SetDevice(selected.Serial)

	ui.Success("Selected device: %s", selected.DisplayName())
	return nil
}

func (h *Handler) RunConnectRemote() error {
	ui.Box("Connect Remote Device")
	ui.NewLine()

	ui.Println("Enter the IP address of the device.")
	ui.Println("Port defaults to 5555 if not specified.")
	ui.NewLine()

	address := ui.GetInput("Address (e.g., 192.168.1.100 or 192.168.1.100:5555)")
	if address == "" {
		return fmt.Errorf("address required")
	}

	ui.Info("Connecting to %s...", address)

	if err := h.DeviceMgr.ConnectRemote(address); err != nil {
		return err
	}

	ui.Success("Connected to %s", address)

	// Refresh and auto-select if this is the only device
	devices, _ := h.DeviceMgr.RefreshDevices()
	if len(devices) == 1 {
		h.DeviceMgr.SelectDevice(devices[0].Serial)
		h.EmuCtrl.SetDevice(devices[0].Serial)
		ui.Info("Auto-selected as active device")
	}

	return nil
}

func (h *Handler) RunDisconnectRemote() error {
	devices, err := h.DeviceMgr.ListDevices()
	if err != nil {
		return err
	}

	// Filter remote devices
	var remoteDevices []device.Device
	for _, d := range devices {
		if d.Type == device.DeviceTypeRemote {
			remoteDevices = append(remoteDevices, d)
		}
	}

	if len(remoteDevices) == 0 {
		ui.Warning("No remote devices connected")
		return nil
	}

	ui.Println("Remote devices:")
	for i, d := range remoteDevices {
		ui.Println("  %d. %s", i+1, d.Serial)
	}
	ui.NewLine()

	choice := ui.GetInput("Select device to disconnect (or 'all')")

	if choice == "all" {
		if err := h.DeviceMgr.DisconnectAll(); err != nil {
			return err
		}
		ui.Success("All remote devices disconnected")
		return nil
	}

	var idx int
	if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(remoteDevices) {
		return fmt.Errorf("invalid selection")
	}

	selected := remoteDevices[idx-1]
	if err := h.DeviceMgr.DisconnectRemote(selected.Serial); err != nil {
		return err
	}

	ui.Success("Disconnected from %s", selected.Serial)
	return nil
}
