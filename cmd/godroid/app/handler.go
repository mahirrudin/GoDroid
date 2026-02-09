package app

import (
	"fmt"

	"github.com/mahirrudin/godroid/internal/config"
	"github.com/mahirrudin/godroid/internal/device"
	"github.com/mahirrudin/godroid/internal/emulator"
	"github.com/mahirrudin/godroid/internal/sdk"
	"github.com/mahirrudin/godroid/internal/ui"
)

// Handler manages menu actions and dependencies.
type Handler struct {
	Config      *config.Config
	SDKDetector *sdk.Detector
	DeviceMgr   *device.Manager
	EmuCtrl     *emulator.Controller
}

// NewHandler creates a new menu handler.
func NewHandler(cfg *config.Config, sdkDetector *sdk.Detector, deviceMgr *device.Manager, emuCtrl *emulator.Controller) *Handler {
	return &Handler{
		Config:      cfg,
		SDKDetector: sdkDetector,
		DeviceMgr:   deviceMgr,
		EmuCtrl:     emuCtrl,
	}
}

// DisplayDeviceStatus shows device count and active device below the banner.
func (h *Handler) DisplayDeviceStatus() {
	devices, _ := h.DeviceMgr.ListDevices()
	ui.ColorPrint(ui.Dim, "📱 Devices: ")
	ui.ColorPrintln(ui.BrightCyan, "%d connected", len(devices))

	if d := h.DeviceMgr.GetCurrentDevice(); d != nil {
		ui.ColorPrint(ui.Dim, "🎯 Active: ")
		ui.ColorPrintln(ui.BrightGreen, "%s", d.DisplayName())
	} else {
		ui.ColorPrint(ui.Dim, "🎯 Active: ")
		ui.ColorPrintln(ui.BrightYellow, "None selected")
	}
	ui.NewLine()
}

// EnsureDevice checks device connection and prompts for selection if multiple devices exist.
// Always prompts when multiple devices are connected to confirm target.
func (h *Handler) EnsureDevice() error {
	devices, err := h.DeviceMgr.ListDevices()
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices connected. Start an emulator or connect a device")
	}

	// Single device - auto-select silently
	if len(devices) == 1 {
		h.DeviceMgr.SelectDevice(devices[0].Serial)
		h.EmuCtrl.SetDevice(devices[0].Serial)
		ui.Info("Using device: %s", devices[0].DisplayName())
		return nil
	}

	// Multiple devices - always prompt to confirm target
	current := h.DeviceMgr.GetCurrentDevice()
	ui.Warning("Multiple devices detected. Select target:")
	for i, d := range devices {
		marker := "  "
		if current != nil && d.Serial == current.Serial {
			marker = "→ "
		}
		ui.Println("%s%d. %s", marker, i+1, d.DisplayName())
	}
	ui.NewLine()

	defaultChoice := ""
	if current != nil {
		for i, d := range devices {
			if d.Serial == current.Serial {
				defaultChoice = fmt.Sprintf("%d", i+1)
				break
			}
		}
	}

	choice := ui.GetInputDefault("Select device number", defaultChoice)
	var idx int
	if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(devices) {
		return fmt.Errorf("invalid selection")
	}

	selected := devices[idx-1]
	if err := h.DeviceMgr.SelectDevice(selected.Serial); err != nil {
		return err
	}
	h.EmuCtrl.SetDevice(selected.Serial)

	ui.Success("Target: %s", selected.DisplayName())
	return nil
}
